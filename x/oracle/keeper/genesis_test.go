package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/oracle/types"
)

func (s *IntegrationTestSuite) TestIterateAllHistoricPrices() {
	keeper, ctx := s.app.OracleKeeper, s.ctx

	historicPrices := []types.HistoricPrice{
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

	newPrices := []types.HistoricPrice{}
	keeper.IterateAllHistoricPrices(
		ctx,
		func(historicPrice types.HistoricPrice) bool {
			newPrices = append(newPrices, historicPrice)
			return false
		},
	)

	fmt.Printf("%+v\n", newPrices)

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

// func (s *IntegrationTestSuite) TestIterateAllMedianPrices() {
// 	keeper, ctx := s.app.OracleKeeper, s.ctx

// 	medians := []sdk.DecCoin{
// 		{Denom: "umee", Amount: sdk.MustNewDecFromStr("35.19")},
// 		{Denom: "atom", Amount: sdk.MustNewDecFromStr("31.22")},
// 	}

// 	for _, median := range medians {
// 		keeper.SetMedian(ctx, median.Denom, median.Amount)
// 	}

// 	newMedians := []sdk.DecCoin{}
// 	keeper.IterateAllMedianPrices(
// 		ctx,
// 		func(median sdk.DecCoin) bool {
// 			newMedians = append(newMedians, median)
// 			return false
// 		},
// 	)

// 	s.Require().Equal(len(medians), len(newMedians))
// }
