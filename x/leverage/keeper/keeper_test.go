package keeper_test

import (
	"fmt"
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

// initialize the common starting scenario from which borrow and repay tests stem:
// Umee and u/umee are registered assets; a "lender" account has 9k umee and 1k u/umee;
// the leverage module has 1k umee in its lending pool (module account); and a "bum"
// account has been created with no assets.
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
	suite.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName,
		sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 10000000000)), // 10k umee
	))
	suite.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr,
		sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 10000000000)), // 10k umee,
	))

	// lender lends 1000 umee and receives 1k u/umee
	err := app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000))
	suite.Require().NoError(err)

	// return the account addresses
	return lenderAddr, bumAddr
}

func (suite *IntegrationTestSuite) TestBorrowAsset_Invalid() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender attempts to borrow 200 u/umee, fails because uTokens cannot be borrowed
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 200000000))
	suite.Require().Error(err)

	// lender attempts to borrow 200 abcd, fails because "abcd" is not a valid denom
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin("uabcd", 200000000))
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
	err := app.LeverageKeeper.BorrowAsset(ctx,bumAddr,tCoin("umee",200))
	suite.Require().Error(err)
}
*/

func (suite *IntegrationTestSuite) TestBorrowAsset_InsufficientLendingPool() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// Any user from the init scenario can perform this test, because it errors on module balance

	// lender attempts to borrow 20000 umee, fails because of insufficient module account balance
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000000))
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestRepayAsset_Invalid() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// Any user from the init scenario can be used for this test.

	// lender attempts to repay 200 abcd, fails because "abcd" is not an accepted asset type
	err := app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin("uabcd", 200000000))
	suite.Require().Error(err)

	// lender attempts to repay 200 u/umee, fails because utokens are not loanable assets
	err = app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 200000000))
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestBorrowAsset_Valid() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 200 umee
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	suite.Require().NoError(err)

	// verify lender's new loan amount in the correct denom (200 umee)
	loanBalance := app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))

	// verify lender's total loan balance (sdk.Coins) is also 200 umee (no other coins present)
	totalLoanBalance, err := app.LeverageKeeper.GetBorrowerBorrows(ctx, lenderAddr)
	suite.Require().Equal(totalLoanBalance, sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 200000000)))

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan = 9200 umee)
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9200000000))

	// verify lender's uToken balance remains at 1000 u/umee from initial conditions
	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	suite.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))
}

func (suite *IntegrationTestSuite) TestRepayAsset_Valid() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 200 umee
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	suite.Require().NoError(err)

	// lender repays 80 umee
	err = app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 80000000))
	suite.Require().NoError(err)

	// verify lender's new loan amount (120 umee)
	loanBalance := app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 120000000))

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan - 80 repaid = 9120 umee)
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9120000000))

	// verify lender's uToken balance remains at 1000 u/umee from initial conditions
	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	suite.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))

	// lender repays 120 umee (loan repaid in full)
	err = app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 120000000))
	suite.Require().NoError(err)

	// verify lender's new loan amount in the correct denom (zero)
	loanBalance = app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 0))

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan - 200 repaid = 9000 umee)
	tokenBalance = app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9000000000))

	// verify lender's uToken balance remains at 1000 u/umee from initial conditions
	uTokenBalance = app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	suite.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))
}

func (suite *IntegrationTestSuite) TestRepayAsset_Overpay() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 200 umee
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	suite.Require().NoError(err)

	// lender repays 300 umee - should automatically reduce to 200 (the loan amount) and succeed
	err = app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 300000000))
	suite.Require().NoError(err)

	// verify lender's new loan amount is 0 umee
	loanBalance := app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 0))

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan - 200 repaid = 9000 umee)
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9000000000))

	// verify lender's uToken balance remains at 1k u/umee from initial conditions
	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	suite.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))

	// lender repays 50 umee - this time it fails because the loan no longer exists
	err = app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 50000000))
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestSetCollateral_Valid() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

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
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// Any user from the starting scenario can be used, since they are only toggling
	// collateral settings.

	// lender disables u/abcd as collateral - fails because "u/abcd" is not a recognized uToken
	err := app.LeverageKeeper.SetCollateralSetting(ctx, lenderAddr, "u/abcd", false)
	suite.Require().Error(err)

	// lender disables uumee as collateral - fails because "uumee" is an asset, not a uToken
	err = app.LeverageKeeper.SetCollateralSetting(ctx, lenderAddr, "uumee", false)
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestGetCollateral_Invalid() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

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
