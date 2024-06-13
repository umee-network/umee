package auction

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis creates a default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		RewardsParams:   RewardsParams{BidDuration: 14 * 24 * 3600}, // 14 days
		RewardAuctionId: 0,
		RewardsAuctions: []RewardsKV{},
		RewardsBids:     []BidKV{},
	}
}

func (gs *GenesisState) Validate() error {
	if gs.RewardsParams.BidDuration <= 60 {
		return errors.New("RewardsParams.BidDuration must be at least 60s")
	}
	if gs.RewardsParams.BidDuration >= 180*24*3600 {
		return errors.New("RewardsParams.BidDuration must be at most 15552000s = 180days")
	}
	for _, elem := range gs.RewardsAuctions {
		coins := sdk.Coins(elem.Rewards.Rewards)
		if err := coins.Validate(); err != nil {
			return err
		}
		if elem.Id > gs.RewardAuctionId {
			return fmt.Errorf("rewards_auctions ID must be at most rewards_auction_id")
		}
	}
	for _, elem := range gs.RewardsBids {
		if elem.Id > gs.RewardAuctionId {
			return fmt.Errorf("rewards_bids ID must be at most rewards_auction_id")
		}
	}

	return nil
}
