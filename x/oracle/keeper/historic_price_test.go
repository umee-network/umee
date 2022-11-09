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
		ExchangeRate: sdk.OneDec(),
		BlockNum:     uint64(ctx.BlockHeight()),
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
			ExchangeRate: newRate,
			BlockNum:     uint64(ctx.BlockHeight()),
		})
	}

	// check all historic prices were set
	historicPrices := app.OracleKeeper.GetHistoricPrices(ctx, displayDenom)
	s.Require().Equal(len(historicPrices), 4)

	// set and check median and median standard deviation
	app.OracleKeeper.SetMedian(ctx, displayDenom)
	median, err := app.OracleKeeper.GetMedian(ctx, displayDenom, uint64(ctx.BlockHeight()))
	s.Require().NoError(err)
	s.Require().Equal(median, sdk.MustNewDecFromStr("1.15"))

	medianDeviation, err := app.OracleKeeper.GetMedianDeviation(ctx, displayDenom, uint64(ctx.BlockHeight()))
	s.Require().NoError(err)
	s.Require().Equal(medianDeviation, sdk.MustNewDecFromStr("0.0225"))

	// delete first historic price, median, and median standard deviation
	app.OracleKeeper.DeleteHistoricPrice(ctx, displayDenom, uint64(ctx.BlockHeight()-3))
	app.OracleKeeper.DeleteMedian(ctx, displayDenom, uint64(ctx.BlockHeight()))
	app.OracleKeeper.DeleteMedianDeviation(ctx, displayDenom, uint64(ctx.BlockHeight()))

	historicPrices = app.OracleKeeper.GetHistoricPrices(ctx, displayDenom)
	s.Require().Equal(len(historicPrices), 3)

	median, err = app.OracleKeeper.GetMedian(ctx, displayDenom, uint64(ctx.BlockHeight()))
	s.Require().Error(err, sdkerrors.Wrap(types.ErrUnknownDenom, displayDenom))
	s.Require().Equal(median, sdk.ZeroDec())

	medianDeviation, err = app.OracleKeeper.GetMedianDeviation(ctx, displayDenom, uint64(ctx.BlockHeight()))
	s.Require().Error(err, sdkerrors.Wrap(types.ErrUnknownDenom, displayDenom))
	s.Require().Equal(median, sdk.ZeroDec())
}
