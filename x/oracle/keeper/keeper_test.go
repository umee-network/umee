package keeper_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	umeeapp "github.com/umee-network/umee/app"
	umeeappbeta "github.com/umee-network/umee/app/beta"
	"github.com/umee-network/umee/x/oracle"
	"github.com/umee-network/umee/x/oracle/keeper"
	"github.com/umee-network/umee/x/oracle/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *umeeappbeta.UmeeApp
	queryClient types.QueryClient
}

const (
	initialPower = int64(10000000000)
)

func (s *IntegrationTestSuite) SetupTest() {
	app := umeeappbeta.Setup(s.T(), false, 1)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})

	app.OracleKeeper = keeper.NewKeeper(
		app.AppCodec(),
		app.GetKey(types.ModuleName),
		app.GetSubspace(types.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper,
		app.StakingKeeper,
		distrtypes.ModuleName,
	)

	oracle.InitGenesis(ctx, app.OracleKeeper, *types.DefaultGenesisState())

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(app.OracleKeeper))

	sh := staking.NewHandler(app.StakingKeeper)

	// Validator created
	amt := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)

	// mint and send coins to validators
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, initCoins))
	_, err := sh(ctx, NewTestMsgCreateValidator(valAddr, valPubKey, amt))
	s.Require().NoError(err)

	staking.EndBlocker(ctx, app.StakingKeeper)

	s.app = app
	s.ctx = ctx
	s.queryClient = types.NewQueryClient(queryHelper)
}

// Test addresses
var (
	valPubKeys = simapp.CreateTestPubKeys(1)
	valPubKey  = valPubKeys[0]
	pubKey     = secp256k1.GenPrivKey().PubKey()
	addr       = sdk.AccAddress(pubKey.Address())
	valAddr    = sdk.ValAddress(pubKey.Address())

	initTokens = sdk.TokensFromConsensusPower(initialPower, sdk.DefaultPowerReduction)
	initCoins  = sdk.NewCoins(sdk.NewCoin(umeeapp.BondDenom, initTokens))
)

// NewTestMsgCreateValidator test msg creator
func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey cryptotypes.PubKey, amt sdk.Int) *stakingtypes.MsgCreateValidator {
	commission := stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
	msg, _ := stakingtypes.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(types.UmeeDenom, amt),
		stakingtypes.Description{}, commission, sdk.OneInt(),
	)

	return msg
}

func (s *IntegrationTestSuite) Test_SetFeederDelegation() {
	app, ctx := s.app, s.ctx

	feederAddr := sdk.AccAddress([]byte("addr________________"))
	feederAcc := app.AccountKeeper.NewAccountWithAddress(ctx, feederAddr)
	app.AccountKeeper.SetAccount(ctx, feederAcc)

	err := s.app.OracleKeeper.ValidateFeeder(ctx, feederAddr, valAddr)
	s.Require().Error(err)

	s.app.OracleKeeper.SetFeederDelegation(ctx, valAddr, feederAddr)

	err = s.app.OracleKeeper.ValidateFeeder(ctx, feederAddr, valAddr)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) Test_GetFeederDelegation() {
	app, ctx := s.app, s.ctx

	feederAddr := sdk.AccAddress([]byte("addr________________"))
	feederAcc := app.AccountKeeper.NewAccountWithAddress(ctx, feederAddr)
	app.AccountKeeper.SetAccount(ctx, feederAcc)

	s.app.OracleKeeper.SetFeederDelegation(ctx, valAddr, feederAddr)
	s.Require().Equal(app.OracleKeeper.GetFeederDelegation(ctx, valAddr), feederAddr)
}

func (s *IntegrationTestSuite) Test_MissCounter() {
	app, ctx := s.app, s.ctx
	missCounter := uint64(rand.Intn(100))

	s.Require().Equal(app.OracleKeeper.GetMissCounter(ctx, valAddr), uint64(0))
	app.OracleKeeper.SetMissCounter(ctx, valAddr, missCounter)
	s.Require().Equal(app.OracleKeeper.GetMissCounter(ctx, valAddr), missCounter)

	app.OracleKeeper.DeleteMissCounter(ctx, valAddr)
	s.Require().Equal(app.OracleKeeper.GetMissCounter(ctx, valAddr), uint64(0))
}

func (s *IntegrationTestSuite) Test_AggregateExchangeRatePrevote() {
	app, ctx := s.app, s.ctx

	prevote := types.AggregateExchangeRatePrevote{
		Hash:        "hash",
		Voter:       addr.String(),
		SubmitBlock: 0,
	}
	app.OracleKeeper.SetAggregateExchangeRatePrevote(ctx, valAddr, prevote)

	_, err := app.OracleKeeper.GetAggregateExchangeRatePrevote(ctx, valAddr)
	s.Require().NoError(err)

	app.OracleKeeper.DeleteAggregateExchangeRatePrevote(ctx, valAddr)

	_, err = app.OracleKeeper.GetAggregateExchangeRatePrevote(ctx, valAddr)
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) Test_AggregateExchangeRateVote() {
	app, ctx := s.app, s.ctx

	var tuples types.ExchangeRateTuples
	tuples = append(tuples, types.ExchangeRateTuple{
		Denom:        "UMEE",
		ExchangeRate: sdk.ZeroDec(),
	})

	vote := types.AggregateExchangeRateVote{
		ExchangeRateTuples: tuples,
		Voter:              addr.String(),
	}
	app.OracleKeeper.SetAggregateExchangeRateVote(ctx, valAddr, vote)

	_, err := app.OracleKeeper.GetAggregateExchangeRateVote(ctx, valAddr)
	s.Require().NoError(err)

	app.OracleKeeper.DeleteAggregateExchangeRateVote(ctx, valAddr)

	_, err = app.OracleKeeper.GetAggregateExchangeRateVote(ctx, valAddr)
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) Test_SetExchangeRateWithEvent() {
	app, ctx := s.app, s.ctx
	app.OracleKeeper.SetExchangeRateWithEvent(ctx, "umee", sdk.OneDec())
}

func (s *IntegrationTestSuite) Test_GetExchangeRate_USD() {
	app, ctx := s.app, s.ctx

	rate, err := app.OracleKeeper.GetExchangeRate(ctx, "uusd")
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.OneDec())
}

func (s *IntegrationTestSuite) Test_GetExchangeRate_InvalidDenom() {
	app, ctx := s.app, s.ctx

	_, err := app.OracleKeeper.GetExchangeRate(ctx, "uxyz")
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) Test_GetExchangeRate_NotSet() {
	app, ctx := s.app, s.ctx

	_, err := app.OracleKeeper.GetExchangeRate(ctx, "uumee")
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) Test_GetExchangeRate_Valid() {
	app, ctx := s.app, s.ctx

	app.OracleKeeper.SetExchangeRate(ctx, "umee", sdk.OneDec())
	_, err := app.OracleKeeper.GetExchangeRate(ctx, "uumee")
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) Test_DeleteExchangeRate() {
	app, ctx := s.app, s.ctx

	app.OracleKeeper.SetExchangeRate(ctx, "uumee", sdk.OneDec())
	app.OracleKeeper.DeleteExchangeRate(ctx, "umee")
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
