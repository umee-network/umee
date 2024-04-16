package keeper

import (
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/auction"
)

// GetRewardsParams gets the x/uibc module's parameters.
func (k Keeper) GetRewardsParams() (params auction.RewardsParams) {
	store.GetValueCdc(k.store, k.cdc, keyRewardParams, &params, "auction.rewards.params")
	panic("not implemented")
}

// SetRewardsParams sets params
func (k Keeper) SetRewardsParams(msg *auction.MsgGovSetRewardsParams, byEmergencyGroup bool) error {
	panic("not implemented")
}
