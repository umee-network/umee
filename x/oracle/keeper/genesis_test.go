package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/oracle/types"
	"gotest.tools/v3/assert"
)

func (s *IntegrationTestSuite) TestIterateAllHistoricPrices() {
	keeper, ctx := s.app.OracleKeeper, s.ctx

	historicPrices := []types.Price{
		{BlockNum: 10, ExchangeRateTuple: types.ExchangeRateTuple{
			Denom: "umee", ExchangeRate: sdk.MustNewDecFromStr("20.45"),
		}},
		{BlockNum: 11, ExchangeRateTuple: types.ExchangeRateTuple{
			Denom: "umee", ExchangeRate: sdk.MustNewDecFromStr("20.44"),
		}},
		{BlockNum: 10, ExchangeRateTuple: types.ExchangeRateTuple{
			Denom: "btc", ExchangeRate: sdk.MustNewDecFromStr("1200.56"),
		}},
		{BlockNum: 11, ExchangeRateTuple: types.ExchangeRateTuple{
			Denom: "btc", ExchangeRate: sdk.MustNewDecFromStr("1200.19"),
		}},
	}

	for _, hp := range historicPrices {
		keeper.SetHistoricPrice(ctx, hp.ExchangeRateTuple.Denom, hp.BlockNum, hp.ExchangeRateTuple.ExchangeRate)
	}

	newPrices := []types.Price{}
	keeper.IterateAllHistoricPrices(
		ctx,
		func(historicPrice types.Price) bool {
			newPrices = append(newPrices, historicPrice)
			return false
		},
	)

	s.Require().Equal(len(historicPrices), len(newPrices))

	// Verify that the historic prices from IterateAllHistoricPrices equal
	// the ones set by SetHistoricPrice
FOUND:
	for _, oldPrice := range historicPrices {
		for _, newPrice := range newPrices {
			if oldPrice.BlockNum == newPrice.BlockNum && oldPrice.ExchangeRateTuple.Denom == newPrice.ExchangeRateTuple.Denom {
				s.Require().Equal(oldPrice.ExchangeRateTuple.ExchangeRate, newPrice.ExchangeRateTuple.ExchangeRate)
				continue FOUND
			}
		}
		s.T().Errorf("did not find match for historic price: %+v", oldPrice)
	}
}

func (s *IntegrationTestSuite) TestIterateAllMedianPrices() {
	keeper, ctx := s.app.OracleKeeper, s.ctx
	medians := []types.ExchangeRateTuple{
		{Denom: "umee", ExchangeRate: sdk.MustNewDecFromStr("20.44")},
		{Denom: "atom", ExchangeRate: sdk.MustNewDecFromStr("2.66")},
		{Denom: "osmo", ExchangeRate: sdk.MustNewDecFromStr("13.64")},
	}

	for _, m := range medians {
		keeper.SetHistoricMedian(ctx, m.Denom, uint64(s.ctx.BlockHeight()), m.ExchangeRate)
	}

	newMedians := []types.Price{}
	keeper.IterateAllMedianPrices(
		ctx,
		func(median types.Price) bool {
			newMedians = append(newMedians, median)
			return false
		},
	)
	assert.Equal(s.T(), len(medians), len(newMedians))

FOUND:
	for _, oldMedian := range medians {
		for _, newMedian := range newMedians {
			if oldMedian.Denom == newMedian.ExchangeRateTuple.Denom {
				s.Require().Equal(oldMedian.ExchangeRate, newMedian.ExchangeRateTuple.ExchangeRate)
				continue FOUND
			}
		}
		s.T().Errorf("did not find match for median price: %+v", oldMedian)
	}
}

func (s *IntegrationTestSuite) TestIterateAllMedianDeviationPrices() {
	keeper, ctx := s.app.OracleKeeper, s.ctx
	medians := []types.ExchangeRateTuple{
		{Denom: "umee", ExchangeRate: sdk.MustNewDecFromStr("21.44")},
		{Denom: "atom", ExchangeRate: sdk.MustNewDecFromStr("3.66")},
		{Denom: "osmo", ExchangeRate: sdk.MustNewDecFromStr("14.64")},
	}

	for _, m := range medians {
		keeper.SetHistoricMedianDeviation(ctx, m.Denom, uint64(s.ctx.BlockHeight()), m.ExchangeRate)
	}

	newMedians := []types.Price{}
	keeper.IterateAllMedianDeviationPrices(
		ctx,
		func(median types.Price) bool {
			newMedians = append(newMedians, median)
			return false
		},
	)
	assert.Equal(s.T(), len(medians), len(newMedians))

FOUND:
	for _, oldMedian := range medians {
		for _, newMedian := range newMedians {
			if oldMedian.Denom == newMedian.ExchangeRateTuple.Denom {
				s.Require().Equal(oldMedian.ExchangeRate, newMedian.ExchangeRateTuple.ExchangeRate)
				continue FOUND
			}
		}
		s.T().Errorf("did not find match for median price: %+v", oldMedian)
	}
}
