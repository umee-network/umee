package keeper

import (
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/auction"
)

func (k Keeper) ExportGenesis() (*auction.GenesisState, error) {
	rewards, err := k.getAllRewardsAuctions()
	if err != nil {
		return nil, err
	}
	bids, err := k.getAllRewardsBids()
	if err != nil {
		return nil, err
	}
	return &auction.GenesisState{
		RewardsParams:   k.GetRewardsParams(),
		RewardAuctionId: k.currentRewardsAuctionID(),
		RewardsAuctions: rewards,
		RewardsBids:     bids,
	}, nil
}

func (k Keeper) InitGenesis(g *auction.GenesisState) error {
	if err := k.SetRewardsParams(&g.RewardsParams, false); err != nil {
		return err
	}
	if err := k.storeAllRewardsAuctions(g.RewardsAuctions); err != nil {
		return err
	}
	if err := k.storeAllRewardsBids(g.RewardsBids); err != nil {
		return err
	}
	store.SetInteger(k.store, keyRewardsCurrentID, g.RewardAuctionId)

	return nil
}
