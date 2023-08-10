package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/x/oracle/types"
)

func (s *IntegrationTestSuite) TestSetHistoraclePricing() {
	app, ctx := s.app, s.ctx

	// update stamp params
	app.OracleKeeper.SetHistoricStampPeriod(ctx, 1)
	app.OracleKeeper.SetMedianStampPeriod(ctx, 3)

	// set historic prices and medians for denom with a similar
	// prefix to test unique prefix safety when iterating for
	// similar prefixes
	displayDenomVariation := displayDenom + "test"
	nonExistingDenom := "nonexistingdenom"

	// add multiple historic prices to store
	exchangeRates := []string{"1.0", "1.2", "1.1", "1.4", "1.1", "1.15", "1.2", "1.3", "1.2"}
	for i, exchangeRate := range exchangeRates {
		app.OracleKeeper.AddHistoricPrice(ctx, displayDenom, sdk.MustNewDecFromStr(exchangeRate))
		app.OracleKeeper.AddHistoricPrice(ctx, displayDenomVariation, sdk.MustNewDecFromStr(exchangeRate))
		if ((i + 1) % int(app.OracleKeeper.MedianStampPeriod(ctx))) == 0 {
			err := app.OracleKeeper.CalcAndSetHistoricMedian(ctx, displayDenom)
			s.Require().NoError(err)
			err = app.OracleKeeper.CalcAndSetHistoricMedian(ctx, displayDenomVariation)
			s.Require().NoError(err)
			err = app.OracleKeeper.CalcAndSetHistoricMedian(ctx, nonExistingDenom)
			s.Require().NoError(err)
		}

		// update blockheight
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	// check medians, num of available medians, and median standard deviation
	medians := app.OracleKeeper.HistoricMedians(ctx, displayDenom, 3)
	s.Require().Equal(len(medians), 3)
	s.Require().Equal(medians[0], types.NewPrice(sdk.MustNewDecFromStr("1.2"), displayDenom, 17))
	s.Require().Equal(medians[1], types.NewPrice(sdk.MustNewDecFromStr("1.125"), displayDenom, 14))
	s.Require().Equal(medians[2], types.NewPrice(sdk.MustNewDecFromStr("1.1"), displayDenom, 11))

	medianDeviation, err := app.OracleKeeper.HistoricMedianDeviation(ctx, displayDenom)
	s.Require().NoError(err)

	s.Require().Equal(medianDeviation, types.NewPrice(sdk.MustNewDecFromStr("0.111803398874989476"), displayDenom, 17))

	// check current price is within the median deviation
	result, err := app.OracleKeeper.WithinHistoricMedianDeviation(ctx, displayDenom)
	s.Require().Equal(result, true)
	s.Require().NoError(err)

	// check median stats of last 3 stamps
	medianOfMedians, numMedians, err := app.OracleKeeper.MedianOfHistoricMedians(ctx, displayDenom, 3)
	s.Require().NoError(err)
	s.Require().Equal(numMedians, uint32(3))
	s.Require().Equal(medianOfMedians, sdk.MustNewDecFromStr("1.125"))
	averageOfMedians, numMedians, err := app.OracleKeeper.AverageOfHistoricMedians(ctx, displayDenom, 3)
	s.Require().NoError(err)
	s.Require().Equal(numMedians, uint32(3))
	s.Require().Equal(averageOfMedians, sdk.MustNewDecFromStr("1.141666666666666666"))
	maxMedian, numMedians, err := app.OracleKeeper.MaxOfHistoricMedians(ctx, displayDenom, 3)
	s.Require().NoError(err)
	s.Require().Equal(numMedians, uint32(3))
	s.Require().Equal(maxMedian, sdk.MustNewDecFromStr("1.2"))
	minMedian, numMedians, err := app.OracleKeeper.MinOfHistoricMedians(ctx, displayDenom, 3)
	s.Require().NoError(err)
	s.Require().Equal(numMedians, uint32(3))
	s.Require().Equal(minMedian, sdk.MustNewDecFromStr("1.1"))

	// check median stats of last 1 stamps
	medianOfMedians, numMedians, err = app.OracleKeeper.MedianOfHistoricMedians(ctx, displayDenom, 1)
	s.Require().NoError(err)
	s.Require().Equal(numMedians, uint32(1))
	s.Require().Equal(medianOfMedians, sdk.MustNewDecFromStr("1.2"))
	averageOfMedians, numMedians, err = app.OracleKeeper.AverageOfHistoricMedians(ctx, displayDenom, 1)
	s.Require().NoError(err)
	s.Require().Equal(numMedians, uint32(1))
	s.Require().Equal(averageOfMedians, sdk.MustNewDecFromStr("1.2"))
	maxMedian, numMedians, err = app.OracleKeeper.MaxOfHistoricMedians(ctx, displayDenom, 1)
	s.Require().NoError(err)
	s.Require().Equal(numMedians, uint32(1))
	s.Require().Equal(maxMedian, sdk.MustNewDecFromStr("1.2"))
	minMedian, numMedians, err = app.OracleKeeper.MinOfHistoricMedians(ctx, displayDenom, 1)
	s.Require().NoError(err)
	s.Require().Equal(numMedians, uint32(1))
	s.Require().Equal(minMedian, sdk.MustNewDecFromStr("1.2"))

	// delete first median
	blockPeriod := (3 - 1) * app.OracleKeeper.MedianStampPeriod(ctx)
	lastStampBlock := uint64(ctx.BlockHeight()) - (uint64(ctx.BlockHeight())%app.OracleKeeper.MedianStampPeriod(ctx) + 1)
	firstStampBlock := lastStampBlock - blockPeriod
	app.OracleKeeper.DeleteHistoricMedian(ctx, displayDenom, firstStampBlock)

	medians = app.OracleKeeper.HistoricMedians(ctx, displayDenom, 3)
	s.Require().Equal(len(medians), 2)
	s.Require().Equal(medians[0], types.NewPrice(sdk.MustNewDecFromStr("1.2"), displayDenom, 17))
	s.Require().Equal(medians[1], types.NewPrice(sdk.MustNewDecFromStr("1.125"), displayDenom, 14))
}
