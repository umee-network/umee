package keeper_test

import (
	"context"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/x/oracle/keeper"
	"github.com/umee-network/umee/v2/x/oracle/types"
)

func (s *IntegrationTestSuite) TestQuerier_ActiveExchangeRates() {
	s.app.OracleKeeper.SetExchangeRate(s.ctx, exchangeRate, sdk.OneDec())
	res, err := s.queryClient.ActiveExchangeRates(context.Background(), &types.QueryActiveExchangeRatesRequest{})
	s.Require().NoError(err)
	s.Require().Equal([]string{exchangeRate}, res.ActiveRates)
}

func (s *IntegrationTestSuite) TestQuerier_ExchangeRates() {
	s.app.OracleKeeper.SetExchangeRate(s.ctx, exchangeRate, sdk.OneDec())
	res, err := s.queryClient.ExchangeRates(context.Background(), &types.QueryExchangeRatesRequest{})
	s.Require().NoError(err)
	s.Require().Equal(sdk.DecCoins{
		sdk.NewDecCoinFromDec(exchangeRate, sdk.OneDec()),
	}, res.ExchangeRates)

	res, err = s.queryClient.ExchangeRates(context.Background(), &types.QueryExchangeRatesRequest{
		Denom: exchangeRate,
	})
	s.Require().NoError(err)
	s.Require().Equal(sdk.DecCoins{
		sdk.NewDecCoinFromDec(exchangeRate, sdk.OneDec()),
	}, res.ExchangeRates)
}

func (s *IntegrationTestSuite) TestQuerier_FeeederDelegation() {
	feederAddr := sdk.AccAddress([]byte("addr________________"))
	feederAcc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, feederAddr)
	inactiveValidator := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	s.app.AccountKeeper.SetAccount(s.ctx, feederAcc)

	err := s.app.OracleKeeper.ValidateFeeder(s.ctx, feederAddr, valAddr)
	s.Require().Error(err)

	_, err = s.queryClient.FeederDelegation(context.Background(), &types.QueryFeederDelegationRequest{
		ValidatorAddr: inactiveValidator,
	})
	s.Require().Error(err)

	s.app.OracleKeeper.SetFeederDelegation(s.ctx, valAddr, feederAddr)

	err = s.app.OracleKeeper.ValidateFeeder(s.ctx, feederAddr, valAddr)
	s.Require().NoError(err)

	res, err := s.queryClient.FeederDelegation(context.Background(), &types.QueryFeederDelegationRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(feederAddr.String(), res.FeederAddr)
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

	res, err := s.app.OracleKeeper.GetAggregateExchangeRatePrevote(s.ctx, valAddr)
	s.Require().NoError(err)
	s.Require().Equal(prevote, res)

	queryRes, err := s.queryClient.AggregatePrevote(context.Background(), &types.QueryAggregatePrevoteRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(types.AggregateExchangeRatePrevote{
		Hash:        "hash",
		Voter:       addr.String(),
		SubmitBlock: 0,
	}, queryRes.AggregatePrevote)
}

func (s *IntegrationTestSuite) TestQuerier_AggregatePrevotes() {
	res, err := s.queryClient.AggregatePrevotes(context.Background(), &types.QueryAggregatePrevotesRequest{})
	s.Require().Equal([]types.AggregateExchangeRatePrevote(nil), res.AggregatePrevotes)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestQuerier_AggregateVote() {
	var tuples types.ExchangeRateTuples
	tuples = append(tuples, types.ExchangeRateTuple{
		Denom:        umeeapp.DisplayDenom,
		ExchangeRate: sdk.ZeroDec(),
	})

	vote := types.AggregateExchangeRateVote{
		ExchangeRateTuples: tuples,
		Voter:              addr.String(),
	}
	s.app.OracleKeeper.SetAggregateExchangeRateVote(s.ctx, valAddr, vote)

	res, err := s.queryClient.AggregateVote(context.Background(), &types.QueryAggregateVoteRequest{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(types.AggregateExchangeRateVote{
		ExchangeRateTuples: tuples,
		Voter:              addr.String(),
	}, res.AggregateVote)
}

func (s *IntegrationTestSuite) TestQuerier_AggregateVotes() {
	res, err := s.queryClient.AggregateVotes(context.Background(), &types.QueryAggregateVotesRequest{})
	s.Require().NoError(err)
	s.Require().Equal([]types.AggregateExchangeRateVote(nil), res.AggregateVotes)
}

func (s *IntegrationTestSuite) TestQuerier_Params() {
	res, err := s.queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(types.DefaultGenesisState().Params, res.Params)
}

func TestEmptyRequest(t *testing.T) {
	q := keeper.NewQuerier(keeper.Keeper{})
	emptyRequestErrorMsg := "empty request"

	resParams, err := q.Params(context.Background(), nil)
	require.Nil(t, resParams)
	require.ErrorContains(t, err, emptyRequestErrorMsg)

	resExchangeRate, err := q.ExchangeRates(context.Background(), nil)
	require.Nil(t, resExchangeRate)
	require.ErrorContains(t, err, emptyRequestErrorMsg)

	resActiveExchangeRates, err := q.ActiveExchangeRates(context.Background(), nil)
	require.Nil(t, resActiveExchangeRates)
	require.ErrorContains(t, err, emptyRequestErrorMsg)

	resFeederDelegation, err := q.FeederDelegation(context.Background(), nil)
	require.Nil(t, resFeederDelegation)
	require.ErrorContains(t, err, emptyRequestErrorMsg)

	resMissCounter, err := q.MissCounter(context.Background(), nil)
	require.Nil(t, resMissCounter)
	require.ErrorContains(t, err, emptyRequestErrorMsg)

	resAggregatePrevote, err := q.AggregatePrevote(context.Background(), nil)
	require.Nil(t, resAggregatePrevote)
	require.ErrorContains(t, err, emptyRequestErrorMsg)

	resAggregateVote, err := q.AggregateVote(context.Background(), nil)
	require.Nil(t, resAggregateVote)
	require.ErrorContains(t, err, emptyRequestErrorMsg)

	resAggregateVotes, err := q.AggregateVotes(context.Background(), nil)
	require.Nil(t, resAggregateVotes)
	require.ErrorContains(t, err, emptyRequestErrorMsg)
}
