package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
)

var ten = sdk.MustNewDecFromStr("10")

// UpdateAccount finishes any unbondings associated with an account which have ended and claims any pending rewards.
// It returns the amount of rewards claimed.
//
// Unlike updateRewards and updatePrograms, this function is not called during EndBlock.
//
// REQUIREMENT: This function must be called during any message or hook which creates an unbonding or updates
// bonded amounts. Leverage hooks which decrease borrower collateral must also call this before acting.
// This ensures that between any two consecutive claims by a single account, bonded amounts were constant
// on that account for each collateral uToken denom.
func (k Keeper) UpdateAccount(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	// unbondings have already been subtracted from bonded amounts when they are started,
	// so it is fine to finish completed unbondings before claiming rewards.
	if err := k.cleanupUnbondings(ctx, addr); err != nil {
		return sdk.NewCoins(), err
	}

	return k.claimRewards(ctx, addr)
}

// updateRewards updates any rewardAccumulators associated with ongoing incentive programs
// based on the time elapsed between LastRewardTime and block time. It decreases active
// incentive programs' RemainingRewards by the amount of coins distributed.
// Also sets module's LastRewardsTime to block time.
func (k Keeper) updateRewards(ctx sdk.Context, blockTime int64) error {
	prevTime := k.GetLastRewardsTime(ctx)
	if prevTime > blockTime {
		return incentive.ErrDecreaseLastRewardTime
	}
	if prevTime == blockTime {
		return nil
	}
	timeElapsed := blockTime - prevTime
	ongoingPrograms, err := k.getAllIncentivePrograms(ctx, incentive.ProgramStatusOngoing)
	if err != nil {
		return err
	}

	for _, p := range ongoingPrograms {
		bondedDenom := p.UToken
		bonded := k.getTotalBonded(ctx, bondedDenom)
		if bonded.IsZero() {
			// If no uTokens are bonded in the incentivized denom, nothing happens with rewards
			continue
		}

		// calculate the amount of time (in seconds) that remained on the incentive
		// program after the previous calculation.
		prevRemainingTime := (p.StartTime + p.Duration) - prevTime
		// if remaining time was zero or negative, but program had not been removed
		// from ongoing programs, using a value of 1 second ensures all its remaining
		// rewards are distributed
		if prevRemainingTime < 1 {
			prevRemainingTime = 1
		}

		// The portion of a program's remaining rewards,
		// ranging from 0 to 1, which will be distributed this
		// block. The max value of 1 means 100% of remaining rewards
		// will be used, as occurs when a program is ending.
		programRewardsFraction := sdk.MinDec(
			sdk.OneDec(),
			sdk.NewDecFromInt(sdk.NewInt(timeElapsed)).
				Quo(sdk.NewDec(prevRemainingTime)))

		// each incentive program has only one reward denom
		rewardDenom := p.RemainingRewards.Denom

		// get this block's rewards (as a token amount) for this incentive program only
		thisBlockRewards := sdk.NewCoin(
			rewardDenom,
			sdk.NewDecFromInt(p.RemainingRewards.Amount).Mul(programRewardsFraction).TruncateInt())

		// get expected increase of bondDenom's rewardAccumulator of reward denom,
		// by dividing this block's rewards proportionally among bonded uTokens,
		// and adjusting for the reward accumulator's exponent
		accumulator := k.getRewardAccumulator(ctx, bondedDenom)
		accumulatorIncrease := sdk.NewDecFromInt(thisBlockRewards.Amount).
			Mul(ten.Power(uint64(accumulator.Exponent))).
			Quo(sdk.NewDecFromInt(bonded.Amount))

		// if accumulator increase is so small it rounds to zero even after power adjustment,
		// no rewards were distributed
		if accumulatorIncrease.IsZero() {
			continue
		}

		// if nonzero, increase reward accumulator using rewards from this incentive program
		// and subtract them from that program's remaining rewards
		accumulator.Rewards = accumulator.Rewards.Add(sdk.NewDecCoinFromDec(rewardDenom, accumulatorIncrease))
		p.RemainingRewards = p.RemainingRewards.Sub(thisBlockRewards)

		// update program state and reward accumulator
		if err := k.setIncentiveProgram(ctx, p, incentive.ProgramStatusOngoing); err != nil {
			return err
		}
		if err := k.setRewardAccumulator(ctx, accumulator); err != nil {
			return err
		}
	}
	// After updates, module's LastRewardTime increases to current block time
	return k.setLastRewardsTime(ctx, blockTime)
}

// updatePrograms moves any incentive programs which have reached their end time from Ongoing to Completed,
// and moves any funded programs which should start from Upcoming to Ongoing. Non-funded programs which
// would start are moved to completed status instead. Uses the current LastRewardsTime but does not update it.
func (k Keeper) updatePrograms(ctx sdk.Context) error {
	blockTime := k.GetLastRewardsTime(ctx)
	ongoingPrograms, err := k.getAllIncentivePrograms(ctx, incentive.ProgramStatusOngoing)
	if err != nil {
		return err
	}
	for _, op := range ongoingPrograms {
		// if an ongoing program is ending, move it to completed programs without modifying any fields
		if blockTime >= op.Duration+op.StartTime {
			if err := k.moveIncentiveProgram(ctx, op.ID, incentive.ProgramStatusCompleted); err != nil {
				return err
			}
		}
	}
	upcomingPrograms, err := k.getAllIncentivePrograms(ctx, incentive.ProgramStatusUpcoming)
	if err != nil {
		return err
	}
	for _, up := range upcomingPrograms {
		// if an upcoming program has reached its start time
		if blockTime >= up.StartTime {
			// prepare to start the program
			newStatus := incentive.ProgramStatusOngoing
			// or immediately cancel it if it was not funded
			if !up.Funded {
				newStatus = incentive.ProgramStatusCompleted
			}
			if err := k.moveIncentiveProgram(ctx, up.ID, newStatus); err != nil {
				return err
			}
		}
	}

	// Note that even if a program had a duration shorter than the time between blocks, this function's
	// order of ending eligible ongoing programs before starting eligible upcoming ones ensures that each
	// program will be active for updateRewards for at least one full block. (The same program will not be
	// both started and ended in the same block.)
	return nil
}

// EndBlock updates incentive programs and reward accumulators, then sets LastRewardTime
// to the current block time. Also protects against negative time elapsed (without causing chain halt).
// In addition to regular error, returns a boolean indicating whether the main logic was skipped due
// to a blockTime issue. These situations are accompanied by error logs.
func (k Keeper) EndBlock(ctx sdk.Context) (skipped bool, err error) {
	blockTime := ctx.BlockTime().Unix()
	if blockTime < 0 {
		k.Logger(ctx).Error(
			incentive.ErrDecreaseLastRewardTime.Error(),
			"negative block time", blockTime,
		)
		return true, nil
	}

	prevTime := k.GetLastRewardsTime(ctx)
	if prevTime <= 0 {
		// if stored LastRewardTime is zero (or negative), either the chain has just started or the genesis file has been
		// modified intentionally. In either case, proceed as if 0 seconds have passed since the last block,
		// thus accruing no rewards and setting the current BlockTime as the new starting point.
		k.Logger(ctx).Error(
			"incentive module LastRewardTime was not initialized",
			"prev", prevTime,
			"blockTime", blockTime,
		)
		if err := k.setLastRewardsTime(ctx, blockTime); err != nil {
			return true, err
		}
		prevTime = blockTime
	}

	if blockTime <= prevTime {
		// Avoids this and related issues: https://github.com/tendermint/tendermint/issues/8773
		k.Logger(ctx).Error(
			"incentive module will wait for block time > prevRewardTime",
			"current", blockTime,
			"prev", prevTime,
		)

		// if LastRewardTime appears to be in the future, do nothing (besides logging) and leave
		// LastRewardTime at its stored value. This will repeat every block until BlockTime exceeds
		// LastRewardTime.
		return true, nil
	}

	// Implications of updating reward accumulators and PrevRewardTime before starting/completing incentive programs:
	// - if an incentive program starts this block, it does not disburse any rewards this block
	// - if an incentive program completes this block, it distributes its remaining rewards before being completed
	// - in the case of a chain halt which misses a program's start time, rewards before its late start are skipped
	// - in the case of a chain halt which misses a program's end time, remaining rewards are correctly distributed
	if err := k.updateRewards(ctx, blockTime); err != nil {
		return false, err
	}
	return false, k.updatePrograms(ctx)
}
