package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/v3/x/incentive/types"
)

var _ types.QueryServer = Querier{}

// Querier implements a QueryServer for the x/incentive module.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

func (q Querier) Params(
	goCtx context.Context,
	req *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (q Querier) IncentiveProgram(
	goCtx context.Context,
	req *types.QueryIncentiveProgramRequest,
) (*types.QueryIncentiveProgramResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	program := types.IncentiveProgram{}

	resp := &types.QueryIncentiveProgramResponse{
		Program: program,
	}

	return resp, types.ErrNotImplemented
}

func (q Querier) IncentivePrograms(
	goCtx context.Context,
	req *types.QueryIncentiveProgramsRequest,
) (*types.QueryIncentiveProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// TODO: get all programs (also: pagination)

	resp := &types.QueryIncentiveProgramsResponse{
		Programs: make([]types.IncentiveProgram, 0),
	}

	return resp, types.ErrNotImplemented
}

// TODO: other queries
