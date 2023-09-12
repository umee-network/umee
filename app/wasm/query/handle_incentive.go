package query

import (
	"context"

	"github.com/cosmos/gogoproto/proto"
	"github.com/umee-network/umee/v6/x/incentive"
)

// HandleLeverageParams handles the get the x/leverage module's parameters.
func (q UmeeQuery) HandleIncentiveParams(
	ctx context.Context,
	qs incentive.QueryServer,
) (proto.Message, error) {
	return qs.Params(ctx, &incentive.QueryParams{})
}

// HandleTotalBonded handles the get the sum of all bonded collateral uTokens.
func (q UmeeQuery) HandleTotalBonded(
	ctx context.Context,
	qs incentive.QueryServer,
) (proto.Message, error) {
	return qs.TotalBonded(ctx, &incentive.QueryTotalBonded{Denom: q.TotalBonded.Denom})
}

// HandleTotalUnbonding handles the get the sum of all unbonding collateral uTokens.
func (q UmeeQuery) HandleTotalUnbonding(
	ctx context.Context,
	qs incentive.QueryServer,
) (proto.Message, error) {
	return qs.TotalUnbonding(ctx, &incentive.QueryTotalUnbonding{Denom: q.TotalUnbonding.Denom})
}

// HandleAccountBonds handles the get all bonded collateral and unbondings associated with an account.
func (q UmeeQuery) HandleAccountBonds(
	ctx context.Context,
	qs incentive.QueryServer,
) (proto.Message, error) {
	return qs.AccountBonds(ctx, &incentive.QueryAccountBonds{Address: q.AccountBonds.Address})
}

// HandlePendingRewards handles the gets all unclaimed incentive rewards associated with an account.
func (q UmeeQuery) HandlePendingRewards(
	ctx context.Context,
	qs incentive.QueryServer,
) (proto.Message, error) {
	return qs.PendingRewards(ctx, &incentive.QueryPendingRewards{Address: q.PendingRewards.Address})
}

// HandleCompletedIncentivePrograms handles the get all incentives programs that have been passed
// by governance,
func (q UmeeQuery) HandleCompletedIncentivePrograms(
	ctx context.Context,
	qs incentive.QueryServer,
) (proto.Message, error) {
	return qs.CompletedIncentivePrograms(ctx, &incentive.QueryCompletedIncentivePrograms{})
}

// HandleOngoingIncentivePrograms handles the get all incentives programs that have been passed
// by governance, funded, and started but not yet completed.
func (q UmeeQuery) HandleOngoingIncentivePrograms(
	ctx context.Context,
	qs incentive.QueryServer,
) (proto.Message, error) {
	return qs.OngoingIncentivePrograms(ctx, &incentive.QueryOngoingIncentivePrograms{})
}

// HandleUpcomingIncentivePrograms handles the get all incentives programs that have been passed
// by governance, but not yet started. They may or may not have been funded.
func (q UmeeQuery) HandleUpcomingIncentivePrograms(
	ctx context.Context,
	qs incentive.QueryServer,
) (proto.Message, error) {
	return qs.UpcomingIncentivePrograms(ctx, &incentive.QueryUpcomingIncentivePrograms{})
}

// HandleIncentiveProgram handles the get a single incentive program by ID.
func (q UmeeQuery) HandleIncentiveProgram(
	ctx context.Context,
	qs incentive.QueryServer,
) (proto.Message, error) {
	return qs.IncentiveProgram(ctx, &incentive.QueryIncentiveProgram{Id: q.IncentiveProgram.Id})
}

// HandleCurrentRates handles the get current rates of given denom
func (q UmeeQuery) HandleCurrentRates(
	ctx context.Context,
	qs incentive.QueryServer,
) (proto.Message, error) {
	return qs.CurrentRates(ctx, &incentive.QueryCurrentRates{UToken: q.CurrentRates.UToken})
}

// HandleActualRates handles the get the actutal rates of given denom
func (q UmeeQuery) HandleActualRates(
	ctx context.Context,
	qs incentive.QueryServer,
) (proto.Message, error) {
	return qs.ActualRates(ctx, &incentive.QueryActualRates{UToken: q.CurrentRates.UToken})
}

// HandleLastRewardTime handles the get last block time at which incentive rewards were calculated.
func (q UmeeQuery) HandleLastRewardTime(
	ctx context.Context,
	qs incentive.QueryServer,
) (proto.Message, error) {
	return qs.LastRewardTime(ctx, &incentive.QueryLastRewardTime{})
}
