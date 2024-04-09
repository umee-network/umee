package keeper

import (
	"github.com/umee-network/umee/v6/x/auction"
)

// GetRewardsParams gets the x/uibc module's parameters.
func (k Keeper) GetRewardsParams() (params auction.RewardsParams) {
	panic("not implemented")
}

// SetRewardsParams sets params
func (k Keeper) SetRewardsParams(msg *auction.MsgGovSetRewardsParams, byEmergencyGroup bool) error {
	panic("not implemented")
}
