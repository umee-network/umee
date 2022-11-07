package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Querier{}

// Querier implements a QueryServer for the x/leverage module.
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

// RateLimits implements types.QueryServer
func (Querier) RateLimits(context.Context, *types.QueryRateLimits) (*types.QueryRateLimitsResponse, error) {
	return &types.QueryRateLimitsResponse{}, nil
}
