package keeper_test

import (
	"fmt"
	"testing"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/umee-network/umee/v6/x/metoken"

	sdkmath "cosmossdk.io/math"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"

	umeeapp "github.com/umee-network/umee/v6/app"
	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/x/leverage"
	"github.com/umee-network/umee/v6/x/leverage/fixtures"
	"github.com/umee-network/umee/v6/x/leverage/keeper"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

const (
	umeeDenom      = appparams.BondDenom
	atomDenom      = fixtures.AtomDenom
	daiDenom       = fixtures.DaiDenom
	pumpDenom      = "upump"
	dumpDenom      = "udump"
	stableDenom    = "stable"
	pairedDenom    = "upaired"
	outageDenom    = "uoutage"
	newStableDenom = "unewStableDenom"
)

var leverage_initial_registry_length = 0

type IntegrationTestSuite struct {
	suite.Suite

	ctx                 sdk.Context
	app                 *umeeapp.UmeeApp
	tk                  keeper.TestKeeper
	queryClient         types.QueryClient
	setupAccountCounter sdkmath.Int
	addrs               []sdk.AccAddress
	msgSrvr             types.MsgServer

	mockOracle *mockOracleKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	require := s.Require()
	app := umeeapp.Setup(s.T())
	ctx := app.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
		Time:    time.Unix(0, 0),
	})

	s.mockOracle = newMockOracleKeeper()

	// we only override the Leverage keeper so we can supply a custom mock oracle
	k, tk := keeper.NewTestKeeper(
		app.AppCodec(),
		app.GetKey(types.ModuleName),
		app.BankKeeper,
		s.mockOracle,
		true,
	)

	s.tk = tk
	app.LeverageKeeper = k
	// since keeper was overridden, we need to set these hooks again
	app.LeverageKeeper.SetTokenHooks()
	app.LeverageKeeper.SetBondHooks() // TODO: add a mock (or real) incentive module here

	// override DefaultGenesis token registry with fixtures.Token
	leverage.InitGenesis(ctx, app.LeverageKeeper, *types.DefaultGenesis())
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, newToken(appparams.BondDenom, "UMEE", 6)))
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, newToken(atomDenom, "ATOM", 6)))
	daiToken := newToken(daiDenom, "DAI", 18) // high exponent token will need bigger maxSupply for testing
	daiToken.MaxSupply = daiToken.MaxSupply.Mul(sdk.NewInt(1_000_000_000_000))
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, daiToken))
	// additional tokens for historacle testing
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, newToken(dumpDenom, "DUMP", 6)))
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, newToken(pumpDenom, "PUMP", 6)))
	// additional token for special pairs
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, newToken(pairedDenom, "PAIRED", 6)))
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, newToken(outageDenom, "OUTAGE", 6)))
	// additional tokens for borrow factor testing
	stable := newToken(stableDenom, "STABLE", 6)
	stable.CollateralWeight = sdk.MustNewDecFromStr("0.8")
	stable.LiquidationThreshold = sdk.MustNewDecFromStr("0.9")
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, stable))
	// additional token for withdraw and borrow limited by meToken testing
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, newToken(newStableDenom, "NEWSTABLE", 6)))

	// set the initial token registry length used in update registry tests
	leverage_initial_registry_length = len(app.LeverageKeeper.GetAllRegisteredTokens(ctx))

	// override DefaultGenesis params with some special asset pairs
	app.LeverageKeeper.SetSpecialAssetPair(
		ctx,
		types.SpecialAssetPair{
			Collateral:           pairedDenom,
			Borrow:               daiDenom,
			CollateralWeight:     sdk.MustNewDecFromStr("0.5"),
			LiquidationThreshold: sdk.MustNewDecFromStr("0.75"),
		},
	)

	// override DefaultGenesis params with fixtures.Params
	app.LeverageKeeper.SetParams(ctx, fixtures.Params())

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(app.LeverageKeeper))

	s.app = app
	s.ctx = ctx
	s.setupAccountCounter = sdkmath.ZeroInt()
	s.queryClient = types.NewQueryClient(queryHelper)
	s.addrs = umeeapp.AddTestAddrsIncremental(app, s.ctx, 1, sdk.NewInt(3000000))
	s.msgSrvr = keeper.NewMsgServerImpl(s.app.LeverageKeeper)
}

// requireEqualCoins compares two sdk.Coins in such a way that sdk.Coins(nil) == sdk.Coins([]sdk.Coin{})
func (s *IntegrationTestSuite) requireEqualCoins(coinsA, coinsB sdk.Coins, msgAndArgs ...interface{}) {
	s.Require().Equal(
		sdk.NewCoins(coinsA...),
		sdk.NewCoins(coinsB...),
		msgAndArgs...,
	)
}

// newToken creates a test token with reasonable initial parameters
func newToken(base, symbol string, exponent uint32) types.Token {
	return fixtures.Token(base, symbol, exponent)
}

// registerToken adds or updates a token in the token registry and requires no error.
func (s *IntegrationTestSuite) registerToken(token types.Token) {
	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, token))
}

// newAccount creates a new account for testing, and funds it with any input tokens.
func (s *IntegrationTestSuite) newAccount(funds ...sdk.Coin) sdk.AccAddress {
	app, ctx := s.app, s.ctx

	// create a unique address
	s.setupAccountCounter = s.setupAccountCounter.Add(sdk.OneInt())
	addrStr := fmt.Sprintf("%-20s", "addr"+s.setupAccountCounter.String()+"_______________")
	addr := sdk.AccAddress([]byte(addrStr))

	// register the account in AccountKeeper
	acct := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acct)

	s.fundAccount(addr, funds...)

	return addr
}

// setupMeTokenAccount creates new meToken account for testing, and funds it with any input tokens.
func (s *IntegrationTestSuite) setupMeTokenAccount(funds ...sdk.Coin) sdk.AccAddress {
	app, ctx, require := s.app, s.ctx, s.Require()

	s.setupAccountCounter = s.setupAccountCounter.Add(sdk.OneInt())
	meTokenAddr := authtypes.NewModuleAddress(metoken.ModuleName)

	require.NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, funds))
	require.NoError(app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, metoken.ModuleName, funds))

	return meTokenAddr
}

// fundAccount mints and sends tokens to an account for testing.
func (s *IntegrationTestSuite) fundAccount(addr sdk.AccAddress, funds ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	coins := sdk.NewCoins(funds...)
	if !coins.IsZero() {
		// mint and send tokens to account
		require.NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
		require.NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins))
	}
}

// supply tokens from an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) supply(addr sdk.AccAddress, coins ...sdk.Coin) {
	srv, ctx, require := s.msgSrvr, s.ctx, s.Require()

	for _, coin := range coins {
		msg := &types.MsgSupply{
			Supplier: addr.String(),
			Asset:    coin,
		}
		_, err := srv.Supply(ctx, msg)
		require.NoError(err, "supply")
	}
}

// supplyFromModule tokens from a module account and require no errors.
// Use when setting up leverage scenarios for meToken.
func (s *IntegrationTestSuite) supplyFromModule(moduleName string, coins ...sdk.Coin) {
	l, ctx, require := s.app.LeverageKeeper, s.ctx, s.Require()

	for _, coin := range coins {
		_, _, err := l.SupplyFromModule(ctx, moduleName, coin)
		require.NoError(err, "supplyFromModule")
	}
}

// withdraw utokens from an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) withdraw(addr sdk.AccAddress, coins ...sdk.Coin) {
	srv, ctx, require := s.msgSrvr, s.ctx, s.Require()

	for _, coin := range coins {
		msg := &types.MsgWithdraw{
			Supplier: addr.String(),
			Asset:    coin,
		}
		_, err := srv.Withdraw(ctx, msg)
		require.NoError(err, "withdraw")
	}
}

// collateralize uTokens from an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) collateralize(addr sdk.AccAddress, uTokens ...sdk.Coin) {
	srv, ctx, require := s.msgSrvr, s.ctx, s.Require()

	for _, coin := range uTokens {
		msg := &types.MsgCollateralize{
			Borrower: addr.String(),
			Asset:    coin,
		}
		_, err := srv.Collateralize(ctx, msg)
		require.NoError(err, "collateralize")
	}
}

// decollateralize uTokens from an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) decollateralize(addr sdk.AccAddress, uTokens ...sdk.Coin) {
	srv, ctx, require := s.msgSrvr, s.ctx, s.Require()

	for _, coin := range uTokens {
		msg := &types.MsgDecollateralize{
			Borrower: addr.String(),
			Asset:    coin,
		}
		_, err := srv.Decollateralize(ctx, msg)
		require.NoError(err, "decollateralize")
	}
}

// borrow tokens as an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) borrow(addr sdk.AccAddress, coins ...sdk.Coin) {
	srv, ctx, require := s.msgSrvr, s.ctx, s.Require()

	for _, coin := range coins {
		msg := &types.MsgBorrow{
			Borrower: addr.String(),
			Asset:    coin,
		}
		_, err := srv.Borrow(ctx, msg)
		require.NoError(err, "borrow")
	}
}

// forceBorrow artificially borrows tokens with an account, ignoring collateral, to set up liquidation scenarios.
// this does not alter uToken exchange rates as artificially accruing interest would.
func (s *IntegrationTestSuite) forceBorrow(addr sdk.AccAddress, coins ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	for _, coin := range coins {
		err := app.LeverageKeeper.Borrow(ctx, addr, coin)
		require.NoError(err, "borrow")
	}
}

// setReserves artificially sets reserves of one or more tokens to given values
func (s *IntegrationTestSuite) setReserves(coins ...sdk.Coin) {
	ctx, require := s.ctx, s.Require()

	for _, coin := range coins {
		err := s.tk.SetReserveAmount(ctx, coin)
		require.NoError(err, "setReserves")
	}
}

// checkInvariants is used during other tests to quickly test all invariants,
// including the inefficient ones we do not run in production
func (s *IntegrationTestSuite) checkInvariants(msg string) {
	app, ctx, require := s.app, s.ctx, s.Require()

	invariants := []sdk.Invariant{
		keeper.InefficientBorrowAmountInvariant(app.LeverageKeeper),
		keeper.InefficientCollateralAmountInvariant(app.LeverageKeeper),
		keeper.ReserveAmountInvariant(app.LeverageKeeper),
		keeper.InterestScalarsInvariant(app.LeverageKeeper),
		keeper.ExchangeRatesInvariant(app.LeverageKeeper),
		keeper.SupplyAPYInvariant(app.LeverageKeeper),
		keeper.BorrowAPYInvariant(app.LeverageKeeper),
	}

	for _, inv := range invariants {
		desc, broken := inv(ctx)
		require.False(broken, msg, "desc", desc)
	}
}
