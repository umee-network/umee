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

	// add multiple historic prices to store
	exchangeRates := []string{"1.0", "1.2", "1.1", "1.4"}
	for _, exchangeRate := range exchangeRates {
		app.OracleKeeper.AddHistoricPrice(ctx, displayDenom, sdk.MustNewDecFromStr(exchangeRate))

		// update blockheight
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

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

	median, err = app.OracleKeeper.GetMedian(ctx, displayDenom, uint64(ctx.BlockHeight()))
	s.Require().Error(err, sdkerrors.Wrap(types.ErrUnknownDenom, displayDenom))
	s.Require().Equal(median, sdk.ZeroDec())

	medianDeviation, err = app.OracleKeeper.GetMedianDeviation(ctx, displayDenom, uint64(ctx.BlockHeight()))
	s.Require().Error(err, sdkerrors.Wrap(types.ErrUnknownDenom, displayDenom))
	s.Require().Equal(median, sdk.ZeroDec())
}
