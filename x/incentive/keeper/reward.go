package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
)

// clearRewardTracker clears all reward trackers matching a specific account + tier + bonded uToken denom
// from the store by setting them to zero
func (k Keeper) clearRewardTracker(ctx sdk.Context, addr sdk.AccAddress, tier incentive.BondTier, bondDenom string) error {
	trackers := k.getFullRewardTracker(ctx, addr, bondDenom, tier)
	for _, rewardCoin := range trackers {
		zeroCoin := sdk.NewDecCoinFromDec(rewardCoin.Denom, sdk.ZeroDec())
		if err := k.SetRewardTracker(ctx, addr, bondDenom, zeroCoin, tier); err != nil {
			return err
		}
	}
	return nil
}

// updateRewardTracker updates all reward trackers matching a specific account + tier + bonded uToken denom
// by setting them to the current values of that tier + uToken denom's reward accumulators
func (k Keeper) updateRewardTracker(ctx sdk.Context, addr sdk.AccAddress, tier incentive.BondTier, bondDenom string) error {
	trackers := k.getFullRewardTracker(ctx, addr, bondDenom, tier)
	accumulators := k.getFullRewardAccumulator(ctx, bondDenom, tier)
	for _, rewardCoin := range trackers {
		accumulator := sdk.NewDecCoinFromDec(rewardCoin.Denom, accumulators.AmountOf(rewardCoin.Denom))
		if err := k.SetRewardTracker(ctx, addr, bondDenom, accumulator, tier); err != nil {
			return err
		}
	}
	return nil
}

// claimReward claims a single account's bonded uToken tier's reward, then updates its reward tracker.
// Returns rewards claimed.
func (k Keeper) claimReward(ctx sdk.Context, addr sdk.AccAddress, tier incentive.BondTier, bonded sdk.Coin,
) (sdk.Coins, error) {
	// TODO - implement claim logic (especially needs high exponent asset compatibility)
	rewards := sdk.NewCoins()
	return rewards, incentive.ErrNotImplemented
	// k.updateRewardTracker(ctx, addr, tier, bonded.Denom)
}
