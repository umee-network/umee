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

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/x/leverage"
	"github.com/umee-network/umee/v2/x/leverage/fixtures"
	"github.com/umee-network/umee/v2/x/leverage/keeper"
	"github.com/umee-network/umee/v2/x/leverage/types"
)

const (
	initialPower = int64(10000000000)
	umeeDenom    = umeeapp.BondDenom
	atomDenom    = fixtures.AtomDenom
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
	app := umeeapp.Setup(s.T(), false, 1)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
		Time:    time.Unix(0, 0),
	})

	umeeToken := newToken(umeeapp.BondDenom, "UMEE")
	atomIBCToken := newToken(atomDenom, "ATOM")

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

	leverage.InitGenesis(ctx, app.LeverageKeeper, *types.DefaultGenesis())
	s.Require().NoError(app.LeverageKeeper.SetTokenSettings(ctx, umeeToken))
	s.Require().NoError(app.LeverageKeeper.SetTokenSettings(ctx, atomIBCToken))

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(app.LeverageKeeper))

	s.app = app
	s.ctx = ctx
	s.setupAccountCounter = sdkmath.ZeroInt()
	s.queryClient = types.NewQueryClient(queryHelper)
}

// creates a test token with reasonable initial parameters
func newToken(base, symbol string) types.Token {
	return fixtures.Token(base, symbol)
}

// creates a coin with a given base denom and amount
func coin(denom string, amount int64) sdk.Coin {
	return sdk.NewInt64Coin(denom, amount)
}

// newAccount creates a new account for testing, and funds it with any input tokens.
func (s *IntegrationTestSuite) newAccount(funds ...sdk.Coin) sdk.AccAddress {
	app, ctx := s.app, s.ctx

	// create a unique address
	s.setupAccountCounter = s.setupAccountCounter.Add(sdk.OneInt())
	addrStr := fmt.Sprintf("%-20s", "addr"+s.setupAccountCounter.String()+"_______________")
	addr := sdk.AccAddress([]byte(addrStr))

	// register the account in AccountKeeper
	acct := app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
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

/*
// withdraw uTokens from an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) withdraw(addr sdk.AccAddress, uTokens ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	for _, coin := range uTokens {
		_, err := app.LeverageKeeper.Withdraw(ctx, addr, coin)
		require.NoError(err, "withdraw")
	}
}
*/

// collateralize uTokens from an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) collateralize(addr sdk.AccAddress, uTokens ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	for _, coin := range uTokens {
		err := app.LeverageKeeper.Collateralize(ctx, addr, coin)
		require.NoError(err, "collateralize")
	}
}

/*
// decollateralize uTokens from an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) decollateralize(addr sdk.AccAddress, uTokens ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	for _, coin := range uTokens {
		err := app.LeverageKeeper.Decollateralize(ctx, addr, coin)
		require.NoError(err, "decollateralize")
	}
}
*/

// borrow tokens as an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) borrow(addr sdk.AccAddress, coins ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	for _, coin := range coins {
		err := app.LeverageKeeper.Borrow(ctx, addr, coin)
		require.NoError(err, "borrow")
	}
}

/*
// repay tokens as an account and require no errors. Use when setting up leverage scenarios.
func (s *IntegrationTestSuite) repay(addr sdk.AccAddress, coins ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	for _, coin := range coins {
		repaid, err := app.LeverageKeeper.Repay(ctx, addr, coin)
		require.NoError(err, "repay")
		// ensure intended repayment amount was not reduced, as doing so would create a misleading test
		require.Equal(repaid, coin, "repay")
	}
}
*/

// setupAccount executes some common boilerplate before a test, where a user account is given tokens of a given denom,
// may also supply them to receive uTokens, and may also enable those uTokens as collateral and borrow tokens in the same denom.
func (s *IntegrationTestSuite) setupAccount(denom string, mintAmount, supplyAmount, borrowAmount int64, collateral bool) sdk.AccAddress {
	// create a unique address
	s.setupAccountCounter = s.setupAccountCounter.Add(sdk.OneInt())
	addrStr := fmt.Sprintf("%-20s", "addr"+s.setupAccountCounter.String()+"_______________")
	addr := sdk.AccAddress([]byte(addrStr))

	// register the account in AccountKeeper
	acct := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
	s.app.AccountKeeper.SetAccount(s.ctx, acct)

	if mintAmount > 0 {
		// mint and send mintAmount tokens to account
		s.Require().NoError(s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName,
			sdk.NewCoins(sdk.NewInt64Coin(denom, mintAmount)),
		))
		s.Require().NoError(s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, addr,
			sdk.NewCoins(sdk.NewInt64Coin(denom, mintAmount)),
		))
	}

	if supplyAmount > 0 {
		// account supplies supplyAmount tokens and receives uTokens
		uTokens, err := s.app.LeverageKeeper.Supply(s.ctx, addr, sdk.NewInt64Coin(denom, supplyAmount))
		s.Require().NoError(err)
		s.Require().Equal(sdk.NewInt64Coin(types.ToUTokenDenom(denom), supplyAmount), uTokens)
	}

	if collateral {
		// account enables associated uToken as collateral
		collat, err := s.app.LeverageKeeper.ExchangeToken(s.ctx, sdk.NewInt64Coin(denom, supplyAmount))
		s.Require().NoError(err)
		err = s.app.LeverageKeeper.Collateralize(s.ctx, addr, collat)
		s.Require().NoError(err)
	}

	if borrowAmount > 0 {
		// account borrows borrowAmount tokens
		err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(denom, borrowAmount))
		s.Require().NoError(err)
	}

	// return the account addresse
	return addr
}

// initialize the common starting scenario from which borrow and repay tests stem:
// Umee and u/umee are registered assets; a "supplier" account has 9k umee and 1k u/umee;
// the leverage module has 1k umee in its lending pool (module account); and a "bum"
// account has been created with no assets.
func (s *IntegrationTestSuite) initBorrowScenario() (supplier, bum sdk.AccAddress) {
	app, ctx := s.app, s.ctx

	// create an account and address which will represent a supplier
	supplierAddr := sdk.AccAddress([]byte("addr______________01"))
	supplierAcc := app.AccountKeeper.NewAccountWithAddress(ctx, supplierAddr)
	app.AccountKeeper.SetAccount(ctx, supplierAcc)

	// create an account and address which will represent a user with no assets
	bumAddr := sdk.AccAddress([]byte("addr______________02"))
	bumAcc := app.AccountKeeper.NewAccountWithAddress(ctx, bumAddr)
	app.AccountKeeper.SetAccount(ctx, bumAcc)

	// mint and send 10k umee to supplier
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName,
		sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 10000000000)), // 10k umee
	))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, supplierAddr,
		sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 10000000000)), // 10k umee,
	))

	// supplier supplies 1000 umee and receives 1k u/umee
	supplyCoin := sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)
	_, err := s.app.LeverageKeeper.Supply(ctx, supplierAddr, supplyCoin)
	s.Require().NoError(err)

	// supplier enables u/umee as collateral
	collat, err := s.app.LeverageKeeper.ExchangeToken(ctx, supplyCoin)
	s.Require().NoError(err)
	err = s.app.LeverageKeeper.Collateralize(ctx, supplierAddr, collat)
	s.Require().NoError(err)

	// return the account addresses
	return supplierAddr, bumAddr
}
