package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v5/x/uibc"
)

var _ uibc.QueryServer = Querier{}

// Querier implements a QueryServer for the x/uibc module.
type Querier struct {
	Builder
}

func NewQuerier(kb Builder) Querier {
	return Querier{Builder: kb}
}

// Params returns params of the x/uibc module.
func (q Querier) Params(goCtx context.Context, _ *uibc.QueryParams) (
	*uibc.QueryParamsResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper(&ctx).GetParams()

	return &uibc.QueryParamsResponse{Params: params}, nil
}

// Outflows queries denom outflows in the current period.
// If req.Denom is not set, then we return total outflows.
func (q Querier) Outflows(goCtx context.Context, req *uibc.QueryOutflows) (
	*uibc.QueryOutflowsResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k := q.Keeper(&ctx)
	var o sdk.Dec
	if len(req.Denom) == 0 {
		o = k.GetTotalOutflow()
	} else {
		d := k.GetTokenOutflows(req.Denom)
		o = d.Amount
	}

	return &uibc.QueryOutflowsResponse{Amount: o}, nil
}

// AllOutflows queries outflows for all denom in the current period.
func (q Querier) AllOutflows(goCtx context.Context, _ *uibc.QueryAllOutflows) (
	*uibc.QueryAllOutflowsResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	o, err := q.Keeper(&ctx).GetAllOutflows()
	if err != nil {
		return nil, err
	}
	return &uibc.QueryAllOutflowsResponse{Outflows: o}, nil
}
