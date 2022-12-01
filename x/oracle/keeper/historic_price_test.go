package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v3/x/oracle/types"
)

func (s *IntegrationTestSuite) TestSetHistoraclePricing() {
	app, ctx := s.app, s.ctx

	// update stamp params
	app.OracleKeeper.SetHistoricStampPeriod(ctx, 1)
	app.OracleKeeper.SetMedianStampPeriod(ctx, 3)

	// add multiple historic prices to store
	exchangeRates := []string{"1.0", "1.2", "1.1", "1.4", "1.1", "1.15", "1.2", "1.3", "1.2"}
	for i, exchangeRate := range exchangeRates {
		app.OracleKeeper.AddHistoricPrice(ctx, displayDenom, sdk.MustNewDecFromStr(exchangeRate))
		if ((i + 1) % int(app.OracleKeeper.MedianStampPeriod(ctx))) == 0 {
			app.OracleKeeper.CalcAndSetMedian(ctx, displayDenom)
		}

		// update blockheight
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	// check median and median standard deviation
	median, err := app.OracleKeeper.HistoricMedian(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(median, sdk.MustNewDecFromStr("1.2"))

	medianDeviation, err := app.OracleKeeper.HistoricMedianDeviation(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(medianDeviation, sdk.MustNewDecFromStr("0.012499999999999998"))

	// check prices are within the median deviation
	price1 := sdk.MustNewDecFromStr("1.19")
	price2 := sdk.MustNewDecFromStr("1.18")
	result, err := app.OracleKeeper.WithinMedianDeviation(ctx, displayDenom, price1)
	s.Require().Equal(result, true)
	s.Require().NoError(err)
	result, err = app.OracleKeeper.WithinMedianDeviation(ctx, displayDenom, price2)
	s.Require().Equal(result, false)
	s.Require().NoError(err)

	// check median stats of entire block range
	medianOfMedians := app.OracleKeeper.MedianOfMedians(ctx, displayDenom, uint64(ctx.BlockHeight()-9), uint64(ctx.BlockHeight()))
	s.Require().Equal(medianOfMedians, sdk.MustNewDecFromStr("1.125"))
	averageOfMedians := app.OracleKeeper.AverageOfMedians(ctx, displayDenom, uint64(ctx.BlockHeight()-9), uint64(ctx.BlockHeight()))
	s.Require().Equal(averageOfMedians, sdk.MustNewDecFromStr("1.141666666666666666"))
	maxMedian := app.OracleKeeper.MaxMedian(ctx, displayDenom, uint64(ctx.BlockHeight()-9), uint64(ctx.BlockHeight()))
	s.Require().Equal(maxMedian, sdk.MustNewDecFromStr("1.2"))
	minMedian := app.OracleKeeper.MinMedian(ctx, displayDenom, uint64(ctx.BlockHeight()-9), uint64(ctx.BlockHeight()))
	s.Require().Equal(minMedian, sdk.MustNewDecFromStr("1.1"))

	// check median stats of shorter block range
	medianOfMedians = app.OracleKeeper.MedianOfMedians(ctx, displayDenom, uint64(ctx.BlockHeight()-6), uint64(ctx.BlockHeight()-3))
	s.Require().Equal(medianOfMedians, sdk.MustNewDecFromStr("1.125"))
	averageOfMedians = app.OracleKeeper.AverageOfMedians(ctx, displayDenom, uint64(ctx.BlockHeight()-6), uint64(ctx.BlockHeight()-3))
	s.Require().Equal(averageOfMedians, sdk.MustNewDecFromStr("1.125"))
	maxMedian = app.OracleKeeper.MaxMedian(ctx, displayDenom, uint64(ctx.BlockHeight()-6), uint64(ctx.BlockHeight()-3))
	s.Require().Equal(maxMedian, sdk.MustNewDecFromStr("1.125"))
	minMedian = app.OracleKeeper.MinMedian(ctx, displayDenom, uint64(ctx.BlockHeight()-6), uint64(ctx.BlockHeight()-3))
	s.Require().Equal(minMedian, sdk.MustNewDecFromStr("1.125"))

	// delete last historic price, median, and median standard deviation
	app.OracleKeeper.DeleteHistoricPrice(ctx, displayDenom, uint64(ctx.BlockHeight()-1))
	app.OracleKeeper.DeleteMedian(ctx, displayDenom, uint64(ctx.BlockHeight()-1))
	app.OracleKeeper.DeleteMedianDeviation(ctx, displayDenom, uint64(ctx.BlockHeight()-1))

	median, err = app.OracleKeeper.HistoricMedian(ctx, displayDenom)
	s.Require().Error(err, sdkerrors.Wrap(types.ErrUnknownDenom, displayDenom))
	s.Require().Equal(median, sdk.ZeroDec())

	medianDeviation, err = app.OracleKeeper.HistoricMedianDeviation(ctx, displayDenom)
	s.Require().Error(err, sdkerrors.Wrap(types.ErrUnknownDenom, displayDenom))
	s.Require().Equal(median, sdk.ZeroDec())

	result, err = app.OracleKeeper.WithinMedianDeviation(ctx, displayDenom, price1)
	s.Require().Error(err, sdkerrors.Wrap(types.ErrUnknownDenom, displayDenom))
	s.Require().Equal(result, false)
}
