package oracle_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	initialPower = int64(10000)
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

	// mint and send coins to validator
	require.NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	require.NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr1, initCoins))
	require.NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	require.NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr2, initCoins))

	sh.CreateValidatorWithValPower(valAddr1, valPubKey1, 7000, true)
	sh.CreateValidatorWithValPower(valAddr2, valPubKey2, 3000, true)

	staking.EndBlocker(ctx, *app.StakingKeeper)

	s.app = app
	s.ctx = ctx
}

// Test addresses
var (
	valPubKeys = simapp.CreateTestPubKeys(2)

	valPubKey1 = valPubKeys[0]
	pubKey1    = secp256k1.GenPrivKey().PubKey()
	addr1      = sdk.AccAddress(pubKey1.Address())
	valAddr1   = sdk.ValAddress(pubKey1.Address())

	valPubKey2 = valPubKeys[1]
	pubKey2    = secp256k1.GenPrivKey().PubKey()
	addr2      = sdk.AccAddress(pubKey2.Address())
	valAddr2   = sdk.ValAddress(pubKey2.Address())

	initTokens = sdk.TokensFromConsensusPower(initialPower, sdk.DefaultPowerReduction)
	initCoins  = sdk.NewCoins(sdk.NewCoin(bondDenom, initTokens))
)

func (s *IntegrationTestSuite) TestEndBlockerVoteThreshold() {
	app, ctx := s.app, s.ctx
	originalBlockHeight := ctx.BlockHeight()
	ctx = ctx.WithBlockHeight(1)
	preVoteBlockDiff := int64(app.OracleKeeper.VotePeriod(ctx) / 2)
	voteBlockDiff := int64(app.OracleKeeper.VotePeriod(ctx)/2 + 1)

	var (
		val1Tuples   types.ExchangeRateTuples
		val2Tuples   types.ExchangeRateTuples
		val1PreVotes types.AggregateExchangeRatePrevote
		val2PreVotes types.AggregateExchangeRatePrevote
		val1Votes    types.AggregateExchangeRateVote
		val2Votes    types.AggregateExchangeRateVote
	)
	for _, denom := range app.OracleKeeper.AcceptList(ctx) {
		val1Tuples = append(val1Tuples, types.ExchangeRateTuple{
			Denom:        denom.SymbolDenom,
			ExchangeRate: sdk.MustNewDecFromStr("1.0"),
		})
		val2Tuples = append(val2Tuples, types.ExchangeRateTuple{
			Denom:        denom.SymbolDenom,
			ExchangeRate: sdk.MustNewDecFromStr("0.5"),
		})
	}

	val1PreVotes = types.AggregateExchangeRatePrevote{
		Hash:        "hash1",
		Voter:       valAddr1.String(),
		SubmitBlock: uint64(ctx.BlockHeight()),
	}
	val2PreVotes = types.AggregateExchangeRatePrevote{
		Hash:        "hash2",
		Voter:       valAddr2.String(),
		SubmitBlock: uint64(ctx.BlockHeight()),
	}

	val1Votes = types.AggregateExchangeRateVote{
		ExchangeRateTuples: val1Tuples,
		Voter:              valAddr1.String(),
	}
	val2Votes = types.AggregateExchangeRateVote{
		ExchangeRateTuples: val2Tuples,
		Voter:              valAddr2.String(),
	}

	// total voting power per denom is 100%
	app.OracleKeeper.SetAggregateExchangeRatePrevote(ctx, valAddr1, val1PreVotes)
	app.OracleKeeper.SetAggregateExchangeRatePrevote(ctx, valAddr2, val2PreVotes)
	oracle.EndBlocker(ctx, app.OracleKeeper)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + voteBlockDiff)
	app.OracleKeeper.SetAggregateExchangeRateVote(ctx, valAddr1, val1Votes)
	app.OracleKeeper.SetAggregateExchangeRateVote(ctx, valAddr2, val2Votes)
	oracle.EndBlocker(ctx, app.OracleKeeper)

	for _, denom := range app.OracleKeeper.AcceptList(ctx) {
		rate, err := app.OracleKeeper.GetExchangeRate(ctx, denom.SymbolDenom)
		s.Require().NoError(err)
		s.Require().Equal(sdk.MustNewDecFromStr("1.0"), rate)
	}

	// update prevotes' block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + preVoteBlockDiff)
	val1PreVotes.SubmitBlock = uint64(ctx.BlockHeight())
	val2PreVotes.SubmitBlock = uint64(ctx.BlockHeight())

	// total voting power per denom is 30%
	app.OracleKeeper.SetAggregateExchangeRatePrevote(ctx, valAddr2, val2PreVotes)
	oracle.EndBlocker(ctx, app.OracleKeeper)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + voteBlockDiff)
	app.OracleKeeper.SetAggregateExchangeRateVote(ctx, valAddr2, val2Votes)
	oracle.EndBlocker(ctx, app.OracleKeeper)

	for _, denom := range app.OracleKeeper.AcceptList(ctx) {
		rate, err := app.OracleKeeper.GetExchangeRate(ctx, denom.SymbolDenom)
		s.Require().ErrorIs(err, sdkerrors.Wrap(types.ErrUnknownDenom, denom.SymbolDenom))
		s.Require().Equal(sdk.ZeroDec(), rate)
	}

	// update prevotes' block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + preVoteBlockDiff)
	val1PreVotes.SubmitBlock = uint64(ctx.BlockHeight())
	val2PreVotes.SubmitBlock = uint64(ctx.BlockHeight())

	// umee has 100% power, and atom has 30%
	val1Tuples = types.ExchangeRateTuples{
		types.ExchangeRateTuple{
			Denom:        "umee",
			ExchangeRate: sdk.MustNewDecFromStr("1.0"),
		},
	}
	val2Tuples = types.ExchangeRateTuples{
		types.ExchangeRateTuple{
			Denom:        "umee",
			ExchangeRate: sdk.MustNewDecFromStr("0.5"),
		},
		types.ExchangeRateTuple{
			Denom:        "atom",
			ExchangeRate: sdk.MustNewDecFromStr("0.5"),
		},
	}
	val1Votes.ExchangeRateTuples = val1Tuples
	val2Votes.ExchangeRateTuples = val2Tuples

	app.OracleKeeper.SetAggregateExchangeRatePrevote(ctx, valAddr1, val1PreVotes)
	app.OracleKeeper.SetAggregateExchangeRatePrevote(ctx, valAddr2, val2PreVotes)
	oracle.EndBlocker(ctx, app.OracleKeeper)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + voteBlockDiff)
	app.OracleKeeper.SetAggregateExchangeRateVote(ctx, valAddr1, val1Votes)
	app.OracleKeeper.SetAggregateExchangeRateVote(ctx, valAddr2, val2Votes)
	oracle.EndBlocker(ctx, app.OracleKeeper)

	rate, err := app.OracleKeeper.GetExchangeRate(ctx, "umee")
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("1.0"), rate)
	rate, err = app.OracleKeeper.GetExchangeRate(ctx, "atom")
	s.Require().ErrorIs(err, sdkerrors.Wrap(types.ErrUnknownDenom, "atom"))
	s.Require().Equal(sdk.ZeroDec(), rate)

	ctx = ctx.WithBlockHeight(originalBlockHeight)
}

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
				Voter:              valAddr1.String(),
			}
			app.OracleKeeper.SetAggregateExchangeRateVote(ctx, valAddr1, vote)
			oracle.EndBlocker(ctx, app.OracleKeeper)
		}

		for denom, denomRates := range exchangeRates {
			// check median
			expectedMedian, err := decmath.Median(denomRates)
			s.Require().NoError(err)

			medians := app.OracleKeeper.AllMedianPrices(ctx)
			medians = *medians.FilterByBlock(uint64(blockHeight)).FilterByDenom(denom)
			actualMedian := medians[0].ExchangeRateTuple.ExchangeRate
			s.Require().Equal(expectedMedian, actualMedian)

			// check median deviation
			expectedMedianDeviation, err := decmath.MedianDeviation(actualMedian, denomRates)
			s.Require().NoError(err)

			medianDeviations := app.OracleKeeper.AllMedianDeviationPrices(ctx)
			medianDeviations = *medianDeviations.FilterByBlock(uint64(blockHeight)).FilterByDenom(denom)
			actualMedianDeviation := medianDeviations[0].ExchangeRateTuple.ExchangeRate
			s.Require().Equal(expectedMedianDeviation, actualMedianDeviation)
		}
	}
	numberOfAssets := int64(len(exchangeRates))

	historicPrices := app.OracleKeeper.AllHistoricPrices(ctx)
	s.Require().Equal(maximumPriceStamps*numberOfAssets, int64(len(historicPrices)))

	for i := int64(0); i < maximumPriceStamps; i++ {
		expectedBlockNum := blockHeight - (historicStampPeriod * (maximumPriceStamps - int64(i+1)))
		actualBlockNum := historicPrices[i].BlockNum
		s.Require().Equal(expectedBlockNum, int64(actualBlockNum))
	}

	medians := app.OracleKeeper.AllMedianPrices(ctx)
	s.Require().Equal(maximumMedianStamps*numberOfAssets, int64(len(medians)))

	for i := int64(0); i < maximumMedianStamps; i++ {
		expectedBlockNum := blockHeight - (medianStampPeriod * (maximumMedianStamps - int64(i+1)))
		actualBlockNum := medians[i].BlockNum
		s.Require().Equal(expectedBlockNum, int64(actualBlockNum))
	}

	medianDeviations := app.OracleKeeper.AllMedianPrices(ctx)
	s.Require().Equal(maximumMedianStamps*numberOfAssets, int64(len(medianDeviations)))

	for i := int64(0); i < maximumMedianStamps; i++ {
		expectedBlockNum := blockHeight - (medianStampPeriod * (maximumMedianStamps - int64(i+1)))
		actualBlockNum := medianDeviations[i].BlockNum
		s.Require().Equal(expectedBlockNum, int64(actualBlockNum))
	}
}

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
