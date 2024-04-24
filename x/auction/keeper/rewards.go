package keeper

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/auction"
)

func (k Keeper) FinalizeRewardsAuction() error {
	now := k.ctx.BlockTime()
	a, id := k.getRewardsAuction(0)
	if !a.EndsAt.After(now) {
		return nil
	}

	newCoins := k.bank.GetAllBalances(*k.ctx, k.accs.RewardsCollect)
	bid, _ := k.getRewardsBid(id)
	if len(bid.Bidder) == 0 {
		err := k.sendCoins(k.accs.RewardsCollect, bid.Bidder, a.Rewards)
		if err != nil {
			return fmt.Errorf("can't send coins to finalize the auction [%w]", err)
		}
		err = k.bank.BurnCoins(*k.ctx, auction.ModuleName, sdk.Coins{coin.UmeeInt(bid.Amount)})
		if err != nil {
			return fmt.Errorf("can't burn rewards auction bid [%w]", err)
		}

	} else if len(a.Rewards) != 0 {
		// rollover the past rewards if there was no bidder
		newCoins = newCoins.Add(a.Rewards...)
	}

	id++
	store.SetInteger(k.store, keyRewardsCurrentID, id)
	params := k.GetRewardsParams()
	endsAt := now.Add(time.Duration(params.BidDuration) * time.Second)
	return k.storeNewRewardsAuction(id, endsAt, newCoins)
}

func (k Keeper) currentRewardsAuction() uint32 {
	id, _ := store.GetInteger[uint32](k.store, keyRewardsCurrentID)
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
	if err := auction.ValidateMinRewardsBid(minBid, msg.Amount); err != nil {
		return err
	}

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	util.Panic(err)

	if !bytes.Equal(sender, lastBid.Bidder) {
		returned := coin.UmeeInt(lastBid.Amount)
		if err = k.sendFromModule(lastBid.Bidder, returned); err != nil {
			return err
		}
		if err = k.sendToModule(sender, msg.Amount); err != nil {
			return err
		}
	} else {
		diff := msg.Amount.SubAmount(lastBid.Amount)
		if err = k.sendToModule(sender, diff); err != nil {
			return err
		}
	}

	bid := auction.Bid{Bidder: sender, Amount: msg.Amount.Amount}
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

// returns nil if id is not found
func (k Keeper) getRewardsAuction(id uint32) (*auction.Rewards, uint32) {
	if id == 0 {
		id = k.currentRewardsAuction()
	}
	const keyMsg = "auction.rewards.coins"
	key := k.keyRewardsCoins(id)
	return store.GetValue[*auction.Rewards](k.store, key, keyMsg), id
}

func (k Keeper) storeNewRewardsAuction(id uint32, endsAt time.Time, coins sdk.Coins) error {
	newRewards := auction.Rewards{EndsAt: endsAt, Rewards: coins}
	const keyMsg = "auction.rewards.coins"
	key := k.keyRewardsCoins(id)
	return store.SetValue(k.store, key, &newRewards, keyMsg)
}
