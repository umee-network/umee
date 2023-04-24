package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/incentive"
)

// updateRewardTracker updates the reward tracker matching a specific account + bonded uToken denom
// by setting it to the current value of that uToken denom's reward accumulator. Used after claiming
// rewards or when setting bonded amount from zero to a nonzero amount (i.e. initializing reward tracker).
func (k Keeper) updateRewardTracker(ctx sdk.Context, addr sdk.AccAddress, bondDenom string,
) error {
	tracker := k.getRewardTracker(ctx, addr, bondDenom)
	accumulator := k.getRewardAccumulator(ctx, bondDenom)

	tracker.Rewards = accumulator.Rewards

	// reward tracker contains address and bond denom, plus updated reward coins
	return k.setRewardTracker(ctx, tracker)
}

// claimRewards claims a single account's uToken's rewards for all bonded uToken denoms. Returns rewards claimed.
func (k Keeper) claimRewards(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	rewards := sdk.NewCoins()
	bondedDenoms, err := k.getAllBondDenoms(ctx, addr)
	if err != nil {
		return sdk.NewCoins(), err
	}
	for _, bondDenom := range bondedDenoms {
		tokens := k.calculateSingleReward(ctx, addr, bondDenom)

		// If all rewards were too small to disburse for this specific bonded denom,
		// skips updating its reward tracker to prevent wasting of fractional rewards.
		// If nonzero, proceed to claim.
		if !tokens.IsZero() {
			// send claimed rewards from incentive module to user account
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, incentive.ModuleName, addr, tokens); err != nil {
				return sdk.NewCoins(), err
			}
			// update the user's reward tracker to indicate that they last claimed rewards at the current
			// value of rewardAccumulator
			if err := k.updateRewardTracker(ctx, addr, bondDenom); err != nil {
				return sdk.NewCoins(), err
			}

			// adds rewards claimed from this single bonded denom to the total
			rewards = rewards.Add(tokens...)
		}
	}
	return rewards, nil
}

// calculateRewards calculates a single account's uToken's pending rewards for all bonded uToken denoms,
// without claiming them or updating its reward trackers. Returns rewards pending.
func (k Keeper) calculateRewards(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	rewards := sdk.NewCoins()
	bondedDenoms, err := k.getAllBondDenoms(ctx, addr)
	if err != nil {
		return sdk.NewCoins(), err
	}
	for _, bondDenom := range bondedDenoms {
		tokens := k.calculateSingleReward(ctx, addr, bondDenom)
		if !tokens.IsZero() {
			// adds rewards pending for this single bonded denom to the total
			rewards = rewards.Add(tokens...)
		}
	}
	return rewards, nil
}

// calculateSingleReward calculates a single account's uToken's rewards for a single bonded uToken denom,
// without claiming them or updating its reward tracker. Returns rewards pending.
func (k Keeper) calculateSingleReward(ctx sdk.Context, addr sdk.AccAddress, bondDenom string) sdk.Coins {
	rewards := sdk.NewCoins()

	accumulator := k.getRewardAccumulator(ctx, bondDenom)
	tracker := k.getRewardTracker(ctx, addr, bondDenom)

	// Rewards are based on the amount accumulator has increased since tracker was last updated
	delta := accumulator.Rewards.Sub(tracker.Rewards)
	if delta.IsZero() {
		return sdk.NewCoins()
	}

	// Actual token amounts must be reduced according to accumulator's exponent
	for _, coin := range delta {
		bonded := k.GetBonded(ctx, addr, bondDenom)

		// reward = bonded * delta / 10^exponent
		rewardDec := sdk.NewDecFromInt(bonded.Amount).Quo(
			ten.Power(uint64(accumulator.Exponent)),
		).Mul(coin.Amount)

		// rewards round down
		reward := sdk.NewCoin(coin.Denom, rewardDec.TruncateInt())
		rewards = rewards.Add(reward)
	}

	return rewards
}
