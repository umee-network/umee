package query

import (
	"encoding/json"
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	lvtypes "github.com/umee-network/umee/v2/x/leverage/types"
	octypes "github.com/umee-network/umee/v2/x/oracle/types"
)

// AssignedQuery defines the query to be called.
type AssignedQuery uint16

const (
	// AssignedQueryExchangeRates represents the call to query the exchange rates
	// of all denoms.
	AssignedQueryExchangeRates AssignedQuery = iota + 1
	// AssignedQueryRegisteredTokens represents the call of leverage get all registered tokens.
	AssignedQueryRegisteredTokens
	// AssignedQueryLeverageParams represents the call of the x/leverage module's parameters.
	AssignedQueryLeverageParams
	// AssignedQueryLiquidationTargets represents the call to query the list of
	// borrower addresses eligible for liquidation.
	AssignedQueryLiquidationTargets
	// AssignedQueryMarketSummary represents the call to query the market
	// summary data of an denom.
	AssignedQueryMarketSummary
	// AssignedQueryActiveExchangeRates represents the call to query all active denoms.
	AssignedQueryActiveExchangeRates
	// AssignedQueryActiveFeederDelegation represents the call to query all the feeder
	// delegation of a validator.
	AssignedQueryActiveFeederDelegation
	// AssignedQueryMissCounter represents the call to query all the oracle
	// miss counter of a validator.
	AssignedQueryMissCounter
	// AssignedQueryAggregatePrevote represents the call to query an aggregate prevote of
	// a validator.
	AssignedQueryAggregatePrevote
	// AssignedQueryAggregatePrevotes represents the call to query an aggregate prevote of
	// all validators.
	AssignedQueryAggregatePrevotes
	// AssignedQueryAggregateVote represents the call to query an aggregate vote of
	// a validator.
	AssignedQueryAggregateVote
	// AssignedQueryAggregateVotes represents the call to query an aggregate vote of
	// all validators.
	AssignedQueryAggregateVotes
	// AssignedQueryOracleParams represents the call of the x/leverage module's
	// parameters.
	AssignedQueryOracleParams
)

// MarshalResponse marshals any response.
func MarshalResponse(resp interface{}) ([]byte, error) {
	bz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v umee query response error on marshal", err)}
	}
	return bz, err
}

// UmeeQuery wraps all the queries availables for cosmwasm smartcontracts.
type UmeeQuery struct {
	// Mandatory field to determine which query to call.
	AssignedQuery AssignedQuery `json:"assigned_query"`
	// Used to get the exchange rates of all denoms.
	ExchangeRates *octypes.QueryExchangeRates `json:"exchange_rates,omitempty"`
	// Used to query all the registered tokens.
	RegisteredTokens *lvtypes.QueryRegisteredTokens `json:"registered_tokens,omitempty"`
	// Used to query the x/leverage module's parameters.
	LeverageParams *lvtypes.QueryParams `json:"leverage_params,omitempty"`
	// request to return a list of borrower addresses eligible for liquidation.
	LiquidationTargets *lvtypes.QueryLiquidationTargets `json:"liquidation_targets,omitempty"`
	// Used to get the summary data of an denom.
	MarketSummary *lvtypes.QueryMarketSummary `json:"market_summary,omitempty"`
	// Used to get all active denoms.
	ActiveExchangeRates *octypes.QueryActiveExchangeRates `json:"active_exchange_rates,omitempty"`
	// Used to get all feeder delegation of a validator.
	FeederDelegation *octypes.QueryFeederDelegation `json:"feeder_delegation,omitempty"`
	// Used to get all the oracle miss counter of a validator.
	MissCounter *octypes.QueryMissCounter `json:"miss_counter,omitempty"`
	// Used to get an aggregate prevote of a validator.
	AggregatePrevote *octypes.QueryAggregatePrevote `json:"aggregate_prevote,omitempty"`
	// Used to get an aggregate prevote of all validators.
	AggregatePrevotes *octypes.QueryAggregatePrevotes `json:"aggregate_prevotes,omitempty"`
	// Used to get an aggregate vote of a validator.
	AggregateVote *octypes.QueryAggregateVote `json:"aggregate_vote,omitempty"`
	// Used to get an aggregate vote of all validators.
	AggregateVotes *octypes.QueryAggregateVotes `json:"aggregate_votes,omitempty"`
	// Used to query the x/oracle module's parameters.
	OracleParams *octypes.QueryParams `json:"oracle_params,omitempty"`
}
