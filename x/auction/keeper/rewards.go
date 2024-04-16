package keeper

import (
	"errors"
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/auction"
)

func umeeCoins(amount sdkmath.Int) sdk.Coins {
	return sdk.Coins{coin.UmeeInt(amount)}
}

func (k Keeper) currentRewardsAuction() uint32 {
	id, _ := store.GetInteger[uint32](k.store, keyRwardsCurrentID)
	return id
}

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

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	util.Panic(err)
	prevBidder, err := sdk.AccAddressFromBech32(lastBid.Bidder)
	util.Panic(err)
	vault := k.accs.RewardsBid
	id := k.currentRewardsAuction()
	if id != msg.Id {
		return errors.New("bad auction ID, can only bid in the current auction = " + strconv.Itoa(int(id)))
	}

	if lastBid.Bidder != msg.Sender {
		if err = k.sendCoins(vault, prevBidder, umeeCoins(lastBid.Amount)); err != nil {
			return err
		}
		if err = k.sendCoins(sender, vault, sdk.Coins{msg.Amount}); err != nil {
			return err
		}
	} else {
		diff := msg.Amount.SubAmount(minBid)
		if err = k.sendCoins(sender, vault, sdk.Coins{diff}); err != nil {
			return err
		}
	}

	bid := auction.Bid{Bidder: msg.Sender, Amount: msg.Amount.Amount}
	return store.SetValue(k.store, keyRewardsHighestBid, &bid, keyMsg)
}



func (k Keeper) currentRewardsAuction() (*auction.QueryRewardsAuctionResponse, error) {
	panic("not implemented")
}
