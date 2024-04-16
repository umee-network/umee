package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
func (q Querier) RewardsAuction(goCtx context.Context, msg *auction.QueryRewardsAuction) (
	*auction.QueryRewardsAuctionResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	bid, id := q.Keeper(&ctx).getRewardsBid(msg.Id)
	r := &auction.QueryRewardsAuctionResponse{Id: id}
	if bid != nil {
		r.Bidder = bid.Bidder
		r.Bid = coin.UmeeInt(bid.Amount)
	}
	rewards, _ := q.Keeper(&ctx).getRewards(msg.Id)
	if rewards != nil {
		r.Rewards = rewards.Rewards
		r.EndsAt = rewards.EndsAt
	}

	return r, nil
}
