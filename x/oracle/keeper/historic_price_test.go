package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v3/x/oracle/types"
)

func (s *IntegrationTestSuite) TestSetHistoraclePricing() {
	app, ctx := s.app, s.ctx

	// set exchange rate in store before adding a historic price
	app.OracleKeeper.SetExchangeRate(ctx, displayDenom, sdk.OneDec())
	rate, err := app.OracleKeeper.GetExchangeRate(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.OneDec())

	app.OracleKeeper.AddHistoricPrice(ctx, displayDenom, rate)
	historicPrice, err := app.OracleKeeper.GetHistoricPrice(ctx, displayDenom, uint64(ctx.BlockHeight()))
	s.Require().NoError(err)
	s.Require().Equal(historicPrice, types.HistoricPrice{
		ExchangeRates: types.ExchangeRateTuple{
			Denom:        displayDenom,
			ExchangeRate: sdk.OneDec(),
		},
		BlockNum: uint64(ctx.BlockHeight()),
	})

	// add multiple historic prices to store
	exchangeRates := []string{"1.2", "1.1", "1.4"}
	for _, exchangeRate := range exchangeRates {
		// update blockheight
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

		// update exchange rate in store before updating historic price
		newRate := sdk.OneDec().Mul(sdk.MustNewDecFromStr(exchangeRate))
		app.OracleKeeper.SetExchangeRate(ctx, displayDenom, newRate)
		rate, err = app.OracleKeeper.GetExchangeRate(ctx, displayDenom)
		s.Require().NoError(err)
		s.Require().Equal(rate, newRate)

		app.OracleKeeper.AddHistoricPrice(ctx, displayDenom, rate)
		historicPrice, err = app.OracleKeeper.GetHistoricPrice(ctx, displayDenom, uint64(ctx.BlockHeight()))
		s.Require().NoError(err)
		s.Require().Equal(historicPrice, types.HistoricPrice{
			ExchangeRates: types.ExchangeRateTuple{
				Denom:        displayDenom,
				ExchangeRate: newRate,
			},
			BlockNum: uint64(ctx.BlockHeight()),
		})
	}

	// check all historic prices were set
	historicPrices := app.OracleKeeper.GetHistoricPrices(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(len(historicPrices), 4)

	// check median was set
	median, err := app.OracleKeeper.GetMedian(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(median, sdk.MustNewDecFromStr("1.15"))

	// delete first historic price and check median was updated
	app.OracleKeeper.DeleteHistoricPrice(ctx, displayDenom, uint64(ctx.BlockHeight()-3))
	s.Require().NoError(err)

	median, err = app.OracleKeeper.GetMedian(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(median, sdk.MustNewDecFromStr("1.2"))

	historicPrices = app.OracleKeeper.GetHistoricPrices(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(len(historicPrices), 3)

	// check historic prices and medians get cleared
	app.OracleKeeper.ClearHistoricPrices(ctx)
	s.Require().NoError(err)

	historicPrices = app.OracleKeeper.GetHistoricPrices(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(len(historicPrices), 0)

	median, err = app.OracleKeeper.GetMedian(ctx, displayDenom)
	s.Require().EqualError(err, sdkerrors.Wrap(types.ErrUnknownDenom, displayDenom).Error())
	s.Require().Equal(median, sdk.ZeroDec())
}
