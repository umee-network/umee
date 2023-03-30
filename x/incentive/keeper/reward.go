package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
)

// UpdateRewards increases the module's LastInterestTime and any rewardAccumulators associated with
// ongoing incentive programs.
func (k Keeper) UpdateRewards(ctx sdk.Context) error {
	currentTime := uint64(ctx.BlockTime().Unix())
	prevRewardTime := k.getLastRewardsTime(ctx)
	if prevRewardTime <= 0 {
		// if stored LastRewardTime is zero (or negative), either the chain has just started
		// or the genesis file has been modified intentionally. In either case, proceed as if
		// 0 seconds have passed since the last block, thus accruing no rewards and setting
		// the current BlockTime as the new starting point.
		prevRewardTime = currentTime
	}

	if currentTime < prevRewardTime {
		// TODO fix this when tendermint solves https://github.com/tendermint/tendermint/issues/8773
		k.Logger(ctx).With("EndBlocker will wait for block time > prevRewardTime").Error(
			incentive.ErrDecreaseLastRewardTime.Error(),
			"current", currentTime,
			"prev", prevRewardTime,
		)

		// if LastRewardTime appears to be in the future, do nothing (besides logging) and leave
		// LastRewardTime at its stored value. This will repeat every block until BlockTime exceeds
		// LastRewardTime.
		return nil
	}

	// TODO: reward accumulator math

	// set LastRewardTime to current block time
	return k.setLastRewardsTime(ctx, currentTime)
}

// clearRewardTracker clears all reward trackers matching a specific account + bonded uToken denom
// from the store by setting them to zero
func (k Keeper) clearRewardTracker(ctx sdk.Context, addr sdk.AccAddress, bondDenom string,
) error {
	trackers := k.getFullRewardTracker(ctx, addr, bondDenom)
	for _, rewardCoin := range trackers {
		zeroCoin := sdk.NewDecCoinFromDec(rewardCoin.Denom, sdk.ZeroDec())
		if err := k.setRewardTracker(ctx, addr, bondDenom, zeroCoin); err != nil {
			return err
		}
	}
	return nil
}

// UpdateRewardTracker updates all reward trackers matching a specific account + bonded uToken denom
// by setting them to the current values of that uToken denom's reward accumulators
func (k Keeper) UpdateRewardTracker(ctx sdk.Context, addr sdk.AccAddress, bondDenom string,
) error {
	trackers := k.getFullRewardTracker(ctx, addr, bondDenom)
	accumulators := k.getFullRewardAccumulator(ctx, bondDenom)
	for _, rewardCoin := range trackers {
		accumulator := sdk.NewDecCoinFromDec(rewardCoin.Denom, accumulators.AmountOf(rewardCoin.Denom))
		if err := k.setRewardTracker(ctx, addr, bondDenom, accumulator); err != nil {
			return err
		}
	}
	return nil
}

// ClaimReward claims a single account's bonded uToken's reward, then updates its reward tracker.
// Returns rewards claimed.
func (k Keeper) ClaimReward(_ sdk.Context, _ sdk.AccAddress, _ sdk.Coin,
) (sdk.Coins, error) {
	// TODO - implement claim logic (especially needs high exponent asset compatibility)
	rewards := sdk.NewCoins()
	return rewards, incentive.ErrNotImplemented
	// k.updateRewardTracker(ctx, addr, bonded.Denom)
}
