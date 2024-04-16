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
	id := k.currentRewardsAuction()
	if id != msg.Id {
		return errors.New("bad auction ID, can only bid in the current auction = " + strconv.Itoa(int(id)))
	}

	keyMsg := "auction.rewards.highest_bid"
	key := k.keyRewardsBid(msg.Id)
	lastBid := store.GetValue[*auction.Bid](k.store, key, keyMsg)
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
	return store.SetValue(k.store, key, &bid, keyMsg)
}

func (k Keeper) getRewardsBid(id uint32) (*auction.Bid, uint32) {
	if id == 0 {
		id = k.currentRewardsAuction()
	}
	keyMsg := "auction.rewards.bid"
	key := k.keyRewardsBid(id)
	return store.GetValue[*auction.Bid](k.store, key, keyMsg), id
}

func (k Keeper) getRewards(id uint32) (*auction.Bid, uint32) {
	if id == 0 {
		id = k.currentRewardsAuction()
	}
	keyMsg := "auction.rewards.coins"
	key := k.keyRewardsCoins(id)
	return store.GetValue[*sdk.Coins](k.store, key, keyMsg), id
}
