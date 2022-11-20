package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/v3/x/incentive"
)

var _ incentive.QueryServer = Querier{}

// Querier implements a QueryServer for the x/incentive module.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

func (q Querier) Params(
	goCtx context.Context,
	req *incentive.QueryParamsRequest,
) (*incentive.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper.GetParams(ctx)

	return &incentive.QueryParamsResponse{Params: params}, nil
}

func (q Querier) IncentiveProgram(
	goCtx context.Context,
	req *incentive.QueryIncentiveProgramRequest,
) (*incentive.QueryIncentiveProgramResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	program := incentive.IncentiveProgram{}

	resp := &incentive.QueryIncentiveProgramResponse{
		Program: program,
	}

	return resp, incentive.ErrNotImplemented
}

func (q Querier) IncentivePrograms(
	goCtx context.Context,
	req *incentive.QueryIncentiveProgramsRequest,
) (*incentive.QueryIncentiveProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// TODO: get all programs (also: pagination)

	resp := &incentive.QueryIncentiveProgramsResponse{
		Programs: make([]incentive.IncentiveProgram, 0),
	}

	return resp, incentive.ErrNotImplemented
}

// TODO: other queries
