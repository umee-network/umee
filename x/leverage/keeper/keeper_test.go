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

	uumee := types.Token{
		BaseDenom:           umeeapp.BondDenom,
		ReserveFactor:       sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:    sdk.MustNewDecFromStr("0.1"),
		BaseBorrowRate:      sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:      sdk.MustNewDecFromStr("0.2"),
		MaxBorrowRate:       sdk.MustNewDecFromStr("1.0"),
		KinkUtilizationRate: sdk.MustNewDecFromStr("0.8"),
	}
	// At the moment, SetRegisteredToken must be followed separately by SetTokenDenom
	// to complete token registration. Therefore, this line does not break the InvalidAsset tests
	// which require 'uumee' to be unregistered.
	app.LeverageKeeper.SetRegisteredToken(ctx, uumee)

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

	// verify the lender's balances
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
	expected := sdk.NewInt64Coin(uTokenDenom, 1000000000) // 1k u/umee
	suite.Require().Equal(expected, supply)

	// withdraw the total amount of assets lent
	uToken := expected
	err = app.LeverageKeeper.WithdrawAsset(ctx, lenderAddr, uToken)
	suite.Require().NoError(err)

	// verify total supply of the uTokens
	supply = app.LeverageKeeper.TotalUTokenSupply(ctx, uTokenDenom)
	suite.Require().Equal(int64(0), supply.Amount.Int64())

	// verify the lender's balances
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(initTokens, tokenBalance.Amount)

	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, uTokenDenom)
	suite.Require().Equal(int64(0), uTokenBalance.Amount.Int64())
}

func (suite *IntegrationTestSuite) TestWithdrawAsset_WithExchangeRate() {
	app, ctx := suite.app, suite.ctx

	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)

	lenderAddr := sdk.AccAddress([]byte("addr________________"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)

	// mint and send coins
	suite.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	suite.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr, initCoins))

	// artificially set uToken exchange rate to 2.0
	err := app.LeverageKeeper.SetExchangeRate(ctx, umeeapp.BondDenom, sdk.MustNewDecFromStr("2.0"))
	suite.Require().NoError(err)

	// lend asset
	err = app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	suite.Require().NoError(err)

	// verify the total supply of the minted uToken (500 instead of 1000 due to exchange rate)
	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)
	supply := app.LeverageKeeper.TotalUTokenSupply(ctx, uTokenDenom)
	expected := sdk.NewInt64Coin(uTokenDenom, 500000000) // 500 u/umee
	suite.Require().Equal(expected, supply)

	// withdraw the total amount of assets lent
	uToken := expected
	err = app.LeverageKeeper.WithdrawAsset(ctx, lenderAddr, uToken)
	suite.Require().NoError(err)

	// verify total supply of the uTokens
	supply = app.LeverageKeeper.TotalUTokenSupply(ctx, uTokenDenom)
	suite.Require().Equal(int64(0), supply.Amount.Int64())

	// verify the lender's balances
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(initTokens, tokenBalance.Amount)

	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, uTokenDenom)
	suite.Require().Equal(int64(0), uTokenBalance.Amount.Int64())
}

func (suite *IntegrationTestSuite) TestSetReserves() {
	app, ctx := suite.app, suite.ctx

	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)

	// get initial reserves
	amount := app.LeverageKeeper.GetReserveAmount(ctx, umeeapp.BondDenom)
	suite.Require().Equal(amount, sdk.ZeroInt())

	// artifically reserve 200 umee
	err := app.LeverageKeeper.IncreaseReserves(ctx, sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(200000000)))
	suite.Require().NoError(err)

	// get new reserves
	amount = app.LeverageKeeper.GetReserveAmount(ctx, umeeapp.BondDenom)
	suite.Require().Equal(amount, sdk.NewInt(200000000))
}

func (suite *IntegrationTestSuite) TestSetExchangeRate() {
	app, ctx := suite.app, suite.ctx

	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)

	// get initial exchange rate
	rate, err := app.LeverageKeeper.GetExchangeRate(ctx, umeeapp.BondDenom)
	suite.Require().NoError(err)
	suite.Require().Equal(rate, sdk.OneDec())

	// artifically set exchange rate to 3.0
	err = app.LeverageKeeper.SetExchangeRate(ctx, umeeapp.BondDenom, sdk.MustNewDecFromStr("3.0"))
	suite.Require().NoError(err)

	// get new exchange rate
	rate, err = app.LeverageKeeper.GetExchangeRate(ctx, umeeapp.BondDenom)
	suite.Require().NoError(err)
	suite.Require().Equal(rate, sdk.OneDec().MulInt64(3))
}

func (suite *IntegrationTestSuite) TestGetToken() {
	app, ctx := suite.app, suite.ctx

	uabc := types.Token{
		BaseDenom:           "uabc",
		ReserveFactor:       sdk.MustNewDecFromStr("0.1"),
		CollateralWeight:    sdk.MustNewDecFromStr("0.2"),
		BaseBorrowRate:      sdk.MustNewDecFromStr("0.3"),
		KinkBorrowRate:      sdk.MustNewDecFromStr("0.4"),
		MaxBorrowRate:       sdk.MustNewDecFromStr("0.5"),
		KinkUtilizationRate: sdk.MustNewDecFromStr("0.6"),
	}
	app.LeverageKeeper.SetRegisteredToken(ctx, uabc)
	app.LeverageKeeper.SetTokenDenom(ctx, uabc.BaseDenom)

	reserveFactor, err := app.LeverageKeeper.GetReserveFactor(ctx, "uabc")
	suite.Require().NoError(err)
	suite.Require().Equal(reserveFactor, sdk.MustNewDecFromStr("0.1"))

	collateralWeight, err := app.LeverageKeeper.GetCollateralWeight(ctx, "uabc")
	suite.Require().NoError(err)
	suite.Require().Equal(collateralWeight, sdk.MustNewDecFromStr("0.2"))

	baseBorrowRate, err := app.LeverageKeeper.GetInterestBase(ctx, "uabc")
	suite.Require().NoError(err)
	suite.Require().Equal(baseBorrowRate, sdk.MustNewDecFromStr("0.3"))

	kinkBorrowRate, err := app.LeverageKeeper.GetInterestAtKink(ctx, "uabc")
	suite.Require().NoError(err)
	suite.Require().Equal(kinkBorrowRate, sdk.MustNewDecFromStr("0.4"))

	maxBorrowRate, err := app.LeverageKeeper.GetInterestMax(ctx, "uabc")
	suite.Require().NoError(err)
	suite.Require().Equal(maxBorrowRate, sdk.MustNewDecFromStr("0.5"))

	kinkUtilizationRate, err := app.LeverageKeeper.GetInterestKinkUtilization(ctx, "uabc")
	suite.Require().NoError(err)
	suite.Require().Equal(kinkUtilizationRate, sdk.MustNewDecFromStr("0.6"))
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

func (suite *IntegrationTestSuite) TestBorrowAsset_Reserved() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// artifically reserve 200 umee
	err := app.LeverageKeeper.IncreaseReserves(ctx, sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(200000000)))
	suite.Require().NoError(err)

	// lender tries to borrow 1000 umee, insufficient balance because 200 of the module's 1000 umee are reserved
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000))
	suite.Require().Error(err)

	// lender borrows 800 umee
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 800000000))
	suite.Require().NoError(err)
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

func (suite *IntegrationTestSuite) TestDeriveExchangeRate() {
	_, addr := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// The init scenario is being used so module balance starts at 1000 umee
	// and the uToken supply starts at 1000 due to lender account

	// artificially increase total borrows (by affecting a single address)
	err := app.LeverageKeeper.SetBorrow(ctx, addr, umeeapp.BondDenom, sdk.NewInt(2000000000)) // 2000 umee
	suite.Require().NoError(err)

	// artifically set reserves
	err = app.LeverageKeeper.IncreaseReserves(ctx, sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(300000000))) // 300 umee
	suite.Require().NoError(err)

	// expected token:uToken exchange rate
	//    = (total borrows + module balance - reserves) / utoken supply
	//    = 2000 + 1000 - 300 / 1000
	//    = 2.7

	// update exchange rates
	err = app.LeverageKeeper.UpdateExchangeRates(ctx)
	suite.Require().NoError(err)

	// get derived exchange rate
	rate, err := app.LeverageKeeper.GetExchangeRate(ctx, umeeapp.BondDenom)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.MustNewDecFromStr("2.7"), rate)
}

func (suite *IntegrationTestSuite) TestAccrueZeroInterest() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 50 umee
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 50000000))
	suite.Require().NoError(err)

	// verify lender's loan amount (50 umee)
	loanBalance := app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 50000000))

	// Because no time has passed since genesis (due to test setup) this will not increase borrowed amount
	err = app.LeverageKeeper.AccrueAllInterest(ctx)
	suite.Require().NoError(err)

	// verify lender's loan amount (50 umee)
	loanBalance = app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 50000000))
}

func (suite *IntegrationTestSuite) TestBorrowUtilizationNoReserves() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// Init scenario is being used because the module account (lending pool) already has 1000 umee

	// 0% utilization (0/1000)
	util, err := app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.ZeroInt())
	suite.Require().NoError(err)
	suite.Require().Equal(util, sdk.ZeroDec())

	// lender borrows 50 umee, reducing module account to 950 umee
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 50000000))
	suite.Require().NoError(err)

	// 5% utilization (50/1000)
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.NewInt(50000000))
	suite.Require().NoError(err)
	suite.Require().Equal(util, sdk.MustNewDecFromStr("0.05"))

	// lender borrows 950 umee, reducing module account to 0 umee
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 950000000))
	suite.Require().NoError(err)

	// 100% utilization (1000/1000)
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.NewInt(1000000000))
	suite.Require().NoError(err)
	suite.Require().Equal(util, sdk.OneDec())
}

func (suite *IntegrationTestSuite) TestBorrowUtilizationWithReserves() {
	lenderAddr, _ := suite.initBorrowScenario()
	app, ctx := suite.app, suite.ctx

	// Init scenario is being used because the module account (lending pool) already has 1000 umee

	// Math note:
	//   Token Utilization = Total Borrows / (Module Account Balance + Total Borrows - Reserve Requirement).
	//   GetBorrowUtilization takes total borrows as input, and automatically retrieves module balance and reserves.

	// Artificially set reserves to 300, leaving 700 lending pool available
	err := app.LeverageKeeper.IncreaseReserves(ctx, sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(300000000)))
	suite.Require().NoError(err)

	// Reserves = 300, module balance = 1000, total borrows = 0.
	// 0% utilization (0/700)
	util, err := app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.ZeroInt())
	suite.Require().NoError(err)
	suite.Require().Equal(util, sdk.ZeroDec())

	// lender borrows 70 umee
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 70000000))
	suite.Require().NoError(err)

	// Reserves = 300, module balance = 930, total borrows = 70.
	// 10% utilization (70/700)
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.NewInt(70000000))
	suite.Require().NoError(err)
	suite.Require().Equal(util, sdk.MustNewDecFromStr("0.10"))

	// lender borrows 630 umee
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 630000000))
	suite.Require().NoError(err)

	// Reserves = 300, module balance = 300, total borrows = 700.
	// 100% utilization (700/700)
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.NewInt(700000000))
	suite.Require().NoError(err)
	suite.Require().Equal(util, sdk.OneDec())

	// Artificially reserve additional 1k umee, to force edge cases and impossible scenarios below.
	err = app.LeverageKeeper.IncreaseReserves(ctx, sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(1000000000)))
	suite.Require().NoError(err)

	// Reserves = 1300, module balance = 300, total borrows = 2000.
	// Edge (but not impossible) case interpreted as 100% utilization (instead of >100% from equation).
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.NewInt(2000000000))
	suite.Require().NoError(err)
	suite.Require().Equal(util, sdk.OneDec())

	// Reserves = 1300, module balance = 300, total borrows = 0.
	// Denominator of utilization equation would be negative.
	// Impossible case interpreted as 100% utilization (instead of negative utilization from equation).
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.ZeroInt())
	suite.Require().NoError(err)
	suite.Require().Equal(util, sdk.OneDec())

	// Reserves = 1300, module balance = 300, total borrows = 1000.
	// Denominator of utilization equation would be zero.
	// Impossible case interpreted as 100% utilization (instead of divide by zero panic).
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.NewInt(1000000000))
	suite.Require().NoError(err)
	suite.Require().Equal(util, sdk.OneDec())
}

func (suite *IntegrationTestSuite) TestDynamicInterest() {
	app, ctx := suite.app, suite.ctx

	uabc := types.Token{
		BaseDenom:           "uabc",
		ReserveFactor:       sdk.MustNewDecFromStr("0"),
		CollateralWeight:    sdk.MustNewDecFromStr("0"),
		BaseBorrowRate:      sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:      sdk.MustNewDecFromStr("0.22"),
		MaxBorrowRate:       sdk.MustNewDecFromStr("1.52"),
		KinkUtilizationRate: sdk.MustNewDecFromStr("0.8"),
	}
	app.LeverageKeeper.SetRegisteredToken(ctx, uabc)
	app.LeverageKeeper.SetTokenDenom(ctx, "uabc")

	// Base interest rate (0% utilization)
	rate, err := app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.ZeroDec())
	suite.Require().NoError(err)
	suite.Require().Equal(rate, sdk.MustNewDecFromStr("0.02"))

	// Between base interest and kink (20% utilization)
	rate, err = app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.MustNewDecFromStr("0.20"))
	suite.Require().NoError(err)
	suite.Require().Equal(rate, sdk.MustNewDecFromStr("0.07"))

	// Kink interest rate (80% utilization)
	rate, err = app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.MustNewDecFromStr("0.80"))
	suite.Require().NoError(err)
	suite.Require().Equal(rate, sdk.MustNewDecFromStr("0.22"))

	// Between kink interest and max (90% utilization)
	rate, err = app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.MustNewDecFromStr("0.90"))
	suite.Require().NoError(err)
	suite.Require().Equal(rate, sdk.MustNewDecFromStr("0.87"))

	// Max interest rate (100% utilization)
	rate, err = app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.OneDec())
	suite.Require().NoError(err)
	suite.Require().Equal(rate, sdk.MustNewDecFromStr("1.52"))

	// Invalid utilization inputs
	_, err = app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.MustNewDecFromStr("-0.10"))
	suite.Require().Error(err)
	_, err = app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.MustNewDecFromStr("1.50"))
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestDynamicInterest_InvalidAsset() {
	app, ctx := suite.app, suite.ctx

	_, err := app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.MustNewDecFromStr("0.33"))
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

func (suite *IntegrationTestSuite) TestInterpolate() {
	app, ctx := suite.app, suite.ctx

	// Define two points (x1,y1) and (x2,y2)
	x1 := sdk.MustNewDecFromStr("3.0")
	x2 := sdk.MustNewDecFromStr("6.0")
	y1 := sdk.MustNewDecFromStr("11.1")
	y2 := sdk.MustNewDecFromStr("17.4")

	// Sloped line, endpoint checks
	x := app.LeverageKeeper.Interpolate(ctx, x1, x1, y1, x2, y2)
	suite.Require().Equal(x, y1)
	x = app.LeverageKeeper.Interpolate(ctx, x2, x1, y1, x2, y2)
	suite.Require().Equal(x, y2)

	// Sloped line, point on segment
	x = app.LeverageKeeper.Interpolate(ctx, sdk.MustNewDecFromStr("4.0"), x1, y1, x2, y2)
	suite.Require().Equal(x, sdk.MustNewDecFromStr("13.2"))

	// Sloped line, point outside of segment
	x = app.LeverageKeeper.Interpolate(ctx, sdk.MustNewDecFromStr("2.0"), x1, y1, x2, y2)
	suite.Require().Equal(x, sdk.MustNewDecFromStr("9.0"))

	// Vertical line: always return y1
	x = app.LeverageKeeper.Interpolate(ctx, sdk.ZeroDec(), x1, y1, x1, y2)
	suite.Require().Equal(x, y1)
	x = app.LeverageKeeper.Interpolate(ctx, x1, x1, y1, x1, y2)
	suite.Require().Equal(x, y1)

	// Undefined line (x1=x2, y1=y2): always return y1
	x = app.LeverageKeeper.Interpolate(ctx, sdk.ZeroDec(), x1, y1, x1, y1)
	suite.Require().Equal(x, y1)
	x = app.LeverageKeeper.Interpolate(ctx, x1, x1, y1, x1, y1)
	suite.Require().Equal(x, y1)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
