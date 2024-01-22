package quota

import (
	context "context"

	sdkmath "cosmossdk.io/math"
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
	var o sdkmath.LegacyDec
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
	k := q.Keeper(&ctx)
	o, err := k.GetAllOutflows()
	if err != nil {
		return nil, err
	}
	outflows := k.coinsWithTokenSymbols(ctx, o)
	return &uibc.QueryAllOutflowsResponse{Outflows: outflows}, nil
}

// Inflows returns sum of inflows and registered IBC denoms inflows in the current quota period.
func (q Querier) Inflows(goCtx context.Context, req *uibc.QueryInflows) (*uibc.QueryInflowsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var (
		inflowCoins sdk.DecCoins
		err         error
		k           = q.Keeper(&ctx)
	)

	if len(req.Denom) != 0 {
		inflowCoins = sdk.NewDecCoins(k.GetTokenInflow(req.Denom))
	} else {
		inflowCoins, err = k.GetAllInflows()
		if err != nil {
			return nil, err
		}
	}
	inflows := k.coinsWithTokenSymbols(ctx, inflowCoins)
	return &uibc.QueryInflowsResponse{Sum: q.Keeper(&ctx).GetInflowSum(), Inflows: inflows}, nil
}

// QuotaExpires returns the current ibc quota expire time.
func (q Querier) QuotaExpires(goCtx context.Context, _ *uibc.QueryQuotaExpires) (*uibc.QueryQuotaExpiresResponse,
	error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	quotaExpireTime, err := q.Keeper(&ctx).GetExpire()
	if err != nil {
		return nil, err
	}

	return &uibc.QueryQuotaExpiresResponse{EndTime: *quotaExpireTime}, nil
}
