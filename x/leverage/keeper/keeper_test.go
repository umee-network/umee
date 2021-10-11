package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeapp "github.com/umee-network/umee/app"
	"github.com/umee-network/umee/x/leverage/types"
)

const (
	initialPower = int64(10000000000)
)

var (
	initTokens = sdk.TokensFromConsensusPower(initialPower, sdk.DefaultPowerReduction)
	initCoins  = sdk.NewCoins(sdk.NewCoin(umeeapp.BondDenom, initTokens))
)

// makes sdk.Coin for testing (easier to read)
func tCoin(denom string, amount int) sdk.Coin {
	// adds "u" to denom (bypassing "u/" prefix if found) and multiplies amount by 10^6
	if strings.HasPrefix(denom, "u/") {
		denom = "u/u" + strings.TrimPrefix(denom, "u/")
	} else {
		denom = "u" + denom
	}
	return sdk.NewCoin(denom, sdk.NewInt(int64(amount*1000000)))
	// coin("umee", 900) = sdk.NewInt64Coin("uumee", 900000000)
	// coin("u/umee", 900) = sdk.NewInt64Coin("u/uumee", 900000000)
}

// makes sdk.Coins (with one element) for testing
func tCoins(denom string, amount int) sdk.Coins {
	return sdk.NewCoins(tCoin(denom, amount))
}

type IntegrationTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *umeeapp.UmeeApp
}

func (suite *IntegrationTestSuite) SetupTest() {
	app := umeeapp.Setup(suite.T(), false, 1)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})

	suite.app = app
	suite.ctx = ctx
}

func (suite *IntegrationTestSuite) TestLendAsset_InvalidAsset() {
	app, ctx := suite.app, suite.ctx

	lenderAddr := sdk.AccAddress([]byte("addr________________"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)

	// mint and send coins
	suite.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	suite.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr, initCoins))

	// lending should fail as we have not set what tokens can be lent
	err := app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestLendAsset_Valid() {
	app, ctx := suite.app, suite.ctx

	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)

	lenderAddr := sdk.AccAddress([]byte("addr________________"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)

	// mint and send coins
	suite.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	suite.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr, initCoins))

	// lend asset
	err := app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	suite.Require().NoError(err)

	// verify the total supply of the minted uToken
	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)
	supply := app.LeverageKeeper.TotalUTokenSupply(ctx, uTokenDenom)
	expected := sdk.NewInt64Coin(uTokenDenom, 1000000000) // 1k u/umee
	suite.Require().Equal(expected, supply)

	// verify the balance of the lender's balances
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(initTokens.Sub(sdk.NewInt(1000000000)), tokenBalance.Amount)

	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, uTokenDenom)
	suite.Require().Equal(int64(1000000000), uTokenBalance.Amount.Int64())
}

func (suite *IntegrationTestSuite) TestWithdrawAsset_Valid() {
	app, ctx := suite.app, suite.ctx

	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)

	lenderAddr := sdk.AccAddress([]byte("addr________________"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)

	// mint and send coins
	suite.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	suite.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr, initCoins))

	// lend asset
	err := app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	suite.Require().NoError(err)

	// verify the total supply of the minted uToken
	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)
	supply := app.LeverageKeeper.TotalUTokenSupply(ctx, uTokenDenom)
	expected := sdk.NewInt64Coin(uTokenDenom, 1000000000) // 1k umee
	suite.Require().Equal(expected, supply)

	// withdraw the total amount of assets lent
	uToken := expected
	err = app.LeverageKeeper.WithdrawAsset(ctx, lenderAddr, uToken)
	suite.Require().NoError(err)

	// verify total supply of the uTokens
	supply = app.LeverageKeeper.TotalUTokenSupply(ctx, uTokenDenom)
	suite.Require().Equal(int64(0), supply.Amount.Int64())

	// verify the balance of the lender's balances
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(initTokens, tokenBalance.Amount)

	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, uTokenDenom)
	suite.Require().Equal(int64(0), uTokenBalance.Amount.Int64())
}

// Helper function: Initialize the common starting scenario from which borrow and repay tests stem:
func (suite *IntegrationTestSuite) initBorrowScenario() (lender, bum sdk.AccAddress) {
	app, ctx := suite.app, suite.ctx
	// register uumee and u/uumee as an accepted asset+utoken pair
	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)
	// create an account and address which will represent a lender
	lenderAddr := sdk.AccAddress([]byte("addr______________00"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)
	// create an account and address which will represent a user with no assets
	bumAddr := sdk.AccAddress([]byte("addr______________02"))
	bumAcc := app.AccountKeeper.NewAccountWithAddress(ctx, bumAddr)
	app.AccountKeeper.SetAccount(ctx, bumAcc)
	// mint and send 10k umee to lender
	suite.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, tCoins("umee", 10000)))
	suite.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr, tCoins("umee", 10000)))
	// lender lends 1k umee and receives 1k u/umee
	err := app.LeverageKeeper.LendAsset(ctx, lenderAddr, tCoin("umee", 1000))
	suite.Require().NoError(err)
	// return the three account addresses
	return lenderAddr, bumAddr
	// The starting scenario is thus:
	// - umee and u/umee are accepted assets
	// - a "lender" user has 9k umee and 1k u/umee
	// - the leverage module has 1k umee due to lender's lending
	// - a "bum" user has an address but no assets
}

func (suite *IntegrationTestSuite) TestBorrowAsset_Invalid() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender attempts to borrow 200 u/umee, fails because uTokens cannot be borrowed
	err := app.LeverageKeeper.BorrowAsset(ctx,
		lenderAddr,
		tCoin("u/umee", 200),
	)
	suite.Require().Error(err)

	// lender attempts to borrow 200 abcd, fails because "abcd" is not a valid denom
	err = app.LeverageKeeper.BorrowAsset(ctx,
		lenderAddr,
		tCoin("abcd", 200),
	)
	suite.Require().Error(err)
}

/*
// TODO: Ready to be commented in once borrowing limits exist.
func (suite *IntegrationTestSuite) TestBorrowAsset_InsufficientCollateral() {
	_, bumAddr := suite.initBorrowScenario() // create initial conditions
	app, ctx := suite.app, suite.ctx            // get ctx after init

	// The "bum" user from the init scenario is being used because it
	// possesses no assets or collateral.

	// bum attempts to borrow 200 umee, fails because of insufficient collateral
	err := app.LeverageKeeper.BorrowAsset(ctx,
		bumAddr,
		tCoin("umee",200),
	)
	suite.Require().Error(err)
}
*/

func (suite *IntegrationTestSuite) TestBorrowAsset_InsufficientLendingPool() {
	lenderAddr, _ := suite.initBorrowScenario() // create initial conditions
	app, ctx := suite.app, suite.ctx            // get ctx after init

	// Any user from the init scenario can perform this test, because it errors on module balance

	// lender attempts to borrow 20k umee, fails because of insufficient module account balance
	err := app.LeverageKeeper.BorrowAsset(ctx,
		lenderAddr,
		tCoin("umee", 20000),
	)
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestRepayAsset_Invalid() {
	lenderAddr, _ := suite.initBorrowScenario() // create initial conditions
	app, ctx := suite.app, suite.ctx            // get ctx after init

	// Any user from the init scenario can be used for this test.

	// lender attempts to repay 200 abcd, fails because "abcd" is not an accepted asset type
	err := app.LeverageKeeper.RepayAsset(ctx,
		lenderAddr,
		tCoin("abcd", 200),
	)
	suite.Require().Error(err)

	// lender attempts to repay 200 u/umee, fails because utokens are not loanable assets
	err = app.LeverageKeeper.RepayAsset(ctx,
		lenderAddr,
		tCoin("u/umee", 200),
	)
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestBorrowAsset_Valid() {
	lenderAddr, _ := suite.initBorrowScenario() // create initial conditions
	app, ctx := suite.app, suite.ctx            // get ctx after init

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 200 umee
	err := app.LeverageKeeper.BorrowAsset(ctx,
		lenderAddr,
		tCoin("umee", 200),
	)
	suite.Require().NoError(err)

	// verify lender's new loan amount in the correct denom
	loanBalance := app.LeverageKeeper.GetLoan(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(loanBalance, tCoin("umee", 200))

	// verify lender's total loan balance (sdk.Coins) is also correct (no other coins present)
	totalLoanBalance, err := app.LeverageKeeper.GetAllLoans(ctx, lenderAddr)
	suite.Require().Equal(totalLoanBalance, tCoins("umee", 200))

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan = 9200)
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(tCoin("umee", 9200), tokenBalance)

	// verify lender's uToken balance remains at 1k from initial conditions
	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	suite.Require().Equal(tCoin("u/umee", 1000), uTokenBalance)
}

func (suite *IntegrationTestSuite) TestRepayAsset_Valid() {
	lenderAddr, _ := suite.initBorrowScenario() // create initial conditions
	app, ctx := suite.app, suite.ctx            // get ctx after init

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 200 umee
	err := app.LeverageKeeper.BorrowAsset(ctx,
		lenderAddr,
		tCoin("umee", 200),
	)
	suite.Require().NoError(err)

	// lender repays 80 umee
	err = app.LeverageKeeper.RepayAsset(ctx,
		lenderAddr,
		tCoin("umee", 80),
	)
	suite.Require().NoError(err)

	// verify lender's new loan amount in the correct denom
	loanBalance := app.LeverageKeeper.GetLoan(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(loanBalance, tCoin("umee", 120))

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan - 80 repaid = 9120)
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(tCoin("umee", 9120), tokenBalance)

	// verify lender's uToken balance remains at 1k from initial conditions
	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	suite.Require().Equal(tCoin("u/umee", 1000), uTokenBalance)

	// lender repays 120 umee
	err = app.LeverageKeeper.RepayAsset(ctx,
		lenderAddr,
		tCoin("umee", 120),
	)
	suite.Require().NoError(err)

	// verify lender's new loan amount in the correct denom (zero, because fully repaid)
	loanBalance = app.LeverageKeeper.GetLoan(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(loanBalance, tCoin("umee", 0))

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan - 200 repaid = 9000)
	tokenBalance = app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(tCoin("umee", 9000), tokenBalance)

	// verify lender's uToken balance remains at 1k from initial conditions
	uTokenBalance = app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	suite.Require().Equal(tCoin("u/umee", 1000), uTokenBalance)
}

func (suite *IntegrationTestSuite) TestRepayAsset_Overpay() {
	lenderAddr, _ := suite.initBorrowScenario() // create initial conditions
	app, ctx := suite.app, suite.ctx            // get ctx after init

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 200 umee
	err := app.LeverageKeeper.BorrowAsset(ctx,
		lenderAddr,
		tCoin("umee", 200),
	)
	suite.Require().NoError(err)

	// lender repays 300 umee - should automatically be reduced to 200 (the loan amount) and succeed
	err = app.LeverageKeeper.RepayAsset(ctx,
		lenderAddr,
		tCoin("umee", 300),
	)
	suite.Require().NoError(err)

	// verify lender's new loan amount is zero
	loanBalance := app.LeverageKeeper.GetLoan(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(loanBalance, tCoin("umee", 0))

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan - 200 repaid = 9000)
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(tokenBalance, tCoin("umee", 9000))

	// verify lender's uToken balance remains at 1k from initial conditions
	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	suite.Require().Equal(uTokenBalance, tCoin("u/umee", 1000))

	// lender repays 50 umee - this time it fails because the loan no longer exists
	err = app.LeverageKeeper.RepayAsset(ctx,
		lenderAddr,
		tCoin("umee", 50),
	)
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestSetCollateral_Valid() {
	lenderAddr, _ := suite.initBorrowScenario() // create initial conditions
	app, ctx := suite.app, suite.ctx            // get ctx after init

	// Any user from the starting scenario can be used, since they are only toggling
	// collateral settings.

	// lender disables u/umee as collateral
	err := app.LeverageKeeper.SetCollateralSetting(ctx,
		lenderAddr,
		"u/uumee",
		false,
	)
	suite.Require().NoError(err)
	enabled := app.LeverageKeeper.GetCollateralSetting(ctx,
		lenderAddr,
		"u/uumee",
	)
	suite.Require().Equal(enabled, false)

	// lender enables u/umee as collateral
	err = app.LeverageKeeper.SetCollateralSetting(ctx,
		lenderAddr,
		"u/uumee",
		true,
	)
	suite.Require().NoError(err)
	enabled = app.LeverageKeeper.GetCollateralSetting(ctx,
		lenderAddr,
		"u/uumee",
	)
	suite.Require().Equal(enabled, true)
}

func (suite *IntegrationTestSuite) TestSetCollateral_Invalid() {
	lenderAddr, _ := suite.initBorrowScenario() // create initial conditions
	app, ctx := suite.app, suite.ctx            // get ctx after init

	// Any user from the starting scenario can be used, since they are only toggling
	// collateral settings.

	// lender disables u/abcd as collateral - fails because "u/abcd" is not a recognized uToken
	err := app.LeverageKeeper.SetCollateralSetting(ctx,
		lenderAddr,
		"u/abcd",
		false,
	)
	suite.Require().Error(err)

	// lender disables uumee as collateral - fails because "uumee" is an asset, not a uToken
	err = app.LeverageKeeper.SetCollateralSetting(ctx,
		lenderAddr,
		"uumee",
		false,
	)
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestGetCollateral_Invalid() {
	lenderAddr, _ := suite.initBorrowScenario() // create initial conditions
	app, ctx := suite.app, suite.ctx            // get ctx after init

	// Any user from the starting scenario can be used, since we are only viewing
	// collateral settings.

	// Regular assets always return false, because only uTokens can be collateral
	enabled := app.LeverageKeeper.GetCollateralSetting(ctx,
		lenderAddr,
		"uumee",
	)
	suite.Require().Equal(enabled, false)

	// Invalid or unrecognized assets always return false
	enabled = app.LeverageKeeper.GetCollateralSetting(ctx,
		lenderAddr,
		"abcd",
	)
	suite.Require().Equal(enabled, false)
}

func (suite *IntegrationTestSuite) TestParams() {
	app, ctx := suite.app, suite.ctx

	params := types.DefaultParams()
	app.LeverageKeeper.SetParams(ctx, params)

	got := app.LeverageKeeper.GetParams(ctx)
	suite.Require().Equal(params, got)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
