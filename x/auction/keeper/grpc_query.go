package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
func (q Querier) RewardsAuction(goCtx context.Context, _ *auction.QueryRewardsAuction) (
	*auction.QueryRewardsAuctionResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	b := q.Keeper(&ctx)
	return b.currentRewardsAuction()
}
