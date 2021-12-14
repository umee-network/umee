package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	umeeapp "github.com/umee-network/umee/app"
	"github.com/umee-network/umee/x/oracle"
	"github.com/umee-network/umee/x/oracle/keeper"
	"github.com/umee-network/umee/x/oracle/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *umeeapp.UmeeApp
	queryClient types.QueryClient
}

const faucetAccountName = "faucet"

func (s *IntegrationTestSuite) SetupTest() {
	app := umeeapp.Setup(s.T(), false, 1)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})
	encodingConfig := umeeapp.MakeEncodingConfig()

	transientKeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)

	keys := sdk.NewKVStoreKeys(paramstypes.StoreKey)

	app.ParamsKeeper = paramskeeper.NewKeeper(
		app.AppCodec(),
		encodingConfig.Amino,
		keys[paramstypes.StoreKey],
		transientKeys[paramstypes.TStoreKey],
	)

	app.ParamsKeeper.Subspace(authtypes.ModuleName)
	app.ParamsKeeper.Subspace(banktypes.ModuleName)
	app.ParamsKeeper.Subspace(stakingtypes.ModuleName)
	app.ParamsKeeper.Subspace(distrtypes.ModuleName)

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

	maccPerms := map[string][]string{
		faucetAccountName:              {authtypes.Minter},
		authtypes.FeeCollectorName:     nil,
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		distrtypes.ModuleName:          nil,
		types.ModuleName:               nil,
	}

	blackListAddrs := map[string]bool{
		authtypes.FeeCollectorName:     true,
		stakingtypes.NotBondedPoolName: true,
		stakingtypes.BondedPoolName:    true,
		distrtypes.ModuleName:          true,
		faucetAccountName:              true,
	}

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		app.AppCodec(),
		app.GetKey(authtypes.ModuleName),
		app.ParamsKeeper.Subspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		maccPerms,
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		app.AppCodec(),
		app.GetKey(types.ModuleName),
		app.AccountKeeper,
		app.ParamsKeeper.Subspace(banktypes.ModuleName),
		blackListAddrs,
	)

	oracle.InitGenesis(ctx, app.OracleKeeper, *types.DefaultGenesisState())

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(app.OracleKeeper))

	totalSupply := sdk.NewCoins(sdk.NewCoin("uumee", InitTokens.MulRaw(int64(len(Addrs)*10))))
	app.BankKeeper.MintCoins(ctx, faucetAccountName, totalSupply)

	s.app = app
	s.ctx = ctx
	s.queryClient = types.NewQueryClient(queryHelper)
}

// Test addresses
var (
	ValPubKeys = simapp.CreateTestPubKeys(5)

	pubKeys = []crypto.PubKey{
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
	}

	Addrs = []sdk.AccAddress{
		sdk.AccAddress(pubKeys[0].Address()),
		sdk.AccAddress(pubKeys[1].Address()),
		sdk.AccAddress(pubKeys[2].Address()),
		sdk.AccAddress(pubKeys[3].Address()),
		sdk.AccAddress(pubKeys[4].Address()),
	}

	ValAddrs = []sdk.ValAddress{
		sdk.ValAddress(pubKeys[0].Address()),
		sdk.ValAddress(pubKeys[1].Address()),
		sdk.ValAddress(pubKeys[2].Address()),
		sdk.ValAddress(pubKeys[3].Address()),
		sdk.ValAddress(pubKeys[4].Address()),
	}

	InitTokens = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	InitCoins  = sdk.NewCoins(sdk.NewCoin(types.UmeeDenom, InitTokens))

	OracleDecPrecision = 8
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

func (s *IntegrationTestSuite) Test_FeederDelegation() {
	app, ctx := s.app, s.ctx

	feederAddr := sdk.AccAddress([]byte("addr________________"))
	feederAcc := app.AccountKeeper.NewAccountWithAddress(ctx, feederAddr)
	app.AccountKeeper.SetAccount(ctx, feederAcc)

	sh := staking.NewHandler(app.StakingKeeper)
	// Validator created
	amt := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	_, err := sh(ctx, NewTestMsgCreateValidator(ValAddrs[0], ValPubKeys[0], amt))
	s.Require().NoError(err)

	s.app.OracleKeeper.SetFeederDelegation(ctx, ValAddrs[0], feederAddr)

	// Should not fail since we set up this feeder
	err = s.app.OracleKeeper.ValidateFeeder(ctx, feederAddr, ValAddrs[0])
	s.Require().NoError(err)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
