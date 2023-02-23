package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/uibc"
)

var _ uibc.QueryServer = Querier{}

// Querier implements a QueryServer for the x/uibc module.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// Params returns params of the x/uibc module.
func (q Querier) Params(goCtx context.Context, _ *uibc.QueryParams) (
	*uibc.QueryParamsResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.GetParams(ctx)

	return &uibc.QueryParamsResponse{Params: params}, nil
}

// Outflows queries denom outflows.
func (q Querier) Outflows(goCtx context.Context, req *uibc.QueryOutflows) (
	*uibc.QueryOutflowsResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if len(req.Denom) == 0 {
		o, err := q.GetAllOutflows(ctx)
		if err != nil {
			return &uibc.QueryOutflowsResponse{}, err
		}
		return &uibc.QueryOutflowsResponse{Outflows: o}, nil
	}

	o, err := q.GetOutflows(ctx, req.Denom)
	if err != nil {
		return &uibc.QueryOutflowsResponse{}, err
	}
	return &uibc.QueryOutflowsResponse{Outflows: sdk.DecCoins{o}}, nil
}
