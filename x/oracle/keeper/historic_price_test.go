package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v3/x/oracle/types"
)

func (s *IntegrationTestSuite) TestSetHistoraclePricing() {
	app, ctx := s.app, s.ctx

	// add multiple historic prices to store
	exchangeRates := []string{"1.0", "1.2", "1.1", "1.4"}
	for i, exchangeRate := range exchangeRates {
		// update blockheight
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(i))

		app.OracleKeeper.AddHistoricPrice(ctx, displayDenom, sdk.MustNewDecFromStr(exchangeRate))
		app.OracleKeeper.CalcAndSetMedian(ctx, displayDenom)
	}

	// set and check median and median standard deviation
	app.OracleKeeper.CalcAndSetMedian(ctx, displayDenom)
	median, err := app.OracleKeeper.GetMedian(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(median, sdk.MustNewDecFromStr("1.15"))

	medianDeviation, err := app.OracleKeeper.GetMedianDeviation(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(medianDeviation, sdk.MustNewDecFromStr("0.0225"))

	// check prices are within the median deviation
	price1 := sdk.MustNewDecFromStr("1.13")
	price2 := sdk.MustNewDecFromStr("1.12")
	result, err := app.OracleKeeper.WithinMedianDeviation(ctx, displayDenom, price1)
	s.Require().Equal(result, true)
	s.Require().NoError(err)
	result, err = app.OracleKeeper.WithinMedianDeviation(ctx, displayDenom, price2)
	s.Require().Equal(result, false)
	s.Require().NoError(err)

	// delete first historic price, median, and median standard deviation
	app.OracleKeeper.DeleteHistoricPrice(ctx, displayDenom, uint64(ctx.BlockHeight()-3))
	app.OracleKeeper.DeleteMedian(ctx, displayDenom)
	app.OracleKeeper.DeleteMedianDeviation(ctx, displayDenom)

	median, err = app.OracleKeeper.GetMedian(ctx, displayDenom)
	s.Require().Error(err, sdkerrors.Wrap(types.ErrUnknownDenom, displayDenom))
	s.Require().Equal(median, sdk.ZeroDec())

	medianDeviation, err = app.OracleKeeper.GetMedianDeviation(ctx, displayDenom)
	s.Require().Error(err, sdkerrors.Wrap(types.ErrUnknownDenom, displayDenom))
	s.Require().Equal(median, sdk.ZeroDec())

	result, err = app.OracleKeeper.WithinMedianDeviation(ctx, displayDenom, price1)
	s.Require().Error(err, sdkerrors.Wrap(types.ErrUnknownDenom, displayDenom))
	s.Require().Equal(result, false)
}
