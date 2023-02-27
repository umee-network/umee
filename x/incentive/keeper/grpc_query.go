package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/v4/x/incentive"
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
	req *incentive.QueryParams,
) (*incentive.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper.GetParams(ctx)

	return &incentive.QueryParamsResponse{Params: params}, nil
}

func (q Querier) IncentiveProgram(
	_ context.Context,
	req *incentive.QueryIncentiveProgram,
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

func (q Querier) UpcomingIncentivePrograms(
	_ context.Context,
	req *incentive.QueryUpcomingIncentivePrograms,
) (*incentive.QueryUpcomingIncentiveProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// TODO: get all programs

	resp := &incentive.QueryUpcomingIncentiveProgramsResponse{
		Programs: make([]incentive.IncentiveProgram, 0),
	}

	return resp, incentive.ErrNotImplemented
}

func (q Querier) OngoingIncentivePrograms(
	_ context.Context,
	req *incentive.QueryOngoingIncentivePrograms,
) (*incentive.QueryOngoingIncentiveProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// TODO: get all programs

	resp := &incentive.QueryOngoingIncentiveProgramsResponse{
		Programs: make([]incentive.IncentiveProgram, 0),
	}

	return resp, incentive.ErrNotImplemented
}

func (q Querier) CompletedIncentivePrograms(
	_ context.Context,
	req *incentive.QueryCompletedIncentivePrograms,
) (*incentive.QueryCompletedIncentiveProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// TODO: get all programs (also: pagination)

	resp := &incentive.QueryCompletedIncentiveProgramsResponse{
		Programs: make([]incentive.IncentiveProgram, 0),
		// TODO: pagination
	}

	return resp, incentive.ErrNotImplemented
}

func (q Querier) PendingRewards(
	_ context.Context,
	req *incentive.QueryPendingRewards,
) (*incentive.QueryPendingRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// TODO: calculate, without modifying, rewards which would result from MsgClaim

	return &incentive.QueryPendingRewardsResponse{}, incentive.ErrNotImplemented
}

func (q Querier) Bonded(
	_ context.Context,
	req *incentive.QueryBonded,
) (*incentive.QueryBondedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// TODO: get one or all denoms, all tiers bonded to this address

	return &incentive.QueryBondedResponse{}, incentive.ErrNotImplemented
}

func (q Querier) TotalBonded(
	_ context.Context,
	req *incentive.QueryTotalBonded,
) (*incentive.QueryTotalBondedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// TODO: bonded uTokens across one or all denoms, all tiers

	return &incentive.QueryTotalBondedResponse{}, incentive.ErrNotImplemented
}

func (q Querier) Unbondings(
	_ context.Context,
	req *incentive.QueryUnbondings,
) (*incentive.QueryUnbondingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// TODO: get all unbondings associated with a single address

	return &incentive.QueryUnbondingsResponse{}, incentive.ErrNotImplemented
}
