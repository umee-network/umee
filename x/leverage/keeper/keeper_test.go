package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/x/leverage"
	"github.com/umee-network/umee/v2/x/leverage/keeper"
	"github.com/umee-network/umee/v2/x/leverage/types"
)

const (
	initialPower = int64(10000000000)
	atomIBCDenom = "ibc/CDC4587874B85BEA4FCEC3CEA5A1195139799A1FEE711A07D972537E18FDA39D"
	umeeDenom    = umeeapp.BondDenom
)

var (
	initTokens          = sdk.TokensFromConsensusPower(initialPower, sdk.DefaultPowerReduction)
	initCoins           = sdk.NewCoins(sdk.NewCoin(umeeapp.BondDenom, initTokens))
	setupAccountCounter = sdk.ZeroInt()
)

// creates a test token with reasonable initial parameters
func newToken(base, symbol string) types.Token {
	return types.Token{
		BaseDenom:              base,
		SymbolDenom:            symbol,
		Exponent:               6,
		ReserveFactor:          sdk.MustNewDecFromStr("0.20"),
		CollateralWeight:       sdk.MustNewDecFromStr("0.25"),
		LiquidationThreshold:   sdk.MustNewDecFromStr("0.25"),
		BaseBorrowRate:         sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:         sdk.MustNewDecFromStr("0.22"),
		MaxBorrowRate:          sdk.MustNewDecFromStr("1.52"),
		KinkUtilization:        sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive:   sdk.MustNewDecFromStr("0.1"),
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdk.MustNewDecFromStr("1.00000000000000000"),
		MaxSupplyUtilization:   sdk.MustNewDecFromStr("1.00000000000000000"),
		MinCollateralLiquidity: sdk.MustNewDecFromStr("0.00000000000000000"),
	}
}

type IntegrationTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *umeeapp.UmeeApp
	tk          keeper.TestKeeper
	queryClient types.QueryClient
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
	atomIBCToken := newToken(atomIBCDenom, "ATOM")

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
	s.queryClient = types.NewQueryClient(queryHelper)
}

// setupAccount executes some common boilerplate before a test, where a user account is given tokens of a given denom,
// may also supply them to receive uTokens, and may also enable those uTokens as collateral and borrow tokens in the same denom.
func (s *IntegrationTestSuite) setupAccount(denom string, mintAmount, supplyAmount, borrowAmount int64, collateral bool) sdk.AccAddress {
	// create a unique address
	setupAccountCounter = setupAccountCounter.Add(sdk.OneInt())
	addr := sdk.AccAddress([]byte("addr" + setupAccountCounter.String()))

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
		err := s.app.LeverageKeeper.Supply(s.ctx, addr, sdk.NewInt64Coin(denom, supplyAmount))
		s.Require().NoError(err)
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

func (s *IntegrationTestSuite) TestSupply_InvalidAsset() {
	addr := sdk.AccAddress([]byte("addr________________"))
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)

	// create coins of an unregistered base asset type "uabcd"
	invalidCoin := sdk.NewInt64Coin("uabcd", 1000000000) // 1k abcd
	invalidCoins := sdk.NewCoins(invalidCoin)

	// mint and send coins
	s.Require().NoError(s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, invalidCoins))
	s.Require().NoError(s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, addr, invalidCoins))

	// supplying should fail as we have not registered token "uabcd"
	err := s.app.LeverageKeeper.Supply(s.ctx, addr, invalidCoin)
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestSupply_Valid() {
	app, ctx := s.app, s.ctx

	addr := sdk.AccAddress([]byte("addr________________1234"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acc)

	// mint and send coins
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, initCoins))

	// supply asset
	err := s.app.LeverageKeeper.Supply(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	s.Require().NoError(err)

	// verify the total supply of the minted uToken
	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)
	supply := s.app.LeverageKeeper.GetUTokenSupply(ctx, uTokenDenom)
	expected := sdk.NewInt64Coin(uTokenDenom, 1000000000) // 1k u/umee
	s.Require().Equal(expected, supply)

	// verify the user's balances
	tokenBalance := app.BankKeeper.GetBalance(ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(initTokens.Sub(sdk.NewInt(1000000000)), tokenBalance.Amount)

	uTokenBalance := app.BankKeeper.GetBalance(ctx, addr, uTokenDenom)
	s.Require().Equal(int64(1000000000), uTokenBalance.Amount.Int64())
}

func (s *IntegrationTestSuite) TestWithdraw_Valid() {
	app, ctx := s.app, s.ctx

	addr := sdk.AccAddress([]byte("addr________________"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acc)

	// mint and send coins
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, initCoins))

	// supply asset
	err := s.app.LeverageKeeper.Supply(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	s.Require().NoError(err)

	// verify the total supply of the minted uToken
	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)
	supply := s.app.LeverageKeeper.GetUTokenSupply(ctx, uTokenDenom)
	expected := sdk.NewInt64Coin(uTokenDenom, 1000000000) // 1k u/umee
	s.Require().Equal(expected, supply)

	// withdraw the total amount of assets supplied
	uToken := expected
	err = s.app.LeverageKeeper.Withdraw(ctx, addr, uToken)
	s.Require().NoError(err)

	// verify total supply of the uTokens
	supply = s.app.LeverageKeeper.GetUTokenSupply(ctx, uTokenDenom)
	s.Require().Equal(int64(0), supply.Amount.Int64())

	// verify the user's balances
	tokenBalance := app.BankKeeper.GetBalance(ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(initTokens, tokenBalance.Amount)

	uTokenBalance := app.BankKeeper.GetBalance(ctx, addr, uTokenDenom)
	s.Require().Equal(int64(0), uTokenBalance.Amount.Int64())
}

func (s *IntegrationTestSuite) TestSetReserves() {
	// get initial reserves
	amount := s.app.LeverageKeeper.GetReserveAmount(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(amount, sdk.ZeroInt())

	// artifically reserve 200 umee
	err := s.tk.SetReserveAmount(s.ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	s.Require().NoError(err)

	// get new reserves
	amount = s.app.LeverageKeeper.GetReserveAmount(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(amount, sdk.NewInt(200000000))
}

func (s *IntegrationTestSuite) TestGetToken() {
	uabc := newToken("uabc", "ABC")
	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, uabc))

	t, err := s.app.LeverageKeeper.GetTokenSettings(s.ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(t.ReserveFactor, sdk.MustNewDecFromStr("0.2"))
	s.Require().Equal(t.CollateralWeight, sdk.MustNewDecFromStr("0.25"))
	s.Require().Equal(t.LiquidationThreshold, sdk.MustNewDecFromStr("0.25"))
	s.Require().Equal(t.BaseBorrowRate, sdk.MustNewDecFromStr("0.02"))
	s.Require().Equal(t.KinkBorrowRate, sdk.MustNewDecFromStr("0.22"))
	s.Require().Equal(t.MaxBorrowRate, sdk.MustNewDecFromStr("1.52"))
	s.Require().Equal(t.KinkUtilization, sdk.MustNewDecFromStr("0.8"))
	s.Require().Equal(t.LiquidationIncentive, sdk.MustNewDecFromStr("0.1"))

	s.Require().NoError(t.AssertBorrowEnabled())
	s.Require().NoError(t.AssertSupplyEnabled())
	s.Require().NoError(s.app.LeverageKeeper.AssertNotBlacklisted(s.ctx, "uabc"))
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
	err := s.app.LeverageKeeper.Supply(ctx, supplierAddr, supplyCoin)
	s.Require().NoError(err)

	// supplier enables u/umee as collateral
	collat, err := s.app.LeverageKeeper.ExchangeToken(ctx, supplyCoin)
	s.Require().NoError(err)
	err = s.app.LeverageKeeper.Collateralize(ctx, supplierAddr, collat)
	s.Require().NoError(err)

	// return the account addresses
	return supplierAddr, bumAddr
}

// mintAndSupplyAtom mints a amount of atoms to an address
// account has been created with no assets.
func (s *IntegrationTestSuite) mintAndSupplyAtom(mintTo sdk.AccAddress, amountToMint, amountToSupply int64) {
	app, ctx := s.app, s.ctx

	// mint and send atom to mint addr
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName,
		sdk.NewCoins(sdk.NewInt64Coin(atomIBCDenom, amountToMint)), // amountToMint Atom
	))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, mintTo,
		sdk.NewCoins(sdk.NewInt64Coin(atomIBCDenom, amountToMint)), // amountToMint Atom,
	))

	// user supplies amountToSupply atom and receives amountToSupply u/atom
	supplyCoin := sdk.NewInt64Coin(atomIBCDenom, amountToSupply)
	err := s.app.LeverageKeeper.Supply(ctx, mintTo, supplyCoin)
	s.Require().NoError(err)

	// user enables u/atom as collateral
	collat, err := s.app.LeverageKeeper.ExchangeToken(ctx, supplyCoin)
	s.Require().NoError(err)
	err = s.app.LeverageKeeper.Collateralize(ctx, mintTo, collat)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestBorrow_Invalid() {
	addr, _ := s.initBorrowScenario()

	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// user attempts to borrow 200 u/umee, fails because uTokens cannot be borrowed
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 200000000))
	s.Require().Error(err)

	// user attempts to borrow 200 abcd, fails because "abcd" is not a valid denom
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin("uabcd", 200000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestBorrow_InsufficientCollateral() {
	_, bumAddr := s.initBorrowScenario() // create initial conditions

	// The "bum" user from the init scenario is being used because it
	// possesses no assets or collateral.

	// bum attempts to borrow 200 umee, fails because of insufficient collateral
	err := s.app.LeverageKeeper.Borrow(s.ctx, bumAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestBorrow_InsufficientLendingPool() {
	// Any user from the init scenario can perform this test, because it errors on module balance
	addr, _ := s.initBorrowScenario()

	// user attempts to borrow 20000 umee, fails because of insufficient module account balance
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestRepay_Invalid() {
	// Any user from the init scenario can be used for this test.
	addr, _ := s.initBorrowScenario()

	// user attempts to repay 200 abcd, fails because "abcd" is not an accepted asset type
	_, err := s.app.LeverageKeeper.Repay(s.ctx, addr, sdk.NewInt64Coin("uabcd", 200000000))
	s.Require().Error(err)

	// user attempts to repay 200 u/umee, fails because utokens are not loanable assets
	_, err = s.app.LeverageKeeper.Repay(s.ctx, addr, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 200000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestBorrow_Valid() {
	addr, _ := s.initBorrowScenario()

	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// user borrows 20 umee
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000))
	s.Require().NoError(err)

	// verify user's new loan amount in the correct denom (20 umee)
	loanBalance := s.app.LeverageKeeper.GetBorrow(s.ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000))

	// verify user's total loan balance (sdk.Coins) is also 20 umee (no other coins present)
	totalLoanBalance := s.app.LeverageKeeper.GetBorrowerBorrows(s.ctx, addr)
	s.Require().Equal(totalLoanBalance, sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 20000000)))

	// verify user's new umee balance (10 - 1k from initial + 20 from loan = 9020 umee)
	tokenBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9020000000))

	// verify user's uToken balance remains at 0 u/umee from initial conditions
	uTokenBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 0))

	// verify user's uToken collateral remains at 1000 u/umee from initial conditions
	collateralBalance := s.app.LeverageKeeper.GetCollateralAmount(s.ctx, addr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(collateralBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))
}

func (s *IntegrationTestSuite) TestBorrow_BorrowLimit() {
	addr, _ := s.initBorrowScenario()

	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// determine an amount of umee to borrow, such that the user will be at about 90% of their borrow limit
	token, _ := s.app.LeverageKeeper.GetTokenSettings(s.ctx, umeeapp.BondDenom)
	uDenom := s.app.LeverageKeeper.FromTokenToUTokenDenom(s.ctx, umeeapp.BondDenom)
	collateral := s.app.LeverageKeeper.GetCollateralAmount(s.ctx, addr, uDenom)
	amountToBorrow := token.CollateralWeight.Mul(sdk.MustNewDecFromStr("0.9")).MulInt(collateral.Amount).TruncateInt()

	// user borrows umee up to 90% of their borrow limit
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewCoin(umeeapp.BondDenom, amountToBorrow))
	s.Require().NoError(err)

	// user tries to borrow the same amount again, fails due to borrow limit
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewCoin(umeeapp.BondDenom, amountToBorrow))
	s.Require().Error(err)

	// user tries to disable u/umee as collateral, fails due to borrow limit
	err = s.app.LeverageKeeper.Decollateralize(s.ctx, addr, sdk.NewCoin(uDenom, collateral.Amount))
	s.Require().Error(err)

	// user tries to withdraw all its u/umee, fails due to borrow limit
	err = s.app.LeverageKeeper.Withdraw(s.ctx, addr, sdk.NewCoin(uDenom, collateral.Amount))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestBorrow_Reserved() {
	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral.
	addr, _ := s.initBorrowScenario()

	// artifically reserve 200 umee
	err := s.tk.SetReserveAmount(s.ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	s.Require().NoError(err)

	// Note: Setting umee collateral weight to 1.0 to allow user to borrow heavily
	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("1.0")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("1.0")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// Supplier tries to borrow 1000 umee, insufficient balance because 200 of the
	// module's 1000 umee are reserved.
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000))
	s.Require().Error(err)

	// user borrows 800 umee
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 800000000))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestRepay_Valid() {
	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral.
	addr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// user borrows 20 umee
	err := s.app.LeverageKeeper.Borrow(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000))
	s.Require().NoError(err)

	// user repays 8 umee
	repaid, err := s.app.LeverageKeeper.Repay(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 8000000))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(8000000), repaid)

	// verify user's new loan amount (12 umee)
	loanBalance := s.app.LeverageKeeper.GetBorrow(ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 12000000))

	// verify user's new umee balance (10 - 1k from initial + 20 from loan - 8 repaid = 9012 umee)
	tokenBalance := app.BankKeeper.GetBalance(ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9012000000))

	// verify user's uToken balance remains at 0 u/umee from initial conditions
	uTokenBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 0))

	// verify user's uToken collateral remains at 1000 u/umee from initial conditions
	collateralBalance := s.app.LeverageKeeper.GetCollateralAmount(s.ctx, addr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(collateralBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))

	// user repays 12 umee (loan repaid in full)
	repaid, err = s.app.LeverageKeeper.Repay(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 12000000))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(12000000), repaid)

	// verify user's new loan amount in the correct denom (zero)
	loanBalance = s.app.LeverageKeeper.GetBorrow(ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 0))

	// verify user's new umee balance (10 - 1k from initial + 20 from loan - 20 repaid = 9000 umee)
	tokenBalance = app.BankKeeper.GetBalance(ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9000000000))

	// verify user's uToken balance remains at 0 u/umee from initial conditions
	uTokenBalance = s.app.BankKeeper.GetBalance(s.ctx, addr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 0))

	// verify user's uToken collateral remains at 1000 u/umee from initial conditions
	collateralBalance = s.app.LeverageKeeper.GetCollateralAmount(s.ctx, addr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(collateralBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))
}

func (s *IntegrationTestSuite) TestRepay_Overpay() {
	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral.
	addr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// user borrows 20 umee
	err := s.app.LeverageKeeper.Borrow(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000))
	s.Require().NoError(err)

	// user repays 30 umee - should automatically reduce to 20 (the loan amount) and succeed
	coinToRepay := sdk.NewInt64Coin(umeeapp.BondDenom, 30000000)
	repaid, err := s.app.LeverageKeeper.Repay(ctx, addr, coinToRepay)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(20000000), repaid)

	// verify that coinToRepay has not been modified
	s.Require().Equal(sdk.NewInt(30000000), coinToRepay.Amount)

	// verify user's new loan amount is 0 umee
	loanBalance := s.app.LeverageKeeper.GetBorrow(ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 0))

	// verify user's new umee balance (10 - 1k from initial + 20 from loan - 20 repaid = 9000 umee)
	tokenBalance := app.BankKeeper.GetBalance(ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9000000000))

	// verify user's uToken balance remains at 0 u/umee from initial conditions
	uTokenBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 0))

	// verify user's uToken collateral remains at 1000 u/umee from initial conditions
	collateralBalance := s.app.LeverageKeeper.GetCollateralAmount(s.ctx, addr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(collateralBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))

	// user repays 50 umee - this time it fails because the loan no longer exists
	_, err = s.app.LeverageKeeper.Repay(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 50000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestLiqudateBorrow_Valid() {
	addr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral.

	// user borrows 90 umee
	err := s.app.LeverageKeeper.Borrow(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 90000000))
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

	// liquidator attempts to liquidate user, but user is ineligible (not over borrow limit)
	// liquidator does not specify a minimum reward (hence 0 u/umee)
	repayment := sdk.NewInt64Coin(umeeapp.BondDenom, 30000000) // 30 umee
	rewardDenom := s.app.LeverageKeeper.FromTokenToUTokenDenom(ctx, umeeapp.BondDenom)
	unrestrictedReward := sdk.NewInt64Coin(umeeapp.BondDenom, 0) // 0 umee (rewardDenom = u/umee)
	_, _, err = s.app.LeverageKeeper.LiquidateBorrow(ctx, liquidatorAddr, addr, repayment, unrestrictedReward)
	s.Require().Error(err)

	// Note: Setting umee collateral weight to 0.0 to allow liquidation
	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("0")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("0")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// liquidator attempts to liquidate user, but specifies too high of a minimum reward
	repayment = sdk.NewInt64Coin(umeeapp.BondDenom, 10000000)        // 10 umee
	excessiveReward := sdk.NewInt64Coin(umeeapp.BondDenom, 20000000) // 20 umee (rewardDenom = u/umee)
	_, _, err = s.app.LeverageKeeper.LiquidateBorrow(ctx, liquidatorAddr, addr, repayment, excessiveReward)
	s.Require().Error(err)

	// liquidator partially liquidates user, receiving some collateral
	repayment = sdk.NewInt64Coin(umeeapp.BondDenom, 10000000) // 10 umee
	repaid, reward, err := s.app.LeverageKeeper.LiquidateBorrow(ctx, liquidatorAddr, addr, repayment, unrestrictedReward)
	s.Require().NoError(err)
	s.Require().Equal(repayment.Amount, repaid)
	s.Require().Equal(sdk.NewInt(11000000), reward)

	// verify user's new loan amount is 80 umee (still over borrow limit)
	loanBalance := s.app.LeverageKeeper.GetBorrow(ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance.String(), sdk.NewInt64Coin(umeeapp.BondDenom, 80000000).String())

	// verify liquidator's new u/umee balance = 11 = (10 + liquidation incentive)
	uTokenBalance := app.BankKeeper.GetBalance(ctx, liquidatorAddr, rewardDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin(rewardDenom, 11000000))

	// verify liquidator's new umee balance (10k - 10) = 9990 umee
	tokenBalance := app.BankKeeper.GetBalance(ctx, liquidatorAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9990000000))

	// liquidator fully liquidates user, receiving more collateral and reducing borrowed amount to zero
	repayment = sdk.NewInt64Coin(umeeapp.BondDenom, 300000000) // 300 umee
	repaid, reward, err = s.app.LeverageKeeper.LiquidateBorrow(ctx, liquidatorAddr, addr, repayment, unrestrictedReward)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(80000000), repaid)
	s.Require().Equal(sdk.NewInt(88000000), reward)

	// verify that repayment has not been modified
	s.Require().Equal(sdk.NewInt(300000000), repayment.Amount)

	// verify liquidator's new u/umee balance = 99 = (90 + liquidation incentive)
	uTokenBalance = app.BankKeeper.GetBalance(ctx, liquidatorAddr, rewardDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin(rewardDenom, 99000000))

	// verify user's new loan amount is zero
	loanBalance = s.app.LeverageKeeper.GetBorrow(ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 0))

	// verify liquidator's new umee balance (10k - 90) = 9910 umee
	tokenBalance = app.BankKeeper.GetBalance(ctx, liquidatorAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9910000000))
}

func (s *IntegrationTestSuite) TestRepayBadDebt() {
	// Creating a supplier so module account has some uumee
	_ = s.setupAccount(umeeDenom, 200000000, 200000000, 0, false) // 200 umee

	// Using an address with no assets
	addr := s.setupAccount(umeeDenom, 0, 0, 0, false)

	// Create an uncollateralized debt position
	badDebt := sdk.NewInt64Coin(umeeDenom, 100000000) // 100 umee
	err := s.tk.SetBorrow(s.ctx, addr, badDebt)
	s.Require().NoError(err)

	// Manually mark the bad debt for repayment
	s.Require().NoError(s.tk.SetBadDebtAddress(s.ctx, addr, umeeDenom, true))

	// Manually set reserves to 60 umee
	reserve := sdk.NewInt64Coin(umeeDenom, 60000000)
	err = s.tk.SetReserveAmount(s.ctx, reserve)
	s.Require().NoError(err)

	// Sweep all bad debts, which should repay 60 umee of the bad debt (partial repayment)
	err = s.app.LeverageKeeper.SweepBadDebts(s.ctx)
	s.Require().NoError(err)

	// Confirm that a debt of 40 umee remains
	remainingDebt := s.app.LeverageKeeper.GetBorrow(s.ctx, addr, umeeDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 40000000), remainingDebt)

	// Confirm that reserves are exhausted
	remainingReserve := s.app.LeverageKeeper.GetReserveAmount(s.ctx, umeeDenom)
	s.Require().Equal(sdk.ZeroInt(), remainingReserve)

	// Manually set reserves to 70 umee
	reserve = sdk.NewInt64Coin(umeeDenom, 70000000)
	err = s.tk.SetReserveAmount(s.ctx, reserve)
	s.Require().NoError(err)

	// Sweep all bad debts, which should fully repay the bad debt this time
	err = s.app.LeverageKeeper.SweepBadDebts(s.ctx)
	s.Require().NoError(err)

	// Confirm that the debt is eliminated
	remainingDebt = s.app.LeverageKeeper.GetBorrow(s.ctx, addr, umeeDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 0), remainingDebt)

	// Confirm that reserves are now at 30 umee
	remainingReserve = s.app.LeverageKeeper.GetReserveAmount(s.ctx, umeeDenom)
	s.Require().Equal(sdk.NewInt(30000000), remainingReserve)

	// Sweep all bad debts - but there are none
	err = s.app.LeverageKeeper.SweepBadDebts(s.ctx)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestDeriveExchangeRate() {
	// The init scenario is being used so module balance starts at 1000 umee
	// and the uToken supply starts at 1000 due to supplier account
	_, addr := s.initBorrowScenario()

	// artificially increase total borrows (by affecting a single address)
	err := s.tk.SetBorrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 2000000000)) // 2000 umee
	s.Require().NoError(err)

	// artificially set reserves
	err = s.tk.SetReserveAmount(s.ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 300000000)) // 300 umee
	s.Require().NoError(err)

	// expected token:uToken exchange rate
	//    = (total borrows + module balance - reserves) / utoken supply
	//    = 2000 + 1000 - 300 / 1000
	//    = 2.7

	// get derived exchange rate
	rate := s.app.LeverageKeeper.DeriveExchangeRate(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(sdk.MustNewDecFromStr("2.7"), rate)
}

func (s *IntegrationTestSuite) TestAccrueZeroInterest() {
	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral.
	addr, _ := s.initBorrowScenario()

	// user borrows 40 umee
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 40000000))
	s.Require().NoError(err)

	// verify user's loan amount (40 umee)
	loanBalance := s.app.LeverageKeeper.GetBorrow(s.ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 40000000))

	// Because no time has passed since genesis (due to test setup) this will not
	// increase borrowed amount.
	err = s.app.LeverageKeeper.AccrueAllInterest(s.ctx)
	s.Require().NoError(err)

	// verify user's loan amount (40 umee)
	loanBalance = s.app.LeverageKeeper.GetBorrow(s.ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 40000000))

	// borrow APY at utilization = 4%
	// when kink utilization = 80%, and base/kink APY are 0.02 and 0.22
	borrowAPY := s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(sdk.MustNewDecFromStr("0.03"), borrowAPY)

	// supply APY when borrow APY is 3%
	// and utilization is 4%, and reservefactor is 20%, and OracleRewardFactor is 1%
	// 0.03 * 0.04 * (1 - 0.21) = 0.000948
	supplyAPY := s.app.LeverageKeeper.DeriveSupplyAPY(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("0.000948"), supplyAPY)
}

func (s *IntegrationTestSuite) TestDynamicInterest() {
	// Init scenario is being used because the module account (lending pool)
	// already has 1000 umee.
	addr, _ := s.initBorrowScenario()

	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("1.0")     // to allow high utilization
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("1.0") // to allow high utilization

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// Base interest rate (0% utilization)
	rate := s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.02"))

	// user borrows 200 umee, utilization 200/1000
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	s.Require().NoError(err)

	// Between base interest and kink (20% utilization)
	rate = s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.07"))

	// user borrows 600 more umee, utilization 800/1000
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 600000000))
	s.Require().NoError(err)

	// Kink interest rate (80% utilization)
	rate = s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.22"))

	// user borrows 100 more umee, utilization 900/1000
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

	// Between kink interest and max (90% utilization)
	rate = s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.87"))

	// user borrows 100 more umee, utilization 1000/1000
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

	// Max interest rate (100% utilization)
	rate = s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("1.52"))
}

func (s *IntegrationTestSuite) TestDynamicInterest_InvalidAsset() {
	rate := s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, "uabc")
	s.Require().Equal(rate, sdk.ZeroDec())
}

func (s *IntegrationTestSuite) TestGetEligibleLiquidationTargets_OneAddrOneAsset() {
	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee enabled as collateral.
	addr, _ := s.initBorrowScenario()

	// user borrows 100 umee (max current allowed) user amount enabled as collateral * CollateralWeight
	// = 1000 * 0.1
	// = 100
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

	zeroAddresses, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{}, zeroAddresses)

	// Note: Setting umee liquidation threshold to 0.05 to make the user eligible to liquidation
	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("0.05")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.05")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	targetAddress, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{addr}, targetAddress)

	// if it tries to borrow any other asset it should return an error
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(atomIBCDenom, 1))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestGetEligibleLiquidationTargets_OneAddrTwoAsset() {
	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee enabled as collateral.
	addr, _ := s.initBorrowScenario()

	// user borrows 100 umee (max current allowed) user amount enabled as collateral * CollateralWeight
	// = 1000 * 0.1
	// = 100
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

	zeroAddresses, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{}, zeroAddresses)

	mintAmountAtom := int64(100000000)  // 100 atom
	supplyAmountAtom := int64(50000000) // 50 atom

	// mints and send to user 100 atom and already
	// enable 50 u/atom as collateral.
	s.mintAndSupplyAtom(addr, mintAmountAtom, supplyAmountAtom)

	// user borrows 4 atom (max current allowed - 1) user amount enabled as collateral * CollateralWeight
	// = (50 * 0.1) - 1
	// = 4
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(atomIBCDenom, 4000000)) // 4 atom
	s.Require().NoError(err)

	// Note: Setting umee liquidation threshold to 0.05 to make the user eligible for liquidation
	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("0.05")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.05")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// Note: Setting atom collateral weight to 0.01 to make the user eligible for liquidation
	atomIBCToken := newToken(atomIBCDenom, "ATOM")
	atomIBCToken.CollateralWeight = sdk.MustNewDecFromStr("0.01")
	atomIBCToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.01")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, atomIBCToken))

	targetAddr, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{addr}, targetAddr)
}

func (s *IntegrationTestSuite) TestGetEligibleLiquidationTargets_TwoAddr() {
	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee enabled as collateral.
	supplierAddr, anotherSupplier := s.initBorrowScenario()

	// supplier borrows 100 umee (max current allowed) supplier amount enabled as collateral * CollateralWeight
	// = 1000 * 0.1
	// = 100
	err := s.app.LeverageKeeper.Borrow(s.ctx, supplierAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

	zeroAddresses, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{}, zeroAddresses)

	mintAmountAtom := int64(100000000)  // 100 atom
	supplyAmountAtom := int64(50000000) // 50 atom

	// mints and send to anotherSupplier 100 atom and already
	// enable 50 u/atom as collateral.
	s.mintAndSupplyAtom(anotherSupplier, mintAmountAtom, supplyAmountAtom)

	// anotherSupplier borrows 4 atom (max current allowed - 1) anotherSupplier amount enabled as collateral * CollateralWeight
	// = (50 * 0.1) - 1
	// = 4
	err = s.app.LeverageKeeper.Borrow(s.ctx, anotherSupplier, sdk.NewInt64Coin(atomIBCDenom, 4000000)) // 4 atom
	s.Require().NoError(err)

	// Note: Setting umee liquidation threshold to 0.05 to make the supplier eligible for liquidation
	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("0.05")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.05")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// Note: Setting atom collateral weight to 0.01 to make the supplier eligible for liquidation
	atomIBCToken := newToken(atomIBCDenom, "ATOM")
	atomIBCToken.CollateralWeight = sdk.MustNewDecFromStr("0.01")
	atomIBCToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.01")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, atomIBCToken))

	supplierAddress, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{supplierAddr, anotherSupplier}, supplierAddress)
}

func (s *IntegrationTestSuite) TestReserveAmountInvariant() {
	// artificially set reserves
	err := s.tk.SetReserveAmount(s.ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 300000000)) // 300 umee
	s.Require().NoError(err)

	// check invariant
	_, broken := keeper.ReserveAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)
}

func (s *IntegrationTestSuite) TestCollateralAmountInvariant() {
	addr, _ := s.initBorrowScenario()

	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// check invariant
	_, broken := keeper.CollateralAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)

	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)

	// withdraw the supplyed umee in the initBorrowScenario
	err := s.app.LeverageKeeper.Withdraw(s.ctx, addr, sdk.NewInt64Coin(uTokenDenom, 1000000000))
	s.Require().NoError(err)

	// check invariant
	_, broken = keeper.CollateralAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)
}

func (s *IntegrationTestSuite) TestBorrowAmountInvariant() {
	addr, _ := s.initBorrowScenario()

	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// user borrows 20 umee
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000))
	s.Require().NoError(err)

	// check invariant
	_, broken := keeper.BorrowAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)

	// user repays 30 umee, actually only 20 because is the min between
	// the amount borrowed and the amount repaid
	_, err = s.app.LeverageKeeper.Repay(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 30000000))
	s.Require().NoError(err)

	// check invariant
	_, broken = keeper.BorrowAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)
}

func (s *IntegrationTestSuite) TestWithdraw_InsufficientCollateral() {
	// Create a supplier with 1 u/umee collateral by supplying 1 umee
	supplierAddr := s.setupAccount(umeeapp.BondDenom, 1000000, 1000000, 0, true)

	// Create an additional supplier so lending pool has extra umee
	_ = s.setupAccount(umeeapp.BondDenom, 1000000, 1000000, 0, true)

	// verify collateral amount and total supply of minted uTokens
	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)
	collateral := s.app.LeverageKeeper.GetCollateralAmount(s.ctx, supplierAddr, uTokenDenom)
	s.Require().Equal(sdk.NewInt64Coin(uTokenDenom, 1000000), collateral) // 1 u/umee
	supply := s.app.LeverageKeeper.GetUTokenSupply(s.ctx, uTokenDenom)
	s.Require().Equal(sdk.NewInt64Coin(uTokenDenom, 2000000), supply) // 2 u/umee

	// withdraw more collateral than available
	uToken := collateral.Add(sdk.NewInt64Coin(uTokenDenom, 1))
	err := s.app.LeverageKeeper.Withdraw(s.ctx, supplierAddr, uToken)
	s.Require().EqualError(err, "1000001u/uumee: insufficient balance")
}

func (s *IntegrationTestSuite) TestTotalCollateral() {
	// Test zero collateral
	uDenom := types.UTokenFromTokenDenom(umeeDenom)
	collateral := s.app.LeverageKeeper.GetTotalCollateral(s.ctx, uDenom)
	s.Require().Equal(sdk.ZeroInt(), collateral)

	// Uses borrow scenario, because supplier possesses collateral
	_, _ = s.initBorrowScenario()

	// Test nonzero collateral
	collateral = s.app.LeverageKeeper.GetTotalCollateral(s.ctx, uDenom)
	s.Require().Equal(sdk.NewInt(1000000000), collateral)
}
