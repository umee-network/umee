package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/ibctransfer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ ibctransfer.QueryServer = Querier{}

// Querier implements a QueryServer for the x/ibc-rate-limit module.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// Params returns params of the x/ibc-rate-limit module.
func (q Querier) Params(goCtx context.Context, req *ibctransfer.QueryParams) (
	*ibctransfer.QueryParamsResponse, error,
) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper.GetParams(ctx)

	return &ibctransfer.QueryParamsResponse{Params: params}, nil
}

// RateLimitsOfIBCDenoms returns rate limits of ibc denoms.
func (q Querier) RateLimitsOfIBCDenoms(goCtx context.Context, req *ibctransfer.QueryRateLimitsOfIBCDenoms) (
	*ibctransfer.QueryRateLimitsOfIBCDenomsResponse, error,
) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if len(req.IbcDenom) == 0 {
		rateLimits, err := q.Keeper.GetRateLimitsOfIBCDenoms(ctx)
		if err != nil {
			return &ibctransfer.QueryRateLimitsOfIBCDenomsResponse{}, err
		}
		return &ibctransfer.QueryRateLimitsOfIBCDenomsResponse{RateLimits: rateLimits}, nil
	}

	rateLimit, err := q.Keeper.GetRateLimitsOfIBCDenom(ctx, req.IbcDenom)
	if err != nil {
		return &ibctransfer.QueryRateLimitsOfIBCDenomsResponse{}, err
	}
	return &ibctransfer.QueryRateLimitsOfIBCDenomsResponse{RateLimits: []ibctransfer.RateLimit{*rateLimit}}, nil
}
