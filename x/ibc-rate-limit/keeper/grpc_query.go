package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Querier{}

// Querier implements a QueryServer for the x/ibc-rate-limit module.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// Params implements types.QueryServer
func (q Querier) Params(goCtx context.Context, req *types.QueryParams) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// RateLimitsOfIBCDenom implements types.QueryServer
func (q Querier) RateLimitsOfIBCDenom(goCtx context.Context, req *types.QueryRateLimitsOfIBCDenom) (*types.QueryRateLimitsOfIBCDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	rateLimit, err := q.Keeper.GetRateLimitsOfIBCDenom(ctx, req.IbcDenom)
	if err != nil {
		return &types.QueryRateLimitsOfIBCDenomResponse{}, nil
	}

	return &types.QueryRateLimitsOfIBCDenomResponse{RateLimit: rateLimit}, nil
}

// RateLimitsOfIBCDenoms implements types.QueryServer
func (q Querier) RateLimitsOfIBCDenoms(goCtx context.Context, req *types.QueryRateLimitsOfIBCDenoms) (*types.QueryRateLimitsOfIBCDenomsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	rateLimits, err := q.Keeper.GetRateLimitsOfIBCDenoms(ctx)
	if err != nil {
		return &types.QueryRateLimitsOfIBCDenomsResponse{}, err
	}

	return &types.QueryRateLimitsOfIBCDenomsResponse{RateLimits: rateLimits}, nil
}
