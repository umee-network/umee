package keeper

import (
	"context"

	"github.com/umee-network/umee/x/oracle/types"
)

var _ types.QueryServer = Querier{}

// Querier implements a QueryServer for the x/oracle module.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

func (q Querier) Actives(goCtx context.Context, req *types.QueryActivesRequest) (*types.QueryActivesResponse, error) {
	return nil, nil
}

func (q Querier) AggregatePrevote(goCtx context.Context,
	req *types.QueryAggregatePrevoteRequest) (*types.QueryAggregatePrevoteResponse,
	error) {
	return nil, nil
}

func (q Querier) AggregatePrevotes(goCtx context.Context,
	req *types.QueryAggregatePrevotesRequest) (*types.QueryAggregatePrevotesResponse,
	error) {
	return nil, nil
}

func (q Querier) AggregateVote(goCtx context.Context,
	req *types.QueryAggregateVoteRequest) (*types.QueryAggregateVoteResponse,
	error) {
	return nil, nil
}

func (q Querier) AggregateVotes(goCtx context.Context,
	req *types.QueryAggregateVotesRequest) (*types.QueryAggregateVotesResponse,
	error) {
	return nil, nil
}

func (q Querier) ExchangeRate(goCtx context.Context,
	req *types.QueryExchangeRateRequest) (*types.QueryExchangeRateResponse,
	error) {
	return nil, nil
}
func (q Querier) ExchangeRates(goCtx context.Context,
	req *types.QueryExchangeRatesRequest) (*types.QueryExchangeRatesResponse,
	error) {
	return nil, nil
}
func (q Querier) FeederDelegation(goCtx context.Context,
	req *types.QueryFeederDelegationRequest) (*types.QueryFeederDelegationResponse,
	error) {
	return nil, nil
}
func (q Querier) MissCounter(goCtx context.Context,
	req *types.QueryMissCounterRequest) (*types.QueryMissCounterResponse,
	error) {
	return nil, nil
}
func (q Querier) Params(goCtx context.Context,
	req *types.QueryParamsRequest) (*types.QueryParamsResponse,
	error) {
	return nil, nil
}
func (q Querier) TobinTax(goCtx context.Context,
	req *types.QueryTobinTaxRequest) (*types.QueryTobinTaxResponse,
	error) {
	return nil, nil
}

func (q Querier) TobinTaxes(goCtx context.Context,
	req *types.QueryTobinTaxesRequest) (*types.QueryTobinTaxesResponse,
	error) {
	return nil, nil
}
func (q Querier) VoteTargets(goCtx context.Context,
	req *types.QueryVoteTargetsRequest) (*types.QueryVoteTargetsResponse,
	error) {
	return nil, nil
}
