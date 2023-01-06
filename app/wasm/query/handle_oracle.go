package query

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	octypes "github.com/umee-network/umee/v4/x/oracle/types"
)

// HandleFeederDelegation gets all the feeder delegation of a validator.
func (q UmeeQuery) HandleFeederDelegation(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	req := &octypes.QueryFeederDelegation{ValidatorAddr: q.FeederDelegation.ValidatorAddr}
	resp, err := qs.FeederDelegation(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Feeder Delegation", err)}
	}

	return MarshalResponse(resp)
}

// HandleMissCounter gets all the oracle miss counter of a validator.
func (q UmeeQuery) HandleMissCounter(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	req := &octypes.QueryMissCounter{ValidatorAddr: q.MissCounter.ValidatorAddr}
	resp, err := qs.MissCounter(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Miss Counter", err)}
	}

	return MarshalResponse(resp)
}

// HandleSlashWindow gets slash window information.
func (q UmeeQuery) HandleSlashWindow(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.SlashWindow(sdk.WrapSDKContext(ctx), &octypes.QuerySlashWindow{})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Slash Window", err)}
	}

	return MarshalResponse(resp)
}

// HandleAggregatePrevote gets an aggregate prevote of a validator.
func (q UmeeQuery) HandleAggregatePrevote(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	req := &octypes.QueryAggregatePrevote{ValidatorAddr: q.AggregatePrevote.ValidatorAddr}
	resp, err := qs.AggregatePrevote(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Aggregate Prevote", err)}
	}

	return MarshalResponse(resp)
}

// HandleAggregatePrevotes gets an aggregate prevote of all validators.
func (q UmeeQuery) HandleAggregatePrevotes(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.AggregatePrevotes(sdk.WrapSDKContext(ctx), &octypes.QueryAggregatePrevotes{})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Aggregate Prevote", err)}
	}

	return MarshalResponse(resp)
}

// HandleAggregateVote gets an aggregate vote of a validator.
func (q UmeeQuery) HandleAggregateVote(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	req := &octypes.QueryAggregateVote{ValidatorAddr: q.AggregateVote.ValidatorAddr}
	resp, err := qs.AggregateVote(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Aggregate Vote", err)}
	}

	return MarshalResponse(resp)
}

// HandleAggregateVotes gets an aggregate vote of all validators.
func (q UmeeQuery) HandleAggregateVotes(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.AggregateVotes(sdk.WrapSDKContext(ctx), &octypes.QueryAggregateVotes{})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Aggregate Votes", err)}
	}

	return MarshalResponse(resp)
}

// HandleOracleParams gets the x/oracle module's parameters.
func (q UmeeQuery) HandleOracleParams(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.Params(sdk.WrapSDKContext(ctx), &octypes.QueryParams{})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Oracle Parameters", err)}
	}

	return MarshalResponse(resp)
}

// HandleExchangeRates gets the exchange rates of all denoms.
func (q UmeeQuery) HandleExchangeRates(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.ExchangeRates(sdk.WrapSDKContext(ctx), &octypes.QueryExchangeRates{Denom: q.ExchangeRates.Denom})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Exchange Rates", err)}
	}

	return MarshalResponse(resp)
}

// HandleActiveExchangeRates gets all active denoms.
func (q UmeeQuery) HandleActiveExchangeRates(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.ActiveExchangeRates(sdk.WrapSDKContext(ctx), &octypes.QueryActiveExchangeRates{})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{
			Kind: fmt.Sprintf("error %+v to assigned query Active Exchange Rates", err),
		}
	}

	return MarshalResponse(resp)
}

// HandleMedians gets medians.
func (q UmeeQuery) HandleMedians(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	req := &octypes.QueryMedians{Denom: q.Medians.Denom, NumStamps: q.Medians.NumStamps}
	resp, err := qs.Medians(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Medians", err)}
	}

	return MarshalResponse(resp)
}

// HandleMedians gets median deviations.
func (q UmeeQuery) HandleMedianDeviations(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	req := &octypes.QueryMedianDeviations{Denom: q.MedianDeviations.Denom}
	resp, err := qs.MedianDeviations(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Median Deviations", err)}
	}

	return MarshalResponse(resp)
}
