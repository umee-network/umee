package keeper

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/auction"
)

func (k Keeper) FinalizeRewardsAuction() error {
	now := k.ctx.BlockTime()
	a, id := k.getRewardsAuction(0)
	if a == nil {
		return k.initNewAuction(id+1, []sdk.Coin{})
	}

	if a.EndsAt.After(now) {
		return nil
	}

	bid, _ := k.getRewardsBid(id)
	if bid != nil && len(bid.Bidder) != 0 {
		bidderAccAddr, err := sdk.AccAddressFromBech32(bid.Bidder)
		if err != nil {
			return err
		}
		err = k.sendCoins(k.accs.RewardsCollect, bidderAccAddr, a.Rewards)
		if err != nil {
			return fmt.Errorf("can't send coins to finalize the auction [%w]", err)
		}
		err = k.bank.BurnCoins(*k.ctx, auction.ModuleName, sdk.Coins{coin.UmeeInt(bid.Amount)})
		if err != nil {
			return fmt.Errorf("can't burn rewards auction bid [%w]", err)
		}
		sdkutil.Emit(k.ctx, &auction.EventRewardsAuctionResult{
			Id:     id,
			Bidder: sdk.AccAddress(bid.Bidder).String(),
		})
	}

	remainingRewards := k.bank.GetAllBalances(*k.ctx, k.accs.RewardsCollect)
	return k.initNewAuction(id+1, remainingRewards)
}

func (k Keeper) initNewAuction(id uint32, rewards sdk.Coins) error {
	store.SetInteger(k.store, keyRewardsCurrentID, id)
	params := k.GetRewardsParams()
	endsAt := k.ctx.BlockTime().Add(time.Duration(params.BidDuration) * time.Second)
	return k.storeNewRewardsAuction(id, endsAt, rewards)
}

func (k Keeper) currentRewardsAuctionID() uint32 {
	id, _ := store.GetInteger[uint32](k.store, keyRewardsCurrentID)
	return id
}

func (k Keeper) rewardsBid(msg *auction.MsgRewardsBid) error {
	id := k.currentRewardsAuctionID()
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

	toAuction := msg.Amount
	if lastBid != nil {
		if msg.Sender == lastBid.Bidder {
			// bidder updates his last bid: send only diff
			toAuction = msg.Amount.SubAmount(lastBid.Amount)
		} else {
			returned := coin.UmeeInt(lastBid.Amount)
			bidderAccAddr, err := sdk.AccAddressFromBech32(lastBid.Bidder)
			if err != nil {
				return err
			}
			if err = k.sendFromModule(bidderAccAddr, returned); err != nil {
				return err
			}
		}
	}

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}

	if err = k.sendToModule(sender, toAuction); err != nil {
		return err
	}

	bid := auction.Bid{Bidder: msg.Sender, Amount: msg.Amount.Amount}
	return store.SetValue(k.store, key, &bid, keyMsg)
}

func (k Keeper) getRewardsBid(id uint32) (*auction.Bid, uint32) {
	if id == 0 {
		id = k.currentRewardsAuctionID()
	}
	keyMsg := "auction.rewards.bid"
	key := k.keyRewardsBid(id)
	return store.GetValue[*auction.Bid](k.store, key, keyMsg), id
}

func (k Keeper) getAllRewardsBids() ([]auction.BidKV, error) {
	elems, err := store.LoadAllKV[*store.Uint32, store.Uint32, *auction.Bid](k.store, keyPrefixRewardsBid)
	if err != nil {
		return nil, err
	}
	bids := make([]auction.BidKV, len(elems))
	for i := range elems {
		bids[i].Id = uint32(elems[i].Key)
		bids[i].Bid = elems[i].Val
	}
	return bids, nil
}

func (k Keeper) storeAllRewardsBids(elems []auction.BidKV) error {
	for _, e := range elems {
		key := k.keyRewardsBid(e.Id)
		if err := store.SetValue(k.store, key, &e.Bid, "auction.store_all_bids"); err != nil {
			return err
		}
	}
	return nil
}

// Returns collected protocol rewards to auction.
// Returns nil if id is not found.
func (k Keeper) getRewardsAuction(id uint32) (*auction.Rewards, uint32) {
	if id == 0 {
		id = k.currentRewardsAuctionID()
	}
	const keyMsg = "auction.rewards.coins"
	key := k.keyRewardsCoins(id)
	return store.GetValue[*auction.Rewards](k.store, key, keyMsg), id
}

func (k Keeper) getAllRewardsAuctions() ([]auction.RewardsKV, error) {
	elems, err := store.LoadAllKV[*store.Uint32, store.Uint32, *auction.Rewards](
		k.store, keyPrefixRewardsCoins)
	if err != nil {
		return nil, err
	}
	rewards := make([]auction.RewardsKV, len(elems))
	for i := range elems {
		rewards[i].Id = uint32(elems[i].Key)
		rewards[i].Rewards = elems[i].Val
	}
	return rewards, nil
}

func (k Keeper) storeNewRewardsAuction(id uint32, endsAt time.Time, coins []sdk.Coin) error {
	newRewards := auction.Rewards{EndsAt: endsAt, Rewards: coins}
	const keyMsg = "auction.rewards.coins"
	key := k.keyRewardsCoins(id)
	return store.SetValue(k.store, key, &newRewards, keyMsg)
}

func (k Keeper) storeAllRewardsAuctions(elems []auction.RewardsKV) error {
	for _, e := range elems {
		key := k.keyRewardsCoins(e.Id)
		if err := store.SetValue(k.store, key, &e.Rewards, "auction.store_all_rewards"); err != nil {
			return err
		}
	}
	return nil
}
