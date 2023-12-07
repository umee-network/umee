package quota

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/x/uibc"
)

var _ uibc.QueryServer = Querier{}

// Querier implements a QueryServer for the x/uibc module.
type Querier struct {
	KeeperBuilder
}

func NewQuerier(kb KeeperBuilder) Querier {
	return Querier{KeeperBuilder: kb}
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
		o = k.GetOutflowSum()
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

// AllInflows implements uibc.QueryServer.
func (q Querier) AllInflows(goCtx context.Context, req *uibc.QueryAllInflows) (*uibc.QueryAllInflowsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var (
		inflows []sdk.DecCoin
		err     error
	)

	if len(req.Denom) != 0 {
		tokenInflow := q.Keeper(&ctx).GetTokenInflow(req.Denom)
		inflows = append(inflows, tokenInflow)
	} else {
		inflows, err = q.Keeper(&ctx).GetAllInflows()
		if err != nil {
			return nil, err
		}
	}

	return &uibc.QueryAllInflowsResponse{Inflows: inflows}, nil
}

// Inflows returns registered IBC denoms inflows in the current quota period.
// If denom is not specified, returns sum of all registered inflows.
func (q Querier) Inflows(goCtx context.Context, req *uibc.QueryInflows) (*uibc.QueryInflowsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var amount sdk.Dec
	if len(req.Denom) != 0 {
		tokenInflow := q.Keeper(&ctx).GetTokenInflow(req.Denom)
		amount = tokenInflow.Amount
	} else {
		amount = q.Keeper(&ctx).GetInflowSum()
	}
	return &uibc.QueryInflowsResponse{Amount: amount}, nil
}
