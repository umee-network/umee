package keeper_test

import (
	"context"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/x/oracle/types"
)

// ActiveExchangeRates
func (s *IntegrationTestSuite) TestQuerier_ActiveExchangeRates() {
	const (
		exchangeRate = "umee"
	)
	app, ctx := s.app, s.ctx

	app.OracleKeeper.SetExchangeRate(ctx, exchangeRate, sdk.OneDec())
	_, err := s.queryClient.ActiveExchangeRates(context.Background(), &types.QueryActiveExchangeRatesRequest{})
	s.Require().NoError(err)
}

// ExchangeRates
func (s *IntegrationTestSuite) TestQuerier_ExchangeRates() {
	const (
		exchangeRate      = "umee"
		exchangeRateDenom = "uumee"
	)
	app, ctx := s.app, s.ctx

	app.OracleKeeper.SetExchangeRate(ctx, exchangeRate, sdk.OneDec())
	res, err := s.queryClient.ExchangeRates(context.Background(), &types.QueryExchangeRatesRequest{
		Denom: exchangeRateDenom,
	})
	s.Require().NoError(err)
	s.Require().Equal(sdk.DecCoins{
		sdk.NewDecCoinFromDec(exchangeRateDenom, sdk.OneDec()),
	}, res.ExchangeRates)
}

// FeederDelegation
func (s *IntegrationTestSuite) TestQuerier_FeeederDelegation() {
	app, ctx := s.app, s.ctx

	feederAddr := sdk.AccAddress([]byte("addr________________"))
	feederAcc := app.AccountKeeper.NewAccountWithAddress(ctx, feederAddr)
	app.AccountKeeper.SetAccount(ctx, feederAcc)

	err := s.app.OracleKeeper.ValidateFeeder(ctx, feederAddr, valAddr)
	s.Require().Error(err)

	s.app.OracleKeeper.SetFeederDelegation(ctx, valAddr, feederAddr)

	err = s.app.OracleKeeper.ValidateFeeder(ctx, feederAddr, valAddr)
	s.Require().NoError(err)

	_, err = s.queryClient.FeederDelegation(context.Background(), &types.QueryFeederDelegationRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
}

// MissCounter
func (s *IntegrationTestSuite) TestQuerier_MissCounter() {
	app, ctx := s.app, s.ctx
	missCounter := uint64(rand.Intn(100))

	res, err := s.queryClient.MissCounter(context.Background(), &types.QueryMissCounterRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(res.MissCounter, uint64(0))

	app.OracleKeeper.SetMissCounter(ctx, valAddr, missCounter)

	res, err = s.queryClient.MissCounter(context.Background(), &types.QueryMissCounterRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(res.MissCounter, missCounter)
}

// AggregatePrevote
func (s *IntegrationTestSuite) TestQuerier_AggregatePrevote() {
	app, ctx := s.app, s.ctx

	prevote := types.AggregateExchangeRatePrevote{
		Hash:        "hash",
		Voter:       addr.String(),
		SubmitBlock: 0,
	}
	app.OracleKeeper.SetAggregateExchangeRatePrevote(ctx, valAddr, prevote)

	_, err := app.OracleKeeper.GetAggregateExchangeRatePrevote(ctx, valAddr)
	s.Require().NoError(err)

	_, err = s.queryClient.AggregatePrevote(context.Background(), &types.QueryAggregatePrevoteRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
}

// AggregatePrevotes
func (s *IntegrationTestSuite) TestQuerier_AggregatePrevotes() {
	_, err := s.queryClient.AggregatePrevotes(context.Background(), &types.QueryAggregatePrevotesRequest{})
	s.Require().NoError(err)
}

// AggregateVote
func (s *IntegrationTestSuite) TestQuerier_AggregateVote() {
	app, ctx := s.app, s.ctx

	var tuples types.ExchangeRateTuples
	tuples = append(tuples, types.ExchangeRateTuple{
		Denom:        "UMEE",
		ExchangeRate: sdk.ZeroDec(),
	})

	vote := types.AggregateExchangeRateVote{
		ExchangeRateTuples: tuples,
		Voter:              addr.String(),
	}
	app.OracleKeeper.SetAggregateExchangeRateVote(ctx, valAddr, vote)

	_, err := s.queryClient.AggregateVote(context.Background(), &types.QueryAggregateVoteRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
}

// AggregateVotes
func (s *IntegrationTestSuite) TestQuerier_AggregateVotes() {
	_, err := s.queryClient.AggregateVotes(context.Background(), &types.QueryAggregateVotesRequest{})
	s.Require().NoError(err)
}

// Params
func (s *IntegrationTestSuite) TestQuerier_Params() {
	_, err := s.queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)
}
