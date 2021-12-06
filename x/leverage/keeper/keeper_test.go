package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeapp "github.com/umee-network/umee/app"
	"github.com/umee-network/umee/x/leverage/keeper"
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

	ctx         sdk.Context
	app         *umeeapp.UmeeApp
	queryClient types.QueryClient
}

func (s *IntegrationTestSuite) SetupTest() {
	app := umeeapp.Setup(s.T(), false, 1)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})

	uumee := types.Token{
		BaseDenom:            umeeapp.BondDenom,
		ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:     sdk.MustNewDecFromStr("0.1"),
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.2"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
	}

	// At the moment, SetRegisteredToken must be followed separately by SetTokenDenom
	// to complete token registration. Therefore, this line does not break the InvalidAsset tests
	// which require 'uumee' to be unregistered.
	app.LeverageKeeper.SetRegisteredToken(ctx, uumee)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(app.LeverageKeeper))

	s.app = app
	s.ctx = ctx
	s.queryClient = types.NewQueryClient(queryHelper)
}

func (s *IntegrationTestSuite) TestLendAsset_InvalidAsset() {
	app, ctx := s.app, s.ctx

	lenderAddr := sdk.AccAddress([]byte("addr________________"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)

	// mint and send coins
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr, initCoins))

	// lending should fail as we have not set what tokens can be lent
	err := app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestLendAsset_Valid() {
	app, ctx := s.app, s.ctx

	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)

	lenderAddr := sdk.AccAddress([]byte("addr________________"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)

	// mint and send coins
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr, initCoins))

	// lend asset
	err := app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	s.Require().NoError(err)

	// verify the total supply of the minted uToken
	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)
	supply := app.LeverageKeeper.TotalUTokenSupply(ctx, uTokenDenom)
	expected := sdk.NewInt64Coin(uTokenDenom, 1000000000) // 1k u/umee
	s.Require().Equal(expected, supply)

	// verify the lender's balances
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(initTokens.Sub(sdk.NewInt(1000000000)), tokenBalance.Amount)

	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, uTokenDenom)
	s.Require().Equal(int64(1000000000), uTokenBalance.Amount.Int64())
}

func (s *IntegrationTestSuite) TestWithdrawAsset_Valid() {
	app, ctx := s.app, s.ctx

	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)

	lenderAddr := sdk.AccAddress([]byte("addr________________"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)

	// mint and send coins
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr, initCoins))

	// lend asset
	err := app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	s.Require().NoError(err)

	// verify the total supply of the minted uToken
	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)
	supply := app.LeverageKeeper.TotalUTokenSupply(ctx, uTokenDenom)
	expected := sdk.NewInt64Coin(uTokenDenom, 1000000000) // 1k u/umee
	s.Require().Equal(expected, supply)

	// withdraw the total amount of assets lent
	uToken := expected
	err = app.LeverageKeeper.WithdrawAsset(ctx, lenderAddr, uToken)
	s.Require().NoError(err)

	// verify total supply of the uTokens
	supply = app.LeverageKeeper.TotalUTokenSupply(ctx, uTokenDenom)
	s.Require().Equal(int64(0), supply.Amount.Int64())

	// verify the lender's balances
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(initTokens, tokenBalance.Amount)

	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, uTokenDenom)
	s.Require().Equal(int64(0), uTokenBalance.Amount.Int64())
}

func (s *IntegrationTestSuite) TestWithdrawAsset_WithExchangeRate() {
	app, ctx := s.app, s.ctx

	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)

	lenderAddr := sdk.AccAddress([]byte("addr________________"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)

	// mint and send coins
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr, initCoins))

	// artificially set uToken exchange rate to 2.0
	err := app.LeverageKeeper.SetExchangeRate(ctx, umeeapp.BondDenom, sdk.MustNewDecFromStr("2.0"))
	s.Require().NoError(err)

	// lend asset
	err = app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	s.Require().NoError(err)

	// verify the total supply of the minted uToken (500 instead of 1000 due to exchange rate)
	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)
	supply := app.LeverageKeeper.TotalUTokenSupply(ctx, uTokenDenom)
	expected := sdk.NewInt64Coin(uTokenDenom, 500000000) // 500 u/umee
	s.Require().Equal(expected, supply)

	// withdraw the total amount of assets lent
	uToken := expected
	err = app.LeverageKeeper.WithdrawAsset(ctx, lenderAddr, uToken)
	s.Require().NoError(err)

	// verify total supply of the uTokens
	supply = app.LeverageKeeper.TotalUTokenSupply(ctx, uTokenDenom)
	s.Require().Equal(int64(0), supply.Amount.Int64())

	// verify the lender's balances
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(initTokens, tokenBalance.Amount)

	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, uTokenDenom)
	s.Require().Equal(int64(0), uTokenBalance.Amount.Int64())
}

func (s *IntegrationTestSuite) TestSetReserves() {
	app, ctx := s.app, s.ctx

	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)

	// get initial reserves
	amount := app.LeverageKeeper.GetReserveAmount(ctx, umeeapp.BondDenom)
	s.Require().Equal(amount, sdk.ZeroInt())

	// artifically reserve 200 umee
	err := app.LeverageKeeper.IncreaseReserves(ctx, sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(200000000)))
	s.Require().NoError(err)

	// get new reserves
	amount = app.LeverageKeeper.GetReserveAmount(ctx, umeeapp.BondDenom)
	s.Require().Equal(amount, sdk.NewInt(200000000))
}

func (s *IntegrationTestSuite) TestSetExchangeRate() {
	app, ctx := s.app, s.ctx

	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)

	// get initial exchange rate
	rate, err := app.LeverageKeeper.GetExchangeRate(ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.OneDec())

	// artifically set exchange rate to 3.0
	err = app.LeverageKeeper.SetExchangeRate(ctx, umeeapp.BondDenom, sdk.MustNewDecFromStr("3.0"))
	s.Require().NoError(err)

	// get new exchange rate
	rate, err = app.LeverageKeeper.GetExchangeRate(ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.OneDec().MulInt64(3))
}

func (s *IntegrationTestSuite) TestGetToken() {
	app, ctx := s.app, s.ctx

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
	s.Require().NoError(err)
	s.Require().Equal(reserveFactor, sdk.MustNewDecFromStr("0.1"))

	collateralWeight, err := app.LeverageKeeper.GetCollateralWeight(ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(collateralWeight, sdk.MustNewDecFromStr("0.2"))

	baseBorrowRate, err := app.LeverageKeeper.GetInterestBase(ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(baseBorrowRate, sdk.MustNewDecFromStr("0.3"))

	kinkBorrowRate, err := app.LeverageKeeper.GetInterestAtKink(ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(kinkBorrowRate, sdk.MustNewDecFromStr("0.4"))

	maxBorrowRate, err := app.LeverageKeeper.GetInterestMax(ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(maxBorrowRate, sdk.MustNewDecFromStr("0.5"))

	kinkUtilizationRate, err := app.LeverageKeeper.GetInterestKinkUtilization(ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(kinkUtilizationRate, sdk.MustNewDecFromStr("0.6"))
}

// initialize the common starting scenario from which borrow and repay tests stem:
// Umee and u/umee are registered assets; a "lender" account has 9k umee and 1k u/umee;
// the leverage module has 1k umee in its lending pool (module account); and a "bum"
// account has been created with no assets.
func (s *IntegrationTestSuite) initBorrowScenario() (lender, bum sdk.AccAddress) {
	app, ctx := s.app, s.ctx

	// register uumee and u/uumee as an accepted asset+utoken pair
	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)

	// set default params
	params := types.DefaultParams()
	app.LeverageKeeper.SetParams(ctx, params)

	// create an account and address which will represent a lender
	lenderAddr := sdk.AccAddress([]byte("addr______________01"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)

	// create an account and address which will represent a user with no assets
	bumAddr := sdk.AccAddress([]byte("addr______________02"))
	bumAcc := app.AccountKeeper.NewAccountWithAddress(ctx, bumAddr)
	app.AccountKeeper.SetAccount(ctx, bumAcc)

	// mint and send 10k umee to lender
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName,
		sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 10000000000)), // 10k umee
	))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr,
		sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 10000000000)), // 10k umee,
	))

	// lender lends 1000 umee and receives 1k u/umee
	err := app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000))
	s.Require().NoError(err)

	// lender enables u/umee as collateral
	collatDenom := app.LeverageKeeper.FromTokenToUTokenDenom(ctx, umeeapp.BondDenom)
	err = app.LeverageKeeper.SetCollateralSetting(ctx, lenderAddr, collatDenom, true)
	s.Require().NoError(err)

	// return the account addresses
	return lenderAddr, bumAddr
}

func (s *IntegrationTestSuite) TestBorrowAsset_Invalid() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender attempts to borrow 200 u/umee, fails because uTokens cannot be borrowed
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 200000000))
	s.Require().Error(err)

	// lender attempts to borrow 200 abcd, fails because "abcd" is not a valid denom
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin("uabcd", 200000000))
	s.Require().Error(err)
}

/*
// TODO: Ready to be commented in once borrowing limits exist.
func (s *IntegrationTestSuite) TestBorrowAsset_InsufficientCollateral() {
	_, bumAddr := s.initBorrowScenario() // create initial conditions
	app, ctx := s.app, s.ctx            // get ctx after init

	// The "bum" user from the init scenario is being used because it
	// possesses no assets or collateral.

	// bum attempts to borrow 200 umee, fails because of insufficient collateral
	err := app.LeverageKeeper.BorrowAsset(ctx,bumAddr,tCoin("umee",200))
	s.Require().Error(err)
}
*/

func (s *IntegrationTestSuite) TestBorrowAsset_InsufficientLendingPool() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// Any user from the init scenario can perform this test, because it errors on module balance

	// lender attempts to borrow 20000 umee, fails because of insufficient module account balance
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestRepayAsset_Invalid() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// Any user from the init scenario can be used for this test.

	// lender attempts to repay 200 abcd, fails because "abcd" is not an accepted asset type
	_, err := app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin("uabcd", 200000000))
	s.Require().Error(err)

	// lender attempts to repay 200 u/umee, fails because utokens are not loanable assets
	_, err = app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 200000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestBorrowAsset_Valid() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 200 umee
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	s.Require().NoError(err)

	// verify lender's new loan amount in the correct denom (200 umee)
	loanBalance := app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))

	// verify lender's total loan balance (sdk.Coins) is also 200 umee (no other coins present)
	totalLoanBalance, err := app.LeverageKeeper.GetBorrowerBorrows(ctx, lenderAddr)
	s.Require().Equal(totalLoanBalance, sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 200000000)))

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan = 9200 umee)
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9200000000))

	// verify lender's uToken balance remains at 1000 u/umee from initial conditions
	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))
}

func (s *IntegrationTestSuite) TestBorrowAsset_Reserved() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// artifically reserve 200 umee
	err := app.LeverageKeeper.IncreaseReserves(ctx, sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(200000000)))
	s.Require().NoError(err)

	// lender tries to borrow 1000 umee, insufficient balance because 200 of the module's 1000 umee are reserved
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000))
	s.Require().Error(err)

	// lender borrows 800 umee
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 800000000))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestRepayAsset_Valid() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 200 umee
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	s.Require().NoError(err)

	// lender repays 80 umee
	repaid, err := app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 80000000))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(80000000), repaid)

	// verify lender's new loan amount (120 umee)
	loanBalance := app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 120000000))

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan - 80 repaid = 9120 umee)
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9120000000))

	// verify lender's uToken balance remains at 1000 u/umee from initial conditions
	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))

	// lender repays 120 umee (loan repaid in full)
	repaid, err = app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 120000000))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(120000000), repaid)

	// verify lender's new loan amount in the correct denom (zero)
	loanBalance = app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 0))

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan - 200 repaid = 9000 umee)
	tokenBalance = app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9000000000))

	// verify lender's uToken balance remains at 1000 u/umee from initial conditions
	uTokenBalance = app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))
}

func (s *IntegrationTestSuite) TestRepayAsset_Overpay() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 200 umee
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	s.Require().NoError(err)

	// lender repays 300 umee - should automatically reduce to 200 (the loan amount) and succeed
	repaid, err := app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 300000000))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(200000000), repaid)

	// verify lender's new loan amount is 0 umee
	loanBalance := app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 0))

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan - 200 repaid = 9000 umee)
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9000000000))

	// verify lender's uToken balance remains at 1k u/umee from initial conditions
	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))

	// lender repays 50 umee - this time it fails because the loan no longer exists
	_, err = app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 50000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestGetCollateral() {
	lenderAddr, bumAddr := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral. The "bum" user is used because
	// it has none.

	// Verify lender collateral is 1k u/umee
	collateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, lenderAddr)
	collatDenom := app.LeverageKeeper.FromTokenToUTokenDenom(ctx, umeeapp.BondDenom)
	s.Require().Equal(collateral, sdk.NewCoins(sdk.NewInt64Coin(collatDenom, 1000000000)))

	// Verify bum collateral is empty
	collateral = app.LeverageKeeper.GetBorrowerCollateral(ctx, bumAddr)
	s.Require().Equal(collateral, sdk.NewCoins())
}

func (s *IntegrationTestSuite) TestBorrowLimit() {
	app, ctx := s.app, s.ctx

	// register uumee and u/uumee as an accepted asset+utoken pair
	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)

	// Create collateral utokens (1k u/umee)
	collatDenom := app.LeverageKeeper.FromTokenToUTokenDenom(ctx, umeeapp.BondDenom)
	collateral := sdk.NewCoins(sdk.NewInt64Coin(collatDenom, 1000000000))

	// Manually compute borrow limit using collateral weight of 0.1
	// and placeholder of 1 uumee = 1 USD value
	expected := collateral[0].Amount.ToDec().Mul(sdk.MustNewDecFromStr("0.1"))

	// Check borrow limit vs. manually computed value
	borrowLimit, err := app.LeverageKeeper.CalculateBorrowLimit(ctx, collateral)
	s.Require().NoError(err)
	s.Require().Equal(expected, borrowLimit)
}

func (s *IntegrationTestSuite) TestLiqudateBorrow_Valid() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral.

	// lender borrows 90 umee
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 90000000))
	s.Require().NoError(err)

	// create an account and address which will represent a liquidator
	liquidatorAddr := sdk.AccAddress([]byte("addr______________03"))
	liquidatorAcc := app.AccountKeeper.NewAccountWithAddress(ctx, liquidatorAddr)
	app.AccountKeeper.SetAccount(ctx, liquidatorAcc)

	// mint and send 10k umee to liquiator
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName,
		sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 10000000000)), // 10k umee
	))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, liquidatorAddr,
		sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 10000000000)), // 10k umee,
	))

	// liquidator attempts to liquidate lender, but lender is ineligible (not over borrow limit)
	repayment := sdk.NewInt64Coin(umeeapp.BondDenom, 30000000) // 30 umee
	rewardDenom := app.LeverageKeeper.FromTokenToUTokenDenom(ctx, umeeapp.BondDenom)
	_, _, err = app.LeverageKeeper.LiquidateBorrow(ctx, liquidatorAddr, lenderAddr, repayment, rewardDenom)
	s.Require().Error(err)

	// amount owed is forcefully increased to 200 umee (over borrow limit)
	err = app.LeverageKeeper.SetBorrow(ctx, lenderAddr, umeeapp.BondDenom, sdk.NewInt(200000000))
	s.Require().NoError(err)

	// liquidator partially liquidates lender, receiving some collateral
	repayment = sdk.NewInt64Coin(umeeapp.BondDenom, 10000000) // 10 umee
	repaid, reward, err := app.LeverageKeeper.LiquidateBorrow(ctx, liquidatorAddr, lenderAddr, repayment, rewardDenom)
	s.Require().NoError(err)
	s.Require().Equal(repayment.Amount, repaid)
	s.Require().Equal(sdk.NewInt(11000000), reward)

	// verify lender's new loan amount is 190 umee (still over borrow limit)
	loanBalance := app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance.String(), sdk.NewInt64Coin(umeeapp.BondDenom, 190000000).String())

	// verify liquidator's new u/umee balance = 11 = (10 + liquidation incentive)
	uTokenBalance := app.BankKeeper.GetBalance(ctx, liquidatorAddr, rewardDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin(rewardDenom, 11000000))

	// verify liquidator's new umee balance (10k - 10) = 9990 umee
	tokenBalance := app.BankKeeper.GetBalance(ctx, liquidatorAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9990000000))

	// liquidator fully liquidates lender, receiving more collateral and reducing borrowed amount to zero
	repayment = sdk.NewInt64Coin(umeeapp.BondDenom, 300000000) // 300 umee
	repaid, reward, err = app.LeverageKeeper.LiquidateBorrow(ctx, liquidatorAddr, lenderAddr, repayment, rewardDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(190000000), repaid)
	s.Require().Equal(sdk.NewInt(209000000), reward)

	// verify liquidator's new u/umee balance = 220 = (200 + liquidation incentive)
	uTokenBalance = app.BankKeeper.GetBalance(ctx, liquidatorAddr, rewardDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin(rewardDenom, 220000000))

	// verify lender's new loan amount is zero
	loanBalance = app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 0))

	// verify liquidator's new umee balance (10k - 200) = 9800 umee
	tokenBalance = app.BankKeeper.GetBalance(ctx, liquidatorAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9800000000))
}

func (s *IntegrationTestSuite) TestDeriveExchangeRate() {
	_, addr := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// The init scenario is being used so module balance starts at 1000 umee
	// and the uToken supply starts at 1000 due to lender account

	// artificially increase total borrows (by affecting a single address)
	err := app.LeverageKeeper.SetBorrow(ctx, addr, umeeapp.BondDenom, sdk.NewInt(2000000000)) // 2000 umee
	s.Require().NoError(err)

	// artifically set reserves
	err = app.LeverageKeeper.IncreaseReserves(ctx, sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(300000000))) // 300 umee
	s.Require().NoError(err)

	// expected token:uToken exchange rate
	//    = (total borrows + module balance - reserves) / utoken supply
	//    = 2000 + 1000 - 300 / 1000
	//    = 2.7

	// update exchange rates
	err = app.LeverageKeeper.UpdateExchangeRates(ctx)
	s.Require().NoError(err)

	// get derived exchange rate
	rate, err := app.LeverageKeeper.GetExchangeRate(ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("2.7"), rate)
}

func (s *IntegrationTestSuite) TestAccrueZeroInterest() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 50 umee
	err := app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 50000000))
	s.Require().NoError(err)

	// verify lender's loan amount (50 umee)
	loanBalance := app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 50000000))

	// Because no time has passed since genesis (due to test setup) this will not increase borrowed amount
	err = app.LeverageKeeper.AccrueAllInterest(ctx)
	s.Require().NoError(err)

	// verify lender's loan amount (50 umee)
	loanBalance = app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 50000000))
}

func (s *IntegrationTestSuite) TestBorrowUtilizationNoReserves() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// Init scenario is being used because the module account (lending pool) already has 1000 umee

	// 0% utilization (0/1000)
	util, err := app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.ZeroInt())
	s.Require().NoError(err)
	s.Require().Equal(util, sdk.ZeroDec())

	// lender borrows 50 umee, reducing module account to 950 umee
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 50000000))
	s.Require().NoError(err)

	// 5% utilization (50/1000)
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.NewInt(50000000))
	s.Require().NoError(err)
	s.Require().Equal(util, sdk.MustNewDecFromStr("0.05"))

	// lender borrows 950 umee, reducing module account to 0 umee
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 950000000))
	s.Require().NoError(err)

	// 100% utilization (1000/1000)
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.NewInt(1000000000))
	s.Require().NoError(err)
	s.Require().Equal(util, sdk.OneDec())
}

func (s *IntegrationTestSuite) TestBorrowUtilizationWithReserves() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// Init scenario is being used because the module account (lending pool) already has 1000 umee

	// Math note:
	//   Token Utilization = Total Borrows / (Module Account Balance + Total Borrows - Reserve Requirement).
	//   GetBorrowUtilization takes total borrows as input, and automatically retrieves module balance and reserves.

	// Artificially set reserves to 300, leaving 700 lending pool available
	err := app.LeverageKeeper.IncreaseReserves(ctx, sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(300000000)))
	s.Require().NoError(err)

	// Reserves = 300, module balance = 1000, total borrows = 0.
	// 0% utilization (0/700)
	util, err := app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.ZeroInt())
	s.Require().NoError(err)
	s.Require().Equal(util, sdk.ZeroDec())

	// lender borrows 70 umee
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 70000000))
	s.Require().NoError(err)

	// Reserves = 300, module balance = 930, total borrows = 70.
	// 10% utilization (70/700)
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.NewInt(70000000))
	s.Require().NoError(err)
	s.Require().Equal(util, sdk.MustNewDecFromStr("0.10"))

	// lender borrows 630 umee
	err = app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 630000000))
	s.Require().NoError(err)

	// Reserves = 300, module balance = 300, total borrows = 700.
	// 100% utilization (700/700)
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.NewInt(700000000))
	s.Require().NoError(err)
	s.Require().Equal(util, sdk.OneDec())

	// Artificially reserve additional 1k umee, to force edge cases and impossible scenarios below.
	err = app.LeverageKeeper.IncreaseReserves(ctx, sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(1000000000)))
	s.Require().NoError(err)

	// Reserves = 1300, module balance = 300, total borrows = 2000.
	// Edge (but not impossible) case interpreted as 100% utilization (instead of >100% from equation).
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.NewInt(2000000000))
	s.Require().NoError(err)
	s.Require().Equal(util, sdk.OneDec())

	// Reserves = 1300, module balance = 300, total borrows = 0.
	// Denominator of utilization equation would be negative.
	// Impossible case interpreted as 100% utilization (instead of negative utilization from equation).
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.ZeroInt())
	s.Require().NoError(err)
	s.Require().Equal(util, sdk.OneDec())

	// Reserves = 1300, module balance = 300, total borrows = 1000.
	// Denominator of utilization equation would be zero.
	// Impossible case interpreted as 100% utilization (instead of divide by zero panic).
	util, err = app.LeverageKeeper.GetBorrowUtilization(ctx, umeeapp.BondDenom, sdk.NewInt(1000000000))
	s.Require().NoError(err)
	s.Require().Equal(util, sdk.OneDec())
}

func (s *IntegrationTestSuite) TestDynamicInterest() {
	app, ctx := s.app, s.ctx

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
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.02"))

	// Between base interest and kink (20% utilization)
	rate, err = app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.MustNewDecFromStr("0.20"))
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.07"))

	// Kink interest rate (80% utilization)
	rate, err = app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.MustNewDecFromStr("0.80"))
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.22"))

	// Between kink interest and max (90% utilization)
	rate, err = app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.MustNewDecFromStr("0.90"))
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.87"))

	// Max interest rate (100% utilization)
	rate, err = app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.OneDec())
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("1.52"))

	// Invalid utilization inputs
	_, err = app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.MustNewDecFromStr("-0.10"))
	s.Require().Error(err)
	_, err = app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.MustNewDecFromStr("1.50"))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestDynamicInterest_InvalidAsset() {
	app, ctx := s.app, s.ctx

	_, err := app.LeverageKeeper.GetDynamicBorrowInterest(ctx, "uabc", sdk.MustNewDecFromStr("0.33"))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestSetCollateral_Valid() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// Any user from the starting scenario can be used, since they are only toggling
	// collateral settings.

	// lender disables u/umee as collateral
	err := app.LeverageKeeper.SetCollateralSetting(ctx,
		lenderAddr,
		"u/uumee",
		false,
	)
	s.Require().NoError(err)
	enabled := app.LeverageKeeper.GetCollateralSetting(ctx,
		lenderAddr,
		"u/uumee",
	)
	s.Require().Equal(enabled, false)

	// lender enables u/umee as collateral
	err = app.LeverageKeeper.SetCollateralSetting(ctx,
		lenderAddr,
		"u/uumee",
		true,
	)
	s.Require().NoError(err)
	enabled = app.LeverageKeeper.GetCollateralSetting(ctx,
		lenderAddr,
		"u/uumee",
	)
	s.Require().Equal(enabled, true)
}

func (s *IntegrationTestSuite) TestSetCollateral_Invalid() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// Any user from the starting scenario can be used, since they are only toggling
	// collateral settings.

	// lender disables u/abcd as collateral - fails because "u/abcd" is not a recognized uToken
	err := app.LeverageKeeper.SetCollateralSetting(ctx, lenderAddr, "u/abcd", false)
	s.Require().Error(err)

	// lender disables uumee as collateral - fails because "uumee" is an asset, not a uToken
	err = app.LeverageKeeper.SetCollateralSetting(ctx, lenderAddr, "uumee", false)
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestGetCollateral_Invalid() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// Any user from the starting scenario can be used, since we are only viewing
	// collateral settings.

	// Regular assets always return false, because only uTokens can be collateral
	enabled := app.LeverageKeeper.GetCollateralSetting(ctx,
		lenderAddr,
		"uumee",
	)
	s.Require().Equal(enabled, false)

	// Invalid or unrecognized assets always return false
	enabled = app.LeverageKeeper.GetCollateralSetting(ctx,
		lenderAddr,
		"abcd",
	)
	s.Require().Equal(enabled, false)
}

func (s *IntegrationTestSuite) TestParams() {
	app, ctx := s.app, s.ctx

	params := types.DefaultParams()
	app.LeverageKeeper.SetParams(ctx, params)

	got := app.LeverageKeeper.GetParams(ctx)
	s.Require().Equal(params, got)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
