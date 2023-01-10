package query

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	octypes "github.com/umee-network/umee/v4/x/oracle/types"
)

// HandleFeederDelegation gets all the feeder delegation of a validator.
func (q UmeeQuery) HandleFeederDelegation(
	ctx sdk.Context,
	qs octypes.QueryServer,
) (proto.Message, error) {
	req := &octypes.QueryFeederDelegation{ValidatorAddr: q.FeederDelegation.ValidatorAddr}
	return qs.FeederDelegation(sdk.WrapSDKContext(ctx), req)
}

// HandleMissCounter gets all the oracle miss counter of a validator.
func (q UmeeQuery) HandleMissCounter(
	ctx sdk.Context,
	qs octypes.QueryServer,
) (proto.Message, error) {
	req := &octypes.QueryMissCounter{ValidatorAddr: q.MissCounter.ValidatorAddr}
	return qs.MissCounter(sdk.WrapSDKContext(ctx), req)
}

// HandleSlashWindow gets slash window information.
func (q UmeeQuery) HandleSlashWindow(
	ctx sdk.Context,
	qs octypes.QueryServer,
) (proto.Message, error) {
	return qs.SlashWindow(sdk.WrapSDKContext(ctx), &octypes.QuerySlashWindow{})
}

// HandleAggregatePrevote gets an aggregate prevote of a validator.
func (q UmeeQuery) HandleAggregatePrevote(
	ctx sdk.Context,
	qs octypes.QueryServer,
) (proto.Message, error) {
	req := &octypes.QueryAggregatePrevote{ValidatorAddr: q.AggregatePrevote.ValidatorAddr}
	return qs.AggregatePrevote(sdk.WrapSDKContext(ctx), req)
}

// HandleAggregatePrevotes gets an aggregate prevote of all validators.
func (q UmeeQuery) HandleAggregatePrevotes(
	ctx sdk.Context,
	qs octypes.QueryServer,
) (proto.Message, error) {
	return qs.AggregatePrevotes(sdk.WrapSDKContext(ctx), &octypes.QueryAggregatePrevotes{})
}

// HandleAggregateVote gets an aggregate vote of a validator.
func (q UmeeQuery) HandleAggregateVote(
	ctx sdk.Context,
	qs octypes.QueryServer,
) (proto.Message, error) {
	req := &octypes.QueryAggregateVote{ValidatorAddr: q.AggregateVote.ValidatorAddr}
	return qs.AggregateVote(sdk.WrapSDKContext(ctx), req)
}

// HandleAggregateVotes gets an aggregate vote of all validators.
func (q UmeeQuery) HandleAggregateVotes(
	ctx sdk.Context,
	qs octypes.QueryServer,
) (proto.Message, error) {
	return qs.AggregateVotes(sdk.WrapSDKContext(ctx), &octypes.QueryAggregateVotes{})
}

// HandleOracleParams gets the x/oracle module's parameters.
func (q UmeeQuery) HandleOracleParams(
	ctx sdk.Context,
	qs octypes.QueryServer,
) (proto.Message, error) {
	return qs.Params(sdk.WrapSDKContext(ctx), &octypes.QueryParams{})
}

// HandleExchangeRates gets the exchange rates of all denoms.
func (q UmeeQuery) HandleExchangeRates(
	ctx sdk.Context,
	qs octypes.QueryServer,
) (proto.Message, error) {
	return qs.ExchangeRates(sdk.WrapSDKContext(ctx), &octypes.QueryExchangeRates{Denom: q.ExchangeRates.Denom})
}

// HandleActiveExchangeRates gets all active denoms.
func (q UmeeQuery) HandleActiveExchangeRates(
	ctx sdk.Context,
	qs octypes.QueryServer,
) (proto.Message, error) {
	return qs.ActiveExchangeRates(sdk.WrapSDKContext(ctx), &octypes.QueryActiveExchangeRates{})
}

// HandleMedians gets medians.
func (q UmeeQuery) HandleMedians(
	ctx sdk.Context,
	qs octypes.QueryServer,
) (proto.Message, error) {
	req := &octypes.QueryMedians{Denom: q.Medians.Denom, NumStamps: q.Medians.NumStamps}
	return qs.Medians(sdk.WrapSDKContext(ctx), req)
}

// HandleMedians gets median deviations.
func (q UmeeQuery) HandleMedianDeviations(
	ctx sdk.Context,
	qs octypes.QueryServer,
) (proto.Message, error) {
	req := &octypes.QueryMedianDeviations{Denom: q.MedianDeviations.Denom}
	return qs.MedianDeviations(sdk.WrapSDKContext(ctx), req)
}
