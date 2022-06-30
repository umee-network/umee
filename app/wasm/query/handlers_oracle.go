package query

import (
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	octypes "github.com/umee-network/umee/v2/x/oracle/types"
)

// HandleExchangeRates gets the exchange rates of all denoms.
func (umeeQuery UmeeQuery) HandleExchangeRates(
	ctx sdk.Context,
	queryServer octypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.ExchangeRates(sdk.WrapSDKContext(ctx), umeeQuery.ExchangeRates)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Exchange Rates", err)}
	}

	return MarshalResponse(resp)
}

// HandleActiveExchangeRates gets all active denoms.
func (umeeQuery UmeeQuery) HandleActiveExchangeRates(
	ctx sdk.Context,
	queryServer octypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.ActiveExchangeRates(sdk.WrapSDKContext(ctx), umeeQuery.ActiveExchangeRates)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{
			Kind: fmt.Sprintf("error %+v to assigned query Active Exchange Rates", err),
		}
	}

	return MarshalResponse(resp)
}

// HandleFeederDelegation gets all the feeder delegation of a validator.
func (umeeQuery UmeeQuery) HandleFeederDelegation(
	ctx sdk.Context,
	queryServer octypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.FeederDelegation(sdk.WrapSDKContext(ctx), umeeQuery.FeederDelegation)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Feeder Delegation", err)}
	}

	return MarshalResponse(resp)
}

// HandleMissCounter gets all the oracle miss counter of a validator.
func (umeeQuery UmeeQuery) HandleMissCounter(
	ctx sdk.Context,
	queryServer octypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.MissCounter(sdk.WrapSDKContext(ctx), umeeQuery.MissCounter)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Miss Counter", err)}
	}

	return MarshalResponse(resp)
}

// HandleAggregatePrevote gets an aggregate prevote of a validator.
func (umeeQuery UmeeQuery) HandleAggregatePrevote(
	ctx sdk.Context,
	queryServer octypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.AggregatePrevote(sdk.WrapSDKContext(ctx), umeeQuery.AggregatePrevote)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Aggregate Prevote", err)}
	}

	return MarshalResponse(resp)
}

// HandleAggregatePrevotes gets an aggregate prevote of all validators.
func (umeeQuery UmeeQuery) HandleAggregatePrevotes(
	ctx sdk.Context,
	queryServer octypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.AggregatePrevotes(sdk.WrapSDKContext(ctx), umeeQuery.AggregatePrevotes)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Aggregate Prevote", err)}
	}

	return MarshalResponse(resp)
}

// HandleAggregateVote gets an aggregate vote of a validator.
func (umeeQuery UmeeQuery) HandleAggregateVote(
	ctx sdk.Context,
	queryServer octypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.AggregateVote(sdk.WrapSDKContext(ctx), umeeQuery.AggregateVote)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Aggregate Vote", err)}
	}

	return MarshalResponse(resp)
}

// HandleAggregateVotes gets an aggregate vote of all validators.
func (umeeQuery UmeeQuery) HandleAggregateVotes(
	ctx sdk.Context,
	queryServer octypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.AggregateVotes(sdk.WrapSDKContext(ctx), umeeQuery.AggregateVotes)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Aggregate Votes", err)}
	}

	return MarshalResponse(resp)
}
