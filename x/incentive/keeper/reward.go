package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
)

// ClearRewardTracker clears all reward trackers matching a specific account + tier + bonded uToken denom
// from the store by setting them to zero
func (k Keeper) ClearRewardTracker(ctx sdk.Context, addr sdk.AccAddress, tier incentive.BondTier, bondDenom string,
) error {
	trackers := k.getFullRewardTracker(ctx, addr, bondDenom, tier)
	for _, rewardCoin := range trackers {
		zeroCoin := sdk.NewDecCoinFromDec(rewardCoin.Denom, sdk.ZeroDec())
		if err := k.setRewardTracker(ctx, addr, bondDenom, zeroCoin, tier); err != nil {
			return err
		}
	}
	return nil
}

// UpdateRewardTracker updates all reward trackers matching a specific account + tier + bonded uToken denom
// by setting them to the current values of that tier + uToken denom's reward accumulators
func (k Keeper) UpdateRewardTracker(ctx sdk.Context, addr sdk.AccAddress, tier incentive.BondTier, bondDenom string,
) error {
	trackers := k.getFullRewardTracker(ctx, addr, bondDenom, tier)
	accumulators := k.getFullRewardAccumulator(ctx, bondDenom, tier)
	for _, rewardCoin := range trackers {
		accumulator := sdk.NewDecCoinFromDec(rewardCoin.Denom, accumulators.AmountOf(rewardCoin.Denom))
		if err := k.setRewardTracker(ctx, addr, bondDenom, accumulator, tier); err != nil {
			return err
		}
	}
	return nil
}

// ClaimReward claims a single account's bonded uToken tier's reward, then updates its reward tracker.
// Returns rewards claimed.
func (k Keeper) ClaimReward(_ sdk.Context, _ sdk.AccAddress, _ incentive.BondTier, _ sdk.Coin,
) (sdk.Coins, error) {
	// TODO - implement claim logic (especially needs high exponent asset compatibility)
	rewards := sdk.NewCoins()
	return rewards, incentive.ErrNotImplemented
	// k.updateRewardTracker(ctx, addr, tier, bonded.Denom)
}
