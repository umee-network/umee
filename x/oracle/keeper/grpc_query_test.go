package keeper_test

import (
	"context"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/x/oracle/types"
)

const (
	exchangeRate      = "umee"
	exchangeRateDenom = "uumee"
)

func (s *IntegrationTestSuite) TestQuerier_ActiveExchangeRates() {
	s.app.OracleKeeper.SetExchangeRate(s.ctx, exchangeRate, sdk.OneDec())
	_, err := s.queryClient.ActiveExchangeRates(context.Background(), &types.QueryActiveExchangeRatesRequest{})
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestQuerier_ExchangeRates() {
	s.app.OracleKeeper.SetExchangeRate(s.ctx, exchangeRate, sdk.OneDec())
	res, err := s.queryClient.ExchangeRates(context.Background(), &types.QueryExchangeRatesRequest{
		Denom: exchangeRateDenom,
	})
	s.Require().NoError(err)
	s.Require().Equal(sdk.DecCoins{
		sdk.NewDecCoinFromDec(exchangeRateDenom, sdk.OneDec()),
	}, res.ExchangeRates)
}

func (s *IntegrationTestSuite) TestQuerier_FeeederDelegation() {
	feederAddr := sdk.AccAddress([]byte("addr________________"))
	feederAcc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, feederAddr)
	s.app.AccountKeeper.SetAccount(s.ctx, feederAcc)

	err := s.app.OracleKeeper.ValidateFeeder(s.ctx, feederAddr, valAddr)
	s.Require().Error(err)

	s.app.OracleKeeper.SetFeederDelegation(s.ctx, valAddr, feederAddr)

	err = s.app.OracleKeeper.ValidateFeeder(s.ctx, feederAddr, valAddr)
	s.Require().NoError(err)

	_, err = s.queryClient.FeederDelegation(context.Background(), &types.QueryFeederDelegationRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestQuerier_MissCounter() {
	missCounter := uint64(rand.Intn(100))

	res, err := s.queryClient.MissCounter(context.Background(), &types.QueryMissCounterRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(res.MissCounter, uint64(0))

	s.app.OracleKeeper.SetMissCounter(s.ctx, valAddr, missCounter)

	res, err = s.queryClient.MissCounter(context.Background(), &types.QueryMissCounterRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(res.MissCounter, missCounter)
}

func (s *IntegrationTestSuite) TestQuerier_AggregatePrevote() {
	prevote := types.AggregateExchangeRatePrevote{
		Hash:        "hash",
		Voter:       addr.String(),
		SubmitBlock: 0,
	}
	s.app.OracleKeeper.SetAggregateExchangeRatePrevote(s.ctx, valAddr, prevote)

	_, err := s.app.OracleKeeper.GetAggregateExchangeRatePrevote(s.ctx, valAddr)
	s.Require().NoError(err)

	_, err = s.queryClient.AggregatePrevote(context.Background(), &types.QueryAggregatePrevoteRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestQuerier_AggregatePrevotes() {
	_, err := s.queryClient.AggregatePrevotes(context.Background(), &types.QueryAggregatePrevotesRequest{})
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestQuerier_AggregateVote() {
	var tuples types.ExchangeRateTuples
	tuples = append(tuples, types.ExchangeRateTuple{
		Denom:        "UMEE",
		ExchangeRate: sdk.ZeroDec(),
	})

	vote := types.AggregateExchangeRateVote{
		ExchangeRateTuples: tuples,
		Voter:              addr.String(),
	}
	s.app.OracleKeeper.SetAggregateExchangeRateVote(s.ctx, valAddr, vote)

	_, err := s.queryClient.AggregateVote(context.Background(), &types.QueryAggregateVoteRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestQuerier_AggregateVotes() {
	_, err := s.queryClient.AggregateVotes(context.Background(), &types.QueryAggregateVotesRequest{})
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestQuerier_Params() {
	_, err := s.queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)
}
