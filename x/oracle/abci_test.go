package oracle_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeapp "github.com/umee-network/umee/v3/app"
	appparams "github.com/umee-network/umee/v3/app/params"
	"github.com/umee-network/umee/v3/x/oracle/keeper"
	"github.com/umee-network/umee/v3/x/oracle"
	"github.com/umee-network/umee/v3/x/oracle/types"
)

const (
	displayDenom string = appparams.DisplayDenom
	bondDenom    string = appparams.BondDenom
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *umeeapp.UmeeApp
}

const (
	initialPower = int64(10000000000)
)

// clearHistoricPrices deletes all historic prices of a given denom in the store.
func clearHistoricPrices(
	ctx sdk.Context,
	k keeper.Keeper,
	denom string,
	) {
	stampPeriod	:= int(k.HistoricStampPeriod(ctx))
	numStamps := int(k.MaximumPriceStamps(ctx))
	for i := 0; i < numStamps; i++  {
		k.DeleteHistoricPrice(ctx, denom, uint64(ctx.BlockHeight()) - uint64(i*stampPeriod))
	}
}

// clearHistoricMedians deletes all historic medians of a given denom in the store.
func clearHistoricMedians(
	ctx sdk.Context,
	k keeper.Keeper,
	denom string,
	) {
	stampPeriod	:= int(k.MedianStampPeriod(ctx))
	numStamps := int(k.MaximumMedianStamps(ctx))
	for i := 0; i < numStamps; i++ {
		k.DeleteHistoricMedian(ctx, denom, uint64(ctx.BlockHeight()) - uint64(i*stampPeriod))
	}
}

// clearHistoricMedianDeviations deletes all historic median deviations of a given
// denom in the store.
func clearHistoricMedianDeviations(
	ctx sdk.Context,
	k keeper.Keeper,
	denom string,
	) {
	stampPeriod	:= int(k.MedianStampPeriod(ctx))
	numStamps := int(k.MaximumMedianStamps(ctx))
	for i := 0; i < numStamps; i++  {
		k.DeleteHistoricMedianDeviation(ctx, denom, uint64(ctx.BlockHeight()) - uint64(i*stampPeriod))
	}
}

func (s *IntegrationTestSuite) SetupTest() {
	require := s.Require()
	isCheckTx := false
	app := umeeapp.Setup(s.T())
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
	})

	oracle.InitGenesis(ctx, app.OracleKeeper, *types.DefaultGenesisState())

	sh := teststaking.NewHelper(s.T(), ctx, *app.StakingKeeper)
	sh.Denom = bondDenom
	amt := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)

	// mint and send coins to validator
	require.NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	require.NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, initCoins))

	sh.CreateValidator(valAddr, valPubKey, amt, true)

	staking.EndBlocker(ctx, *app.StakingKeeper)

	s.app = app
	s.ctx = ctx
}

// Test addresses
var (
	valPubKeys = simapp.CreateTestPubKeys(1)

	valPubKey = valPubKeys[0]
	pubKey    = secp256k1.GenPrivKey().PubKey()
	addr      = sdk.AccAddress(pubKey.Address())
	valAddr   = sdk.ValAddress(pubKey.Address())

	initTokens = sdk.TokensFromConsensusPower(initialPower, sdk.DefaultPowerReduction)
	initCoins  = sdk.NewCoins(sdk.NewCoin(bondDenom, initTokens))
)

var historacleTestCases = []struct {
	exchangeRates                         []string
	expectedHistoricMedians               []sdk.Dec
	expectedHistoricMedianDeviation       sdk.Dec
	expectedWithinHistoricMedianDeviation bool
	expectedMedianOfHistoricMedians       sdk.Dec
	expectedAverageOfHistoricMedians      sdk.Dec
	expectedMinOfHistoricMedians          sdk.Dec
	expectedMaxOfHistoricMedians          sdk.Dec
}{
	{
		[]string{"1.0", "1.2", "1.1", "1.4", "1.1", "1.15",
			"1.2", "1.3", "1.2", "1.12", "1.2", "1.15",
			"1.17", "1.1", "1.0", "1.16", "1.15", "1.12"},
		[]sdk.Dec{
			sdk.MustNewDecFromStr("1.155"),
			sdk.MustNewDecFromStr("1.16"),
			sdk.MustNewDecFromStr("1.175"),
			sdk.MustNewDecFromStr("1.2"),
		},
		sdk.MustNewDecFromStr("0.009724999999999997"),
		false,
		sdk.MustNewDecFromStr("1.1675"),
		sdk.MustNewDecFromStr("1.1725"),
		sdk.MustNewDecFromStr("1.155"),
		sdk.MustNewDecFromStr("1.2"),
	},
	{
		[]string{"2.3", "2.12", "2.14", "2.24", "2.18", "2.15",
			"2.51", "2.59", "2.67", "2.76", "2.89", "2.85",
			"3.17", "3.15", "3.35", "3.56", "3.55", "3.49"},
		[]sdk.Dec{
			sdk.MustNewDecFromStr("3.02"),
			sdk.MustNewDecFromStr("2.715"),
			sdk.MustNewDecFromStr("2.405"),
			sdk.MustNewDecFromStr("2.24"),
		},
		sdk.MustNewDecFromStr("0.145091666666666664"),
		false,
		sdk.MustNewDecFromStr("2.56"),
		sdk.MustNewDecFromStr("2.595"),
		sdk.MustNewDecFromStr("2.24"),
		sdk.MustNewDecFromStr("3.02"),
	},
	{
		[]string{"5.2", "5.25", "5.31", "5.22", "5.14", "5.15",
			"4.85", "4.72", "4.52", "4.47", "4.36", "4.22",
			"4.11", "4.04", "3.92", "3.82", "3.85", "3.83"},
		[]sdk.Dec{
			sdk.MustNewDecFromStr("4.165"),
			sdk.MustNewDecFromStr("4.495"),
			sdk.MustNewDecFromStr("4.995"),
			sdk.MustNewDecFromStr("5.15"),
		},
		sdk.MustNewDecFromStr("0.194024999999999997"),
		false,
		sdk.MustNewDecFromStr("4.745"),
		sdk.MustNewDecFromStr("4.70125"),
		sdk.MustNewDecFromStr("4.165"),
		sdk.MustNewDecFromStr("5.15"),
	},
}

func (s *IntegrationTestSuite) TestEndblockerExperimentalFlag() {
	app, ctx := s.app, s.ctx

	// add historic price and calcSet median stats
	app.OracleKeeper.AddHistoricPrice(s.ctx, displayDenom, sdk.MustNewDecFromStr("1.0"))
	app.OracleKeeper.CalcAndSetHistoricMedian(s.ctx, displayDenom)
	medianPruneBlock := ctx.BlockHeight() + int64(types.DefaultMaximumMedianStamps*types.DefaultMedianStampPeriod)
	ctx = ctx.WithBlockHeight(medianPruneBlock)

	// with experimental flag off median doesn't get deleted
	oracle.EndBlocker(ctx, app.OracleKeeper, false)
	medians := []types.Price{}
	app.OracleKeeper.IterateAllMedianPrices(
		ctx,
		func(median types.Price) bool {
			medians = append(medians, median)
			return false
		},
	)
	s.Require().Equal(1, len(medians))

	// with experimental flag on median gets deleted
	oracle.EndBlocker(ctx, app.OracleKeeper, true)
	experimentalMedians := []types.Price{}
	app.OracleKeeper.IterateAllMedianPrices(
		ctx,
		func(median types.Price) bool {
			medians = append(experimentalMedians, median)
			return false
		},
	)
	s.Require().Equal(0, len(experimentalMedians))
}

func (s *IntegrationTestSuite) TestEndblockerHistoracle() {
	app, ctx := s.app, s.ctx

	// update historacle params
	app.OracleKeeper.SetHistoricStampPeriod(ctx, 5)
	app.OracleKeeper.SetMedianStampPeriod(ctx, 15)
	app.OracleKeeper.SetMaximumPriceStamps(ctx, 12)
	app.OracleKeeper.SetMaximumMedianStamps(ctx, 4)

	for _, tc := range historacleTestCases {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(app.OracleKeeper.MedianStampPeriod(ctx)-1))

		for _, exchangeRate := range tc.exchangeRates {
			var tuples types.ExchangeRateTuples
			for _, denom := range app.OracleKeeper.AcceptList(ctx) {
				tuples = append(tuples, types.ExchangeRateTuple{
					Denom:        denom.SymbolDenom,
					ExchangeRate: sdk.MustNewDecFromStr(exchangeRate),
				})
			}

			prevote := types.AggregateExchangeRatePrevote{
				Hash:        "hash",
				Voter:       valAddr.String(),
				SubmitBlock: uint64(ctx.BlockHeight()),
			}
			app.OracleKeeper.SetAggregateExchangeRatePrevote(ctx, valAddr, prevote)
			oracle.EndBlocker(ctx, app.OracleKeeper, true)

			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(app.OracleKeeper.VotePeriod(ctx)))
			vote := types.AggregateExchangeRateVote{
				ExchangeRateTuples: tuples,
				Voter:              valAddr.String(),
			}
			app.OracleKeeper.SetAggregateExchangeRateVote(ctx, valAddr, vote)
			oracle.EndBlocker(ctx, app.OracleKeeper, true)
		}

		for _, denom := range app.OracleKeeper.AcceptList(ctx) {
			// query for past 6 medians (should only get 4 back since max median stamps is set to 4)
			medians := app.OracleKeeper.HistoricMedians(ctx, denom.SymbolDenom, 6)
			s.Require().Equal(4, len(medians))
			s.Require().Equal(tc.expectedHistoricMedians, medians)

			medianHistoricDeviation, err := app.OracleKeeper.HistoricMedianDeviation(ctx, denom.SymbolDenom)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedHistoricMedianDeviation, medianHistoricDeviation)

			withinHistoricMedianDeviation, err := app.OracleKeeper.WithinHistoricMedianDeviation(ctx, denom.SymbolDenom)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedWithinHistoricMedianDeviation, withinHistoricMedianDeviation)

			medianOfHistoricMedians, err := app.OracleKeeper.MedianOfHistoricMedians(ctx, denom.SymbolDenom, 6)
			s.Require().Equal(tc.expectedMedianOfHistoricMedians, medianOfHistoricMedians)

			averageOfHistoricMedians, err := app.OracleKeeper.AverageOfHistoricMedians(ctx, denom.SymbolDenom, 6)
			s.Require().Equal(tc.expectedAverageOfHistoricMedians, averageOfHistoricMedians)

			minOfHistoricMedians, err := app.OracleKeeper.MinOfHistoricMedians(ctx, denom.SymbolDenom, 6)
			s.Require().Equal(tc.expectedMinOfHistoricMedians, minOfHistoricMedians)

			maxOfHistoricMedians, err := app.OracleKeeper.MaxOfHistoricMedians(ctx, denom.SymbolDenom, 6)
			s.Require().Equal(tc.expectedMaxOfHistoricMedians, maxOfHistoricMedians)

			clearHistoricPrices(ctx, app.OracleKeeper, denom.SymbolDenom)
			clearHistoricMedians(ctx, app.OracleKeeper, denom.SymbolDenom)
			clearHistoricMedianDeviations(ctx, app.OracleKeeper, denom.SymbolDenom)
		}

		ctx = ctx.WithBlockHeight(0)
	}
}

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
