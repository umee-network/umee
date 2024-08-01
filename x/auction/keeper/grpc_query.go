package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/auction"
)

// Querier implements a QueryServer for the x/uibc module.
type Querier struct {
	Builder
}

func NewQuerier(kb Builder) auction.QueryServer {
	return Querier{Builder: kb}
}

// Params returns params of the x/auction module.
func (q Querier) RewardsParams(goCtx context.Context, _ *auction.QueryRewardsParams) (
	*auction.QueryRewardsParamsResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper(&ctx).GetRewardsParams()

	return &auction.QueryRewardsParamsResponse{Params: params}, nil
}

// RewardsAuction returns params of the x/auction module.
// Bidder and Bid are nul if there is no active auction.
func (q Querier) RewardsAuction(goCtx context.Context, msg *auction.QueryRewardsAuction) (
	*auction.QueryRewardsAuctionResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k := q.Keeper(&ctx)
	rewards, id := k.getRewardsAuction(msg.Id)
	if rewards == nil {
		return nil, status.Error(codes.NotFound, "wrong ID")
	}
	r := &auction.QueryRewardsAuctionResponse{Id: id}
	r.Rewards = rewards.Rewards
	r.EndsAt = rewards.EndsAt

	bid := q.Keeper(&ctx).getRewardsBid(id)
	if bid != nil {
		r.Bidder = bid.Bidder
		r.Bid = coin.UmeeInt(bid.Amount)
	}

	return r, nil
}

// RewardsAuctions returns all the auctions with bids and rewards
func (q Querier) RewardsAuctions(goCtx context.Context, req *auction.QueryRewardsAuctions) (
	*auction.QueryRewardsAuctionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k := q.Keeper(&ctx)
	rewardsStore := prefix.NewStore(k.store, keyPrefixRewardsCoins)
	auctions := make([]*auction.QueryRewardsAuctionResponse, 0)

	pageRes, err := query.Paginate(rewardsStore, req.Pagination, func(key, value []byte) error {
		var rewards auction.Rewards
		err := rewards.Unmarshal(value)
		if err != nil {
			return err
		}
		var auctionID store.Uint32
		err = auctionID.Unmarshal(key)
		if err != nil {
			return err
		}
		auction := &auction.QueryRewardsAuctionResponse{
			Id: uint32(auctionID),
		}
		auction.Rewards = rewards.Rewards
		auction.EndsAt = rewards.EndsAt
		auctions = append(auctions, auction)
		return nil
	})

	if err != nil {
		return nil, err
	}

	for index, auction := range auctions {
		// get the bids info
		bid := k.getRewardsBid(auction.Id)
		if bid != nil {
			auctions[index].Bidder = bid.Bidder
			auctions[index].Bid = coin.UmeeInt(bid.Amount)
		}
	}

	return &auction.QueryRewardsAuctionsResponse{
		Auctions:   auctions,
		Pagination: pageRes,
	}, nil
}
