package oracle_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v5/x/oracle"
	"github.com/umee-network/umee/v5/x/oracle/types"
	"gotest.tools/v3/assert"
)

const (
	umeeAddr        = "umee1s84d29zk3k20xk9f0hvczkax90l9t94g72n6wm"
	umeevaloperAddr = "umeevaloper1kqh6nt4f48vptvq4j5cgr0nfd2x4z9ulvrtqrh"
	denom           = "umee"
	upperDenom      = "UMEE"
)

var exchangeRate = sdk.MustNewDecFromStr("1.2")

func (s *IntegrationTestSuite) TestGenesis_InitGenesis() {
	keeper, ctx := s.app.OracleKeeper, s.ctx

	tcs := []struct {
		name      string
		g         types.GenesisState
		expectErr bool
		errMsg    string
	}{
		{
			"FeederDelegations.FeederAddress: empty address",
			types.GenesisState{
				FeederDelegations: []types.FeederDelegation{
					{
						FeederAddress: "",
					},
				},
			},
			true,
			"empty address string is not allowed",
		},
		{
			"FeederDelegations.ValidatorAddress: empty address",
			types.GenesisState{
				FeederDelegations: []types.FeederDelegation{
					{
						FeederAddress:    umeeAddr,
						ValidatorAddress: "",
					},
				},
			},
			true,
			"empty address string is not allowed",
		},
		{
			"valid",
			types.GenesisState{
				Params: types.DefaultParams(),
				ExchangeRates: types.ExchangeRateTuples{
					types.ExchangeRateTuple{
						Denom:        denom,
						ExchangeRate: exchangeRate,
					},
				},
				HistoricPrices: types.Prices{
					types.Price{
						ExchangeRateTuple: types.ExchangeRateTuple{
							Denom:        denom,
							ExchangeRate: exchangeRate,
						},
						BlockNum: 0,
					},
				},
				Medians: types.Prices{
					types.Price{
						ExchangeRateTuple: types.ExchangeRateTuple{
							Denom:        denom,
							ExchangeRate: exchangeRate,
						},
						BlockNum: 0,
					},
				},
				MedianDeviations: types.Prices{
					types.Price{
						ExchangeRateTuple: types.ExchangeRateTuple{
							Denom:        denom,
							ExchangeRate: exchangeRate,
						},
						BlockNum: 0,
					},
				},
			},
			false,
			"",
		},
		{
			"FeederDelegations.ValidatorAddress: empty address",
			types.GenesisState{
				MissCounters: []types.MissCounter{
					{
						ValidatorAddress: "",
					},
				},
			},
			true,
			"empty address string is not allowed",
		},
		{
			"AggregateExchangeRatePrevotes.Voter: empty address",
			types.GenesisState{
				AggregateExchangeRatePrevotes: []types.AggregateExchangeRatePrevote{
					{
						Voter: "",
					},
				},
			},
			true,
			"empty address string is not allowed",
		},
		{
			"AggregateExchangeRateVotes.Voter: empty address",
			types.GenesisState{
				AggregateExchangeRateVotes: []types.AggregateExchangeRateVote{
					{
						Voter: "",
					},
				},
			},
			true,
			"empty address string is not allowed",
		},
	}

	for _, tc := range tcs {
		s.Run(
			tc.name, func() {
				if tc.expectErr {
					s.Assertions.PanicsWithError(tc.errMsg, func() { oracle.InitGenesis(ctx, keeper, tc.g) })
				} else {
					s.Assertions.NotPanics(func() { oracle.InitGenesis(ctx, keeper, tc.g) })
				}
			},
		)
	}
}

func (s *IntegrationTestSuite) TestGenesis_ExportGenesis() {
	keeper, ctx := s.app.OracleKeeper, s.ctx
	params := types.DefaultParams()

	feederDelegations := []types.FeederDelegation{
		{
			FeederAddress:    umeeAddr,
			ValidatorAddress: umeevaloperAddr,
		},
	}
	exchangeRateTuples := types.ExchangeRateTuples{
		types.ExchangeRateTuple{
			Denom:        upperDenom,
			ExchangeRate: exchangeRate,
		},
	}
	missCounters := []types.MissCounter{
		{
			ValidatorAddress: umeevaloperAddr,
		},
	}
	aggregateExchangeRatePrevotes := []types.AggregateExchangeRatePrevote{
		{
			Voter: umeevaloperAddr,
		},
	}
	aggregateExchangeRateVotes := []types.AggregateExchangeRateVote{
		{
			Voter: umeevaloperAddr,
		},
	}
	historicPrices := []types.Price{
		{
			ExchangeRateTuple: types.ExchangeRateTuple{
				Denom:        denom,
				ExchangeRate: exchangeRate,
			},
			BlockNum: 0,
		},
	}
	medians := []types.Price{
		{
			ExchangeRateTuple: types.ExchangeRateTuple{
				Denom:        denom,
				ExchangeRate: exchangeRate,
			},
			BlockNum: 0,
		},
	}
	medianDeviations := []types.Price{
		{
			ExchangeRateTuple: types.ExchangeRateTuple{
				Denom:        denom,
				ExchangeRate: exchangeRate,
			},
			BlockNum: 0,
		},
	}

	genesisState := types.GenesisState{
		Params:                        params,
		FeederDelegations:             feederDelegations,
		ExchangeRates:                 exchangeRateTuples,
		MissCounters:                  missCounters,
		AggregateExchangeRatePrevotes: aggregateExchangeRatePrevotes,
		AggregateExchangeRateVotes:    aggregateExchangeRateVotes,
		Medians:                       medians,
		HistoricPrices:                historicPrices,
		MedianDeviations:              medianDeviations,
	}

	oracle.InitGenesis(ctx, keeper, genesisState)

	result := oracle.ExportGenesis(s.ctx, s.app.OracleKeeper)
	assert.DeepEqual(s.T(), params, result.Params)
	assert.DeepEqual(s.T(), feederDelegations, result.FeederDelegations)
	assert.DeepEqual(s.T(), exchangeRateTuples, result.ExchangeRates)
	assert.DeepEqual(s.T(), missCounters, result.MissCounters)
	assert.DeepEqual(s.T(), aggregateExchangeRatePrevotes, result.AggregateExchangeRatePrevotes)
	assert.DeepEqual(s.T(), aggregateExchangeRateVotes, result.AggregateExchangeRateVotes)
	assert.DeepEqual(s.T(), medians, result.Medians)
	assert.DeepEqual(s.T(), historicPrices, result.HistoricPrices)
	assert.DeepEqual(s.T(), medianDeviations, result.MedianDeviations)
}
