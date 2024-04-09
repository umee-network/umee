package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/x/auction"
)

var _ auction.QueryServer = Querier{}

// Querier implements a QueryServer for the x/uibc module.
type Querier struct {
	Builder
}

func NewQuerier(kb Builder) Querier {
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
	// ctx := sdk.UnwrapSDKContext(goCtx)
	// b := q.Keeper(&ctx).GetParams()

	return &auction.QueryRewardsAuctionResponse{Params: params}, nil
}
