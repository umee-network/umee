package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeapp "github.com/umee-network/umee/v3/app"
	appparams "github.com/umee-network/umee/v3/app/params"
	"github.com/umee-network/umee/v3/x/leverage"
	"github.com/umee-network/umee/v3/x/leverage/fixtures"
	"github.com/umee-network/umee/v3/x/leverage/keeper"
	"github.com/umee-network/umee/v3/x/leverage/types"
)

const (
	umeeDenom = appparams.BondDenom
	atomDenom = fixtures.AtomDenom
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx                 sdk.Context
	app                 *umeeapp.UmeeApp
	tk                  keeper.TestKeeper
	queryClient         types.QueryClient
	setupAccountCounter sdkmath.Int
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	require := s.Require()
	app := umeeapp.Setup(s.T(), false, 1)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
		Time:    time.Unix(0, 0),
	})

	// Enable liquidation queries for testing
	keeper.EnableLiquidator = "true"

	// we only override the Leverage keeper so we can supply a custom mock oracle
	k, tk := keeper.NewTestKeeper(
		s.Require(),
		app.AppCodec(),
		app.GetKey(types.ModuleName),
		app.GetSubspace(types.ModuleName),
		app.BankKeeper,
		newMockOracleKeeper(),
	)

	s.tk = tk
	app.LeverageKeeper = k
	app.LeverageKeeper = *app.LeverageKeeper.SetHooks(types.NewMultiHooks())

	// override DefaultGenesis token registry with fixtures.Token
	leverage.InitGenesis(ctx, app.LeverageKeeper, *types.DefaultGenesis())
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, newToken(appparams.BondDenom, "UMEE")))
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, newToken(atomDenom, "ATOM")))

	// override DefaultGenesis params with fixtures.Params
	app.LeverageKeeper.SetParams(ctx, fixtures.Params())

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(app.LeverageKeeper))

	s.app = app
	s.ctx = ctx
	s.setupAccountCounter = sdkmath.ZeroInt()
	s.queryClient = types.NewQueryClient(queryHelper)
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
func newToken(base, symbol string) types.Token {
	return fixtures.Token(base, symbol)
}

// coin creates a coin with a given base denom and amount
func coin(denom string, amount int64) sdk.Coin {
	return sdk.NewInt64Coin(denom, amount)
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
	app, ctx, require := s.app, s.ctx, s.Require()

	for _, coin := range coins {
		_, err := app.LeverageKeeper.Supply(ctx, addr, coin)
		require.NoError(err, "supply")
	}
}

// withdraw utokens from an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) withdraw(addr sdk.AccAddress, coins ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	for _, coin := range coins {
		_, err := app.LeverageKeeper.Withdraw(ctx, addr, coin)
		require.NoError(err, "withdraw")
	}
}

// collateralize uTokens from an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) collateralize(addr sdk.AccAddress, uTokens ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	for _, coin := range uTokens {
		err := app.LeverageKeeper.Collateralize(ctx, addr, coin)
		require.NoError(err, "collateralize")
	}
}

// decollateralize uTokens from an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) decollateralize(addr sdk.AccAddress, uTokens ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	for _, coin := range uTokens {
		err := app.LeverageKeeper.Decollateralize(ctx, addr, coin)
		require.NoError(err, "decollateralize")
	}
}

// borrow tokens as an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) borrow(addr sdk.AccAddress, coins ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	for _, coin := range coins {
		err := app.LeverageKeeper.Borrow(ctx, addr, coin)
		require.NoError(err, "borrow")
	}
}

// forceBorrow artificially borrows tokens with an account, ignoring collateral, to set up liquidation scenarios.
// this does not alter uToken exchange rates as artificially accruing interest would.
func (s *IntegrationTestSuite) forceBorrow(addr sdk.AccAddress, coins ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	for _, coin := range coins {
		borrowed := s.tk.GetBorrow(ctx, addr, coin.Denom)
		err := s.tk.SetBorrow(ctx, addr, borrowed.Add(coin))
		require.NoError(err, "forceBorrow")
	}

	err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, coins)
	require.NoError(err, "forceBorroww")
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
