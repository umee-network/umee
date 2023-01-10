package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/uibc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
func (q Querier) Params(goCtx context.Context, req *uibc.QueryParams) (
	*uibc.QueryParamsResponse, error,
) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper.GetParams(ctx)

	return &uibc.QueryParamsResponse{Params: params}, nil
}

// Quota returns quota of ibc denoms.
func (q Querier) Quota(goCtx context.Context, req *uibc.QueryQuota) (
	*uibc.QueryQuotaResponse, error,
) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if len(req.IbcDenom) == 0 {
		quotaOfIBCDenoms, err := q.Keeper.GetQuotaOfIBCDenoms(ctx)
		if err != nil {
			return &uibc.QueryQuotaResponse{}, err
		}
		return &uibc.QueryQuotaResponse{Quota: quotaOfIBCDenoms}, nil
	}

	quotaOfIBCDenom, err := q.Keeper.GetQuotaOfIBCDenom(ctx, req.IbcDenom)
	if err != nil {
		return &uibc.QueryQuotaResponse{}, err
	}
	return &uibc.QueryQuotaResponse{Quota: []uibc.Quota{*quotaOfIBCDenom}}, nil
}
