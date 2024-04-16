package keeper

import (
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/auction"
)

func (k Keeper) rewardsBid(msg *auction.MsgRewardsBid) error {
	keyMsg := "auction.rewards.highest_bid"
	lastBid := store.GetValue[*auction.Bid](k.store, keyRewardsHighestBid, keyMsg)
	minBid := auction.MinRewardsBid
	if lastBid != nil {
		minBid = lastBid.Amount.Add(minBid)
	}
	if err := auction.ValidateMinRewarsdsBid(minBid, msg.Amount); err != nil {
		return err
	}
	bid := auction.Bid{Bidder: msg.Sender, Amount: msg.Amount.Amount}
	return store.SetValue(k.store, keyRewardsHighestBid, &bid, keyMsg)
}

func (k Keeper) currentRewardsAuction() (*auction.QueryRewardsAuctionResponse, error) {
	panic("not implemented")
}
