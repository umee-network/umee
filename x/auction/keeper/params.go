package keeper

import (
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/auction"
)

// GetRewardsParams gets the x/uibc module's parameters.
func (k Keeper) GetRewardsParams() (params auction.RewardsParams) {
	store.GetValueCdc(k.store, k.cdc, keyRewardsParams, &params, "auction.rewards.params")
	return params
}

// SetRewardsParams sets params
func (k Keeper) SetRewardsParams(p *auction.RewardsParams, byEmergencyGroup bool) error {
	return store.SetValueCdc(k.store, k.cdc, keyRewardsParams, p, "auction.rewards.params")
}
