package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/v5/x/incentive"
	leveragetypes "github.com/umee-network/umee/v5/x/leverage/types"
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

	program, _, err := k.getIncentiveProgram(ctx, req.Id)
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

	programs, err := k.getAllIncentivePrograms(ctx, incentive.ProgramStatusUpcoming)
	if err != nil {
		return nil, err
	}

	resp := &incentive.QueryUpcomingIncentiveProgramsResponse{
		Programs: programs,
	}

	return resp, err
}

func (q Querier) OngoingIncentivePrograms(
	goCtx context.Context,
	req *incentive.QueryOngoingIncentivePrograms,
) (*incentive.QueryOngoingIncentiveProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)

	programs, err := k.getAllIncentivePrograms(ctx, incentive.ProgramStatusOngoing)
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

	programs, err := k.getAllIncentivePrograms(
		ctx,
		incentive.ProgramStatusCompleted,
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
	goCtx context.Context,
	req *incentive.QueryPendingRewards,
) (*incentive.QueryPendingRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)
	pending, err := k.calculateRewards(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &incentive.QueryPendingRewardsResponse{Rewards: pending}, err
}

func (q Querier) AccountBonds(
	goCtx context.Context,
	req *incentive.QueryAccountBonds,
) (*incentive.QueryAccountBondsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	totalBonded := sdk.NewCoins()
	totalUnbonding := sdk.NewCoins()
	accountUnbondings := []incentive.Unbonding{}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)
	denoms, err := k.getAllBondDenoms(ctx, addr)
	if err != nil {
		return nil, err
	}
	for _, denom := range denoms {
		bonded, unbonding, unbondings := k.BondSummary(ctx, addr, denom)
		totalBonded = totalBonded.Add(bonded)
		totalUnbonding = totalUnbonding.Add(unbonding)
		// Only nonzero unbondings will be stored, so this list is already filtered
		accountUnbondings = append(accountUnbondings, unbondings...)
	}

	return &incentive.QueryAccountBondsResponse{
		Bonded:     totalBonded,
		Unbonding:  totalUnbonding,
		Unbondings: accountUnbondings,
	}, nil
}

func (q Querier) TotalBonded(
	goCtx context.Context,
	req *incentive.QueryTotalBonded,
) (*incentive.QueryTotalBondedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)

	var total sdk.Coins
	if req.Denom != "" {
		total = sdk.NewCoins(k.getTotalBonded(ctx, req.Denom))
	} else {
		total = k.getAllTotalBonded(ctx)
	}

	return &incentive.QueryTotalBondedResponse{Bonded: total}, nil
}

func (q Querier) TotalUnbonding(
	goCtx context.Context,
	req *incentive.QueryTotalUnbonding,
) (*incentive.QueryTotalUnbondingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)

	var total sdk.Coins
	if req.Denom != "" {
		total = sdk.NewCoins(k.getTotalUnbonding(ctx, req.Denom))
	} else {
		total = k.getAllTotalUnbonding(ctx)
	}

	return &incentive.QueryTotalUnbondingResponse{Unbonding: total}, nil
}

func (q Querier) CurrentRates(
	goCtx context.Context,
	req *incentive.QueryCurrentRates,
) (*incentive.QueryCurrentRatesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)

	// extimate the annual rewards a reference amount of bonded uTokens would earn in a year
	referenceUToken, rewards, err := k.calculateReferenceAPY(ctx, req.UToken)
	if err != nil {
		return nil, err
	}

	return &incentive.QueryCurrentRatesResponse{
		ReferenceBond: referenceUToken,
		Rewards:       rewards,
	}, nil
}

func (q Querier) ActualRates(
	goCtx context.Context,
	req *incentive.QueryActualRates,
) (*incentive.QueryActualRatesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)

	// extimate the annual rewards a reference amount of bonded uTokens would earn in a year
	referenceUToken, rewards, err := k.calculateReferenceAPY(ctx, req.UToken)
	if err != nil {
		return nil, err
	}

	// compute oracle price ratio of rewards to reference bond amount
	referenceToken, err := k.leverageKeeper.ExchangeUToken(ctx, referenceUToken)
	if err != nil {
		return nil, err
	}

	referenceBondValue, err := k.leverageKeeper.TotalTokenValue(
		ctx, sdk.NewCoins(referenceToken), leveragetypes.PriceModeSpot,
	)
	if err != nil {
		return nil, err
	}
	referenceRewardValue, err := k.leverageKeeper.TotalTokenValue(ctx, rewards, leveragetypes.PriceModeSpot)
	if err != nil {
		return nil, err
	}
	if referenceBondValue.IsZero() {
		return nil, leveragetypes.ErrInvalidOraclePrice.Wrap(referenceToken.Denom)
	}

	return &incentive.QueryActualRatesResponse{
		APY: referenceRewardValue.Quo(referenceBondValue),
	}, nil
}

func (q Querier) LastRewardTime(
	goCtx context.Context,
	req *incentive.QueryLastRewardTime,
) (*incentive.QueryLastRewardTimeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	return &incentive.QueryLastRewardTimeResponse{
		Time: q.Keeper.GetLastRewardsTime(sdk.UnwrapSDKContext(goCtx)),
	}, nil
}
