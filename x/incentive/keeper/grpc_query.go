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

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)

	return &incentive.QueryParamsResponse{Params: params}, nil
}

func (q Querier) IncentiveProgram(
	goCtx context.Context,
	req *incentive.QueryIncentiveProgram,
) (*incentive.QueryIncentiveProgramResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)

	program, _, err := k.GetIncentiveProgram(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	resp := &incentive.QueryIncentiveProgramResponse{
		Program: program,
	}

	return resp, nil
}

func (q Querier) UpcomingIncentivePrograms(
	goCtx context.Context,
	req *incentive.QueryUpcomingIncentivePrograms,
) (*incentive.QueryUpcomingIncentiveProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)

	programs, err := k.GetAllIncentivePrograms(ctx, incentive.ProgramStatusUpcoming)
  if err != nil {
		return nil, err
	}

	resp := &incentive.QueryUpcomingIncentiveProgramsResponse{
		Programs: programs,
	}

	return resp, incentive.ErrNotImplemented
}

func (q Querier) OngoingIncentivePrograms(
	goCtx context.Context,
	req *incentive.QueryOngoingIncentivePrograms,
) (*incentive.QueryOngoingIncentiveProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)

	 programs, err := q.Keeper.getAllIncentivePrograms(ctx, incentive.ProgramStatusOngoing)
	if err != nil {
		return nil, err
	}

	resp := &incentive.QueryOngoingIncentiveProgramsResponse{
		Programs: programs,
	}

	return resp, nil
}

func (q Querier) CompletedIncentivePrograms(
	goCtx context.Context,
	req *incentive.QueryCompletedIncentivePrograms,
) (*incentive.QueryCompletedIncentiveProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)
  
	programs, err := q.Keeper.getPaginatedIncentivePrograms(
		ctx,
		incentive.ProgramStatusCompleted,
		req.Pagination.Offset,
		req.Pagination.Limit,
	)
	if err != nil {
		return nil, err
	}

	resp := &incentive.QueryCompletedIncentiveProgramsResponse{
		Programs: programs,
	}

	return resp, nil
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
	goCtx context.Context,
	req *incentive.QueryUnbondings,
) (*incentive.QueryUnbondingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	unbondings := k.GetUnbondings(ctx, addr)

	return &incentive.QueryUnbondingsResponse{Unbondings: unbondings}, incentive.ErrNotImplemented
}
