package keeper

import (
	// "github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/auction"
)

func (k Keeper) rewardsBid(msg *auction.MsgRewardsBid) error {

	// store.GetValue(k.store, k.cdc, keyRewardsParams, &params, "auction.rewards.params")

	panic("not implemented")
}

func (k Keeper) currentRewardsAuction() (*auction.QueryRewardsAuctionResponse, error) {
	panic("not implemented")
}
