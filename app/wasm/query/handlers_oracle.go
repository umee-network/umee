package query

import (
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	octypes "github.com/umee-network/umee/v2/x/oracle/types"
)

// HandleExchangeRates gets the exchange rates of all denoms.
func (q UmeeQuery) HandleExchangeRates(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.ExchangeRates(sdk.WrapSDKContext(ctx), q.ExchangeRates)
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
	resp, err := qs.ActiveExchangeRates(sdk.WrapSDKContext(ctx), q.ActiveExchangeRates)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{
			Kind: fmt.Sprintf("error %+v to assigned query Active Exchange Rates", err),
		}
	}

	return MarshalResponse(resp)
}

// HandleFeederDelegation gets all the feeder delegation of a validator.
func (q UmeeQuery) HandleFeederDelegation(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.FeederDelegation(sdk.WrapSDKContext(ctx), q.FeederDelegation)
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
	resp, err := qs.MissCounter(sdk.WrapSDKContext(ctx), q.MissCounter)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Miss Counter", err)}
	}

	return MarshalResponse(resp)
}

// HandleAggregatePrevote gets an aggregate prevote of a validator.
func (q UmeeQuery) HandleAggregatePrevote(
	ctx sdk.Context,
	qs octypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.AggregatePrevote(sdk.WrapSDKContext(ctx), q.AggregatePrevote)
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
	resp, err := qs.AggregatePrevotes(sdk.WrapSDKContext(ctx), q.AggregatePrevotes)
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
	resp, err := qs.AggregateVote(sdk.WrapSDKContext(ctx), q.AggregateVote)
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
	resp, err := qs.AggregateVotes(sdk.WrapSDKContext(ctx), q.AggregateVotes)
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
	resp, err := qs.Params(sdk.WrapSDKContext(ctx), q.OracleParams)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Oracle Parameters", err)}
	}

	return MarshalResponse(resp)
}
