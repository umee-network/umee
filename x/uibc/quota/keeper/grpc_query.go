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

// Quota returns quotas of denoms.
func (q Querier) Quota(goCtx context.Context, req *uibc.QueryQuota) (
	*uibc.QueryQuotaResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if len(req.Denom) == 0 {
		quotas, err := q.GetAllQuotas(ctx)
		if err != nil {
			return &uibc.QueryQuotaResponse{}, err
		}
		return &uibc.QueryQuotaResponse{Quotas: quotas}, nil
	}

	quota, err := q.GetQuota(ctx, req.Denom)
	if err != nil {
		return &uibc.QueryQuotaResponse{}, err
	}
	return &uibc.QueryQuotaResponse{Quotas: sdk.DecCoins{quota}}, nil
}
