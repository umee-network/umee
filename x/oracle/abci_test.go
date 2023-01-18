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

	umeeapp "github.com/umee-network/umee/v4/app"
	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/util/decmath"
	"github.com/umee-network/umee/v4/x/oracle"
	"github.com/umee-network/umee/v4/x/oracle/types"
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

var exchangeRates = map[string][]sdk.Dec{
	"ATOM": {
		sdk.MustNewDecFromStr("12.99"),
		sdk.MustNewDecFromStr("12.22"),
		sdk.MustNewDecFromStr("13.1"),
		sdk.MustNewDecFromStr("11.6"),
	},
	"UMEE": {
		sdk.MustNewDecFromStr("1.89"),
		sdk.MustNewDecFromStr("2.05"),
		sdk.MustNewDecFromStr("2.34"),
		sdk.MustNewDecFromStr("1.71"),
	},
}

func (s *IntegrationTestSuite) TestEndblockerHistoracle() {
	app, ctx := s.app, s.ctx
	blockHeight := ctx.BlockHeight()

	var historicStampPeriod int64 = 5
	var medianStampPeriod int64 = 20
	var maximumPriceStamps int64 = 4
	var maximumMedianStamps int64 = 3

	app.OracleKeeper.SetHistoricStampPeriod(ctx, uint64(historicStampPeriod))
	app.OracleKeeper.SetMedianStampPeriod(ctx, uint64(medianStampPeriod))
	app.OracleKeeper.SetMaximumPriceStamps(ctx, uint64(maximumPriceStamps))
	app.OracleKeeper.SetMaximumMedianStamps(ctx, uint64(maximumMedianStamps))

	// Start at the last block of the first stamp period
	blockHeight += medianStampPeriod
	blockHeight += -1
	ctx = ctx.WithBlockHeight(blockHeight)

	for i := int64(0); i <= maximumMedianStamps; i++ {
		for j := int64(0); j < maximumPriceStamps; j++ {

			blockHeight += historicStampPeriod
			ctx = ctx.WithBlockHeight(blockHeight)

			var tuples = types.ExchangeRateTuples{}
			for denom, prices := range exchangeRates {
				tuples = append(tuples, types.ExchangeRateTuple{
					Denom:        denom,
					ExchangeRate: prices[j],
				})
			}

			vote := types.AggregateExchangeRateVote{
				ExchangeRateTuples: tuples,
				Voter:              valAddr.String(),
			}
			app.OracleKeeper.SetAggregateExchangeRateVote(ctx, valAddr, vote)
			oracle.EndBlocker(ctx, app.OracleKeeper)
		}

		for denom, denomRates := range exchangeRates {
			// check median
			expectedMedian, err := decmath.Median(denomRates)
			s.Require().NoError(err)

			medians := app.OracleKeeper.AllMedianPrices(ctx)
			medians = *medians.FilterByBlock(uint64(blockHeight)).FilterByDenom(denom)
			actualMedian := medians[0].ExchangeRateTuple.ExchangeRate
			s.Require().Equal(actualMedian, expectedMedian)

			// check median deviation
			expectedMedianDeviation, err := decmath.MedianDeviation(actualMedian, denomRates)
			s.Require().NoError(err)

			medianDeviations := app.OracleKeeper.AllMedianDeviationPrices(ctx)
			medianDeviations = *medianDeviations.FilterByBlock(uint64(blockHeight)).FilterByDenom(denom)
			actualMedianDeviation := medianDeviations[0].ExchangeRateTuple.ExchangeRate
			s.Require().Equal(actualMedianDeviation, expectedMedianDeviation)
		}
	}
	historicPrices := app.OracleKeeper.AllHistoricPrices(ctx)
	s.Require().Equal(int64(len(historicPrices)), maximumPriceStamps*2)

	medians := app.OracleKeeper.AllMedianPrices(ctx)
	s.Require().Equal(int64(len(medians)), maximumMedianStamps*2)

	medianDeviations := app.OracleKeeper.AllMedianPrices(ctx)
	s.Require().Equal(int64(len(medianDeviations)), maximumMedianStamps*2)
}

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
