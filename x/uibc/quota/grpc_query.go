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

// Inflows returns sum of inflows and registered IBC denoms inflows in the current quota period.
func (q Querier) Inflows(goCtx context.Context, req *uibc.QueryInflows) (*uibc.QueryInflowsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var (
		inflows sdk.DecCoins
		err     error
	)

	if len(req.Denom) != 0 {
		inflows = append(inflows, q.Keeper(&ctx).GetTokenInflow(req.Denom))
	} else {
		inflows, err = q.Keeper(&ctx).GetAllInflows()
		if err != nil {
			return nil, err
		}
	}
	return &uibc.QueryInflowsResponse{Sum: q.Keeper(&ctx).GetInflowSum(), Inflows: inflows}, nil
}

// QuotaExpires returns the current ibc quota expire time.
func (q Querier) QuotaExpires(goCtx context.Context, _ *uibc.QueryQuotaExpires) (*uibc.QueryQuotaExpiresResponse,
	error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	quotaExpireTime, err := q.Keeper(&ctx).GetExpire()
	if err != nil {
		return nil, err
	}

	return &uibc.QueryQuotaExpiresResponse{EndTime: *quotaExpireTime}, nil
}
