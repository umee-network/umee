package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/v6/util/coin"
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
