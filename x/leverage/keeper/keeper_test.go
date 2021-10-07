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

// Helper function: Initialize the common starting scenario from which borrow and repay tests stem:
func (suite *IntegrationTestSuite) initBorrowScenario() (lender, borrower, bum sdk.AccAddress) {
	app, ctx := suite.app, suite.ctx
	// register uumee and u/uumee as an accepted asset+utoken pair
	app.LeverageKeeper.SetTokenDenom(ctx, umeeapp.BondDenom)
	// create an account and address which will represent a lender
	lenderAddr := sdk.AccAddress([]byte("addr______________00"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)
	// create an account and address which will represent a borrower
	borrowerrAddr := sdk.AccAddress([]byte("addr______________01"))
	borrowerAcc := app.AccountKeeper.NewAccountWithAddress(ctx, borrowerrAddr)
	app.AccountKeeper.SetAccount(ctx, borrowerAcc)
	// create an account and address which will represent a user with no assets
	bumAddr := sdk.AccAddress([]byte("addr______________02"))
	bumAcc := app.AccountKeeper.NewAccountWithAddress(ctx, bumAddr)
	app.AccountKeeper.SetAccount(ctx, bumAcc)
	// mint and send 10k umee to lender
	suite.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	suite.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr, initCoins))
	// lender lends 1k umee and receives 1k u/umee
	err := app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	suite.Require().NoError(err)
	// mint and send 10k umee to borrower
	suite.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	suite.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, borrowerrAddr, initCoins))
	// ensure suite's context remembers the events above
	suite.ctx = ctx
	// return the three account addresses
	return lenderAddr, borrowerrAddr, bumAddr
	// The starting scenario is thus:
	// - umee and u/umee are accepted assets
	// - a "lender" user has 9k umee and 1k u/umee
	// - the leverage module has 1k umee due to lender's activity
	// - a "borrower" user has 10k umee
	// - a "bum" user has an address but no assets
}

func (suite *IntegrationTestSuite) TestBorrowAsset_Invalid() {
	lenderAddr, _, _ := suite.initBorrowScenario() // create initial conditions
	app, ctx := suite.app, suite.ctx               // get ctx after init

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender attempts to borrow 200 u/umee, fails because uTokens cannot be borrowed
	err := app.LeverageKeeper.BorrowAsset(ctx,
		lenderAddr,
		sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 200000000), // 200 u/umee
	)
	suite.Require().Error(err)

	// lender attempts to borrow 200 abcd, fails because "abcd" is not a valid denom
	err = app.LeverageKeeper.BorrowAsset(ctx,
		lenderAddr,
		sdk.NewInt64Coin("abcd", 200000000), // 200 abcd
	)
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestBorrowAsset_Valid() {
	lenderAddr, _, _ := suite.initBorrowScenario() // create initial conditions
	app, ctx := suite.app, suite.ctx               // get ctx after init

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 200 umee
	err := app.LeverageKeeper.BorrowAsset(ctx,
		lenderAddr,
		sdk.NewInt64Coin(umeeapp.BondDenom, 200000000), // 200 umee
	)
	suite.Require().NoError(err)

	// verify lender's new loan amount in the correct denom
	bal := app.LeverageKeeper.GetLoan(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(bal, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000)) // 200 umee

	// verify lender's total loan amount (sdk.Coins) is also correct
	loans, err := app.LeverageKeeper.GetAllLoans(ctx, lenderAddr)
	suite.Require().Equal(loans, sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))) // 200 umee

	// verify lender's new umee balance (10 - 1k from initial + 200 from loan = 9200)
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	suite.Require().Equal(initTokens.Sub(sdk.NewInt(800000000)), tokenBalance.Amount)

	// verify lender's uToken balance remains at 1k from initial conditions
	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	suite.Require().Equal(int64(1000000000), uTokenBalance.Amount.Int64())
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
