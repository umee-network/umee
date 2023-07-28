package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeapp "github.com/umee-network/umee/v5/app"
	"github.com/umee-network/umee/v5/x/metoken"
	"github.com/umee-network/umee/v5/x/metoken/keeper"
	"github.com/umee-network/umee/v5/x/metoken/mocks"
)

type KeeperTestSuite struct {
	ctx         sdk.Context
	app         *umeeapp.UmeeApp
	queryClient metoken.QueryClient
	msgServer   metoken.MsgServer

	setupAccountCounter sdkmath.Int
	addrs               []sdk.AccAddress
}

// initTestSuite creates a full keeper with all the external dependencies mocked
func initTestSuite(t *testing.T, registry []metoken.Index, balances []metoken.IndexBalances) *KeeperTestSuite {
	t.Parallel()
	isCheckTx := false
	app := umeeapp.Setup(t)
	ctx := app.NewContext(
		isCheckTx, tmproto.Header{
			ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
			Height:  9,
		},
	).WithBlockTime(time.Now())

	oracleMock := mocks.NewMockOracleKeeper()
	oracleMock.AllMedianPricesFunc.SetDefaultHook(mocks.ValidPricesFunc())

	kb := keeper.NewKeeperBuilder(
		app.AppCodec(),
		app.GetKey(metoken.ModuleName),
		app.BankKeeper,
		app.LeverageKeeper,
		oracleMock,
	)
	app.MetokenKeeperB = kb

	genState := metoken.DefaultGenesisState()
	genState.Registry = registry
	genState.Balances = balances
	kb.Keeper(&ctx).InitGenesis(*genState)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	metoken.RegisterQueryServer(queryHelper, keeper.NewQuerier(app.MetokenKeeperB))

	require.NoError(
		t,
		app.LeverageKeeper.SetTokenSettings(
			ctx,
			mocks.ValidToken(mocks.USDTBaseDenom, mocks.USDTSymbolDenom, 6),
		),
	)
	require.NoError(
		t,
		app.LeverageKeeper.SetTokenSettings(
			ctx,
			mocks.ValidToken(mocks.USDCBaseDenom, mocks.USDCSymbolDenom, 6),
		),
	)
	require.NoError(
		t,
		app.LeverageKeeper.SetTokenSettings(
			ctx,
			mocks.ValidToken(mocks.ISTBaseDenom, mocks.ISTSymbolDenom, 6),
		),
	)
	require.NoError(
		t,
		app.LeverageKeeper.SetTokenSettings(
			ctx,
			mocks.ValidToken(mocks.WBTCBaseDenom, mocks.WBTCSymbolDenom, 8),
		),
	)
	require.NoError(
		t,
		app.LeverageKeeper.SetTokenSettings(
			ctx,
			mocks.ValidToken(mocks.ETHBaseDenom, mocks.ETHSymbolDenom, 18),
		),
	)

	return &KeeperTestSuite{
		ctx:                 ctx,
		app:                 app,
		queryClient:         metoken.NewQueryClient(queryHelper),
		msgServer:           keeper.NewMsgServerImpl(app.MetokenKeeperB),
		setupAccountCounter: sdkmath.ZeroInt(),
		addrs:               umeeapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(3000000)),
	}
}

// newAccount creates a new account for testing, and funds it with any input tokens.
func (s *KeeperTestSuite) newAccount(t *testing.T, funds ...sdk.Coin) sdk.AccAddress {
	app, ctx := s.app, s.ctx

	// create a unique address
	s.setupAccountCounter = s.setupAccountCounter.Add(sdk.OneInt())
	addrStr := fmt.Sprintf("%-20s", "addr"+s.setupAccountCounter.String()+"_______________")
	addr := sdk.AccAddress([]byte(addrStr))

	// register the account in AccountKeeper
	acct := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acct)

	s.fundAccount(t, addr, funds...)

	return addr
}

// fundAccount mints and sends tokens to an account for testing.
func (s *KeeperTestSuite) fundAccount(t *testing.T, addr sdk.AccAddress, funds ...sdk.Coin) {
	app, ctx := s.app, s.ctx

	coins := sdk.NewCoins(funds...)
	if !coins.IsZero() {
		// mint and send tokens to account
		require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
		require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins))
	}
}
