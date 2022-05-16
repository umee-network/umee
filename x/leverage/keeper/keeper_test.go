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

	umeeToken := types.Token{
		BaseDenom:            umeeapp.BondDenom,
		ReserveFactor:        sdk.MustNewDecFromStr("0.20"),
		CollateralWeight:     sdk.MustNewDecFromStr("0.25"),
		LiquidationThreshold: sdk.MustNewDecFromStr("0.25"),
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.22"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("1.52"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:          "UMEE",
		Exponent:             6,
		EnableLend:           true,
		EnableBorrow:         true,
		Blacklist:            false,
	}
	atomIBCToken := types.Token{
		BaseDenom:            atomIBCDenom,
		ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:     sdk.MustNewDecFromStr("0.5"),
		LiquidationThreshold: sdk.MustNewDecFromStr("0.5"),
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.2"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:          "ATOM",
		Exponent:             6,
		EnableLend:           true,
		EnableBorrow:         true,
		Blacklist:            false,
	}

	// we only override the Leverage keeper so we can supply a custom mock oracle
	k, tk := keeper.NewTestKeeper(
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
	s.Require().NoError(app.LeverageKeeper.SetRegisteredToken(ctx, umeeToken))
	s.Require().NoError(app.LeverageKeeper.SetRegisteredToken(ctx, atomIBCToken))

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(app.LeverageKeeper))

	s.app = app
	s.ctx = ctx
	s.queryClient = types.NewQueryClient(queryHelper)
}

// setupAccount executes some common boilerplate before a test, where a lender account is given tokens of a given denom,
// may also lend them to receive uTokens, and may also enable those uTokens as collateral and borrow tokens in the same denom.
func (s *IntegrationTestSuite) setupAccount(denom string, mintAmount, lendAmount, borrowAmount int64, collateral bool) sdk.AccAddress {
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

	if lendAmount > 0 {
		// account lends lendAmount tokens and receives uTokens
		err := s.app.LeverageKeeper.LendAsset(s.ctx, addr, sdk.NewInt64Coin(denom, lendAmount))
		s.Require().NoError(err)
	}

	if collateral {
		// account enables associated uToken as collateral
		collatDenom := s.app.LeverageKeeper.FromTokenToUTokenDenom(s.ctx, denom)
		err := s.app.LeverageKeeper.SetCollateralSetting(s.ctx, addr, collatDenom, true)
		s.Require().NoError(err)
	}

	if borrowAmount > 0 {
		// account borrows borrowAmount tokens
		err := s.app.LeverageKeeper.BorrowAsset(s.ctx, addr, sdk.NewInt64Coin(denom, borrowAmount))
		s.Require().NoError(err)
	}

	// return the account addresse
	return addr
}

func (s *IntegrationTestSuite) TestLendAsset_InvalidAsset() {
	lenderAddr := sdk.AccAddress([]byte("addr________________"))
	lenderAcc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, lenderAddr)
	s.app.AccountKeeper.SetAccount(s.ctx, lenderAcc)

	// create coins of an unregistered base asset type "uabcd"
	invalidCoin := sdk.NewInt64Coin("uabcd", 1000000000) // 1k abcd
	invalidCoins := sdk.NewCoins(invalidCoin)

	// mint and send coins
	s.Require().NoError(s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, invalidCoins))
	s.Require().NoError(s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, lenderAddr, invalidCoins))

	// lending should fail as we have not registered token "uabcd"
	err := s.app.LeverageKeeper.LendAsset(s.ctx, lenderAddr, invalidCoin)
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestLendAsset_Valid() {
	app, ctx := s.app, s.ctx

	lenderAddr := sdk.AccAddress([]byte("addr________________1234"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)

	// mint and send coins
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr, initCoins))

	// lend asset
	err := s.app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	s.Require().NoError(err)

	// verify the total supply of the minted uToken
	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)
	supply := s.app.LeverageKeeper.GetUTokenSupply(ctx, uTokenDenom)
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

	lenderAddr := sdk.AccAddress([]byte("addr________________"))
	lenderAcc := app.AccountKeeper.NewAccountWithAddress(ctx, lenderAddr)
	app.AccountKeeper.SetAccount(ctx, lenderAcc)

	// mint and send coins
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, lenderAddr, initCoins))

	// lend asset
	err := s.app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000)) // 1k umee
	s.Require().NoError(err)

	// verify the total supply of the minted uToken
	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)
	supply := s.app.LeverageKeeper.GetUTokenSupply(ctx, uTokenDenom)
	expected := sdk.NewInt64Coin(uTokenDenom, 1000000000) // 1k u/umee
	s.Require().Equal(expected, supply)

	// withdraw the total amount of assets loaned
	uToken := expected
	err = s.app.LeverageKeeper.WithdrawAsset(ctx, lenderAddr, uToken)
	s.Require().NoError(err)

	// verify total supply of the uTokens
	supply = s.app.LeverageKeeper.GetUTokenSupply(ctx, uTokenDenom)
	s.Require().Equal(int64(0), supply.Amount.Int64())

	// verify the lender's balances
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(initTokens, tokenBalance.Amount)

	uTokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, uTokenDenom)
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
	uabc := types.Token{
		BaseDenom:            "uabc",
		ReserveFactor:        sdk.MustNewDecFromStr("0.1"),
		CollateralWeight:     sdk.MustNewDecFromStr("0.2"),
		LiquidationThreshold: sdk.MustNewDecFromStr("0.2"),
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.3"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.4"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("0.5"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.6"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:          "ABC",
		Exponent:             6,
		EnableLend:           true,
		EnableBorrow:         true,
		Blacklist:            false,
	}
	s.Require().NoError(s.app.LeverageKeeper.SetRegisteredToken(s.ctx, uabc))

	reserveFactor, err := s.app.LeverageKeeper.GetReserveFactor(s.ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(reserveFactor, sdk.MustNewDecFromStr("0.1"))

	collateralWeight, err := s.app.LeverageKeeper.GetCollateralWeight(s.ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(collateralWeight, sdk.MustNewDecFromStr("0.2"))

	liquidationThreshold, err := s.app.LeverageKeeper.GetLiquidationThreshold(s.ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(liquidationThreshold, sdk.MustNewDecFromStr("0.2"))

	baseBorrowRate, err := s.app.LeverageKeeper.GetInterestBase(s.ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(baseBorrowRate, sdk.MustNewDecFromStr("0.3"))

	kinkBorrowRate, err := s.app.LeverageKeeper.GetInterestAtKink(s.ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(kinkBorrowRate, sdk.MustNewDecFromStr("0.4"))

	maxBorrowRate, err := s.app.LeverageKeeper.GetInterestMax(s.ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(maxBorrowRate, sdk.MustNewDecFromStr("0.5"))

	kinkUtilizationRate, err := s.app.LeverageKeeper.GetInterestKinkUtilization(s.ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(kinkUtilizationRate, sdk.MustNewDecFromStr("0.6"))
}

// initialize the common starting scenario from which borrow and repay tests stem:
// Umee and u/umee are registered assets; a "lender" account has 9k umee and 1k u/umee;
// the leverage module has 1k umee in its lending pool (module account); and a "bum"
// account has been created with no assets.
func (s *IntegrationTestSuite) initBorrowScenario() (lender, bum sdk.AccAddress) {
	app, ctx := s.app, s.ctx

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
	err := s.app.LeverageKeeper.LendAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000))
	s.Require().NoError(err)

	// lender enables u/umee as collateral
	collatDenom := s.app.LeverageKeeper.FromTokenToUTokenDenom(ctx, umeeapp.BondDenom)
	err = s.app.LeverageKeeper.SetCollateralSetting(ctx, lenderAddr, collatDenom, true)
	s.Require().NoError(err)

	// return the account addresses
	return lenderAddr, bumAddr
}

// mintAndLendAtom mints a amount of atoms to an address
// account has been created with no assets.
func (s *IntegrationTestSuite) mintAndLendAtom(mintTo sdk.AccAddress, amountToMint, amountToLend int64) {
	app, ctx := s.app, s.ctx

	// mint and send atom to mint addr
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName,
		sdk.NewCoins(sdk.NewInt64Coin(atomIBCDenom, amountToMint)), // amountToMint Atom
	))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, mintTo,
		sdk.NewCoins(sdk.NewInt64Coin(atomIBCDenom, amountToMint)), // amountToMint Atom,
	))

	// lender lends amountToLend atom and receives amountToLend u/atom
	err := s.app.LeverageKeeper.LendAsset(ctx, mintTo, sdk.NewInt64Coin(atomIBCDenom, amountToLend))
	s.Require().NoError(err)

	// lender enables u/atom as collateral
	collatDenom := s.app.LeverageKeeper.FromTokenToUTokenDenom(ctx, atomIBCDenom)
	err = s.app.LeverageKeeper.SetCollateralSetting(ctx, mintTo, collatDenom, true)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestBorrowAsset_Invalid() {
	lenderAddr, _ := s.initBorrowScenario()

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender attempts to borrow 200 u/umee, fails because uTokens cannot be borrowed
	err := s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 200000000))
	s.Require().Error(err)

	// lender attempts to borrow 200 abcd, fails because "abcd" is not a valid denom
	err = s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin("uabcd", 200000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestBorrowAsset_InsufficientCollateral() {
	_, bumAddr := s.initBorrowScenario() // create initial conditions

	// The "bum" user from the init scenario is being used because it
	// possesses no assets or collateral.

	// bum attempts to borrow 200 umee, fails because of insufficient collateral
	err := s.app.LeverageKeeper.BorrowAsset(s.ctx, bumAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestBorrowAsset_InsufficientLendingPool() {
	// Any user from the init scenario can perform this test, because it errors on module balance
	lenderAddr, _ := s.initBorrowScenario()

	// lender attempts to borrow 20000 umee, fails because of insufficient module account balance
	err := s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestRepayAsset_Invalid() {
	// Any user from the init scenario can be used for this test.
	lenderAddr, _ := s.initBorrowScenario()

	// lender attempts to repay 200 abcd, fails because "abcd" is not an accepted asset type
	_, err := s.app.LeverageKeeper.RepayAsset(s.ctx, lenderAddr, sdk.NewInt64Coin("uabcd", 200000000))
	s.Require().Error(err)

	// lender attempts to repay 200 u/umee, fails because utokens are not loanable assets
	_, err = s.app.LeverageKeeper.RepayAsset(s.ctx, lenderAddr, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 200000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestBorrowAsset_Valid() {
	lenderAddr, _ := s.initBorrowScenario()

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 20 umee
	err := s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000))
	s.Require().NoError(err)

	// verify lender's new loan amount in the correct denom (20 umee)
	loanBalance := s.app.LeverageKeeper.GetBorrow(s.ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000))

	// verify lender's total loan balance (sdk.Coins) is also 20 umee (no other coins present)
	totalLoanBalance := s.app.LeverageKeeper.GetBorrowerBorrows(s.ctx, lenderAddr)
	s.Require().Equal(totalLoanBalance, sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 20000000)))

	// verify lender's new umee balance (10 - 1k from initial + 20 from loan = 9020 umee)
	tokenBalance := s.app.BankKeeper.GetBalance(s.ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9020000000))

	// verify lender's uToken balance remains at 0 u/umee from initial conditions
	uTokenBalance := s.app.BankKeeper.GetBalance(s.ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 0))

	// verify lender's uToken collateral remains at 1000 u/umee from initial conditions
	collateralBalance := s.app.LeverageKeeper.GetCollateralAmount(s.ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(collateralBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))
}

func (s *IntegrationTestSuite) TestBorrowAsset_BorrowLimit() {
	lenderAddr, _ := s.initBorrowScenario()

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// determine an amount of umee to borrow, such that the lender will be at about 90% of their borrow limit
	token, _ := s.app.LeverageKeeper.GetRegisteredToken(s.ctx, umeeapp.BondDenom)
	uDenom := s.app.LeverageKeeper.FromTokenToUTokenDenom(s.ctx, umeeapp.BondDenom)
	collateral := s.app.LeverageKeeper.GetCollateralAmount(s.ctx, lenderAddr, uDenom)
	amountToBorrow := token.CollateralWeight.Mul(sdk.MustNewDecFromStr("0.9")).MulInt(collateral.Amount).TruncateInt()

	// lender borrows umee up to 90% of their borrow limit
	err := s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewCoin(umeeapp.BondDenom, amountToBorrow))
	s.Require().NoError(err)

	// lender tries to borrow the same amount again, fails due to borrow limit
	err = s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewCoin(umeeapp.BondDenom, amountToBorrow))
	s.Require().Error(err)

	// lender tries to disable u/umee as collateral, fails due to borrow limit
	err = s.app.LeverageKeeper.SetCollateralSetting(s.ctx, lenderAddr, uDenom, false)
	s.Require().Error(err)

	// lender tries to withdraw all its u/umee, fails due to borrow limit
	err = s.app.LeverageKeeper.WithdrawAsset(s.ctx, lenderAddr, sdk.NewCoin(uDenom, collateral.Amount))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestBorrowAsset_Reserved() {
	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral.
	lenderAddr, _ := s.initBorrowScenario()

	// artifically reserve 200 umee
	err := s.tk.SetReserveAmount(s.ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	s.Require().NoError(err)

	// Note: Setting umee collateral weight to 1.0 to allow lender to borrow heavily
	umeeToken := types.Token{
		BaseDenom:            umeeapp.BondDenom,
		ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:     sdk.MustNewDecFromStr("1.0"),
		LiquidationThreshold: sdk.MustNewDecFromStr("1.0"),
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.2"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:          "UMEE",
		Exponent:             6,
		EnableLend:           true,
		EnableBorrow:         true,
		Blacklist:            false,
	}
	s.Require().NoError(s.app.LeverageKeeper.SetRegisteredToken(s.ctx, umeeToken))

	// Lender tries to borrow 1000 umee, insufficient balance because 200 of the
	// module's 1000 umee are reserved.
	err = s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 1000000000))
	s.Require().Error(err)

	// lender borrows 800 umee
	err = s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 800000000))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestRepayAsset_Valid() {
	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral.
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// lender borrows 20 umee
	err := s.app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000))
	s.Require().NoError(err)

	// lender repays 8 umee
	repaid, err := s.app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 8000000))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(8000000), repaid)

	// verify lender's new loan amount (12 umee)
	loanBalance := s.app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 12000000))

	// verify lender's new umee balance (10 - 1k from initial + 20 from loan - 8 repaid = 9012 umee)
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9012000000))

	// verify lender's uToken balance remains at 0 u/umee from initial conditions
	uTokenBalance := s.app.BankKeeper.GetBalance(s.ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 0))

	// verify lender's uToken collateral remains at 1000 u/umee from initial conditions
	collateralBalance := s.app.LeverageKeeper.GetCollateralAmount(s.ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(collateralBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))

	// lender repays 12 umee (loan repaid in full)
	repaid, err = s.app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 12000000))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(12000000), repaid)

	// verify lender's new loan amount in the correct denom (zero)
	loanBalance = s.app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 0))

	// verify lender's new umee balance (10 - 1k from initial + 20 from loan - 20 repaid = 9000 umee)
	tokenBalance = app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9000000000))

	// verify lender's uToken balance remains at 0 u/umee from initial conditions
	uTokenBalance = s.app.BankKeeper.GetBalance(s.ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 0))

	// verify lender's uToken collateral remains at 1000 u/umee from initial conditions
	collateralBalance = s.app.LeverageKeeper.GetCollateralAmount(s.ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(collateralBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))
}

func (s *IntegrationTestSuite) TestRepayAsset_Overpay() {
	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral.
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// lender borrows 20 umee
	err := s.app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000))
	s.Require().NoError(err)

	// lender repays 30 umee - should automatically reduce to 20 (the loan amount) and succeed
	coinToRepay := sdk.NewInt64Coin(umeeapp.BondDenom, 30000000)
	repaid, err := s.app.LeverageKeeper.RepayAsset(ctx, lenderAddr, coinToRepay)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(20000000), repaid)

	// verify that coinToRepay has not been modified
	s.Require().Equal(sdk.NewInt(30000000), coinToRepay.Amount)

	// verify lender's new loan amount is 0 umee
	loanBalance := s.app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 0))

	// verify lender's new umee balance (10 - 1k from initial + 20 from loan - 20 repaid = 9000 umee)
	tokenBalance := app.BankKeeper.GetBalance(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9000000000))

	// verify lender's uToken balance remains at 0 u/umee from initial conditions
	uTokenBalance := s.app.BankKeeper.GetBalance(s.ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 0))

	// verify lender's uToken collateral remains at 1000 u/umee from initial conditions
	collateralBalance := s.app.LeverageKeeper.GetCollateralAmount(s.ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(collateralBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))

	// lender repays 50 umee - this time it fails because the loan no longer exists
	_, err = s.app.LeverageKeeper.RepayAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 50000000))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestLiqudateBorrow_Valid() {
	lenderAddr, _ := s.initBorrowScenario()
	app, ctx := s.app, s.ctx

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral.

	// lender borrows 90 umee
	err := s.app.LeverageKeeper.BorrowAsset(ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 90000000))
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
	// liquidator does not specify a minimum reward (hence 0 u/umee)
	repayment := sdk.NewInt64Coin(umeeapp.BondDenom, 30000000) // 30 umee
	rewardDenom := s.app.LeverageKeeper.FromTokenToUTokenDenom(ctx, umeeapp.BondDenom)
	unrestrictedReward := sdk.NewInt64Coin(umeeapp.BondDenom, 0) // 0 umee (rewardDenom = u/umee)
	_, _, err = s.app.LeverageKeeper.LiquidateBorrow(ctx, liquidatorAddr, lenderAddr, repayment, unrestrictedReward)
	s.Require().Error(err)

	// Note: Setting umee collateral weight to 0.0 to allow liquidation
	umeeToken := types.Token{
		BaseDenom:            umeeapp.BondDenom,
		ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:     sdk.MustNewDecFromStr("0.0"),
		LiquidationThreshold: sdk.MustNewDecFromStr("0.0"),
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.2"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:          "UMEE",
		Exponent:             6,
		EnableLend:           true,
		EnableBorrow:         true,
		Blacklist:            false,
	}
	s.Require().NoError(s.app.LeverageKeeper.SetRegisteredToken(s.ctx, umeeToken))

	// liquidator attempts to liquidate lender, but specifies too high of a minimum reward
	repayment = sdk.NewInt64Coin(umeeapp.BondDenom, 10000000)        // 10 umee
	excessiveReward := sdk.NewInt64Coin(umeeapp.BondDenom, 20000000) // 20 umee (rewardDenom = u/umee)
	repaid, reward, err := s.app.LeverageKeeper.LiquidateBorrow(ctx, liquidatorAddr, lenderAddr, repayment, excessiveReward)
	s.Require().Error(err)

	// liquidator partially liquidates lender, receiving some collateral
	repayment = sdk.NewInt64Coin(umeeapp.BondDenom, 10000000) // 10 umee
	repaid, reward, err = s.app.LeverageKeeper.LiquidateBorrow(ctx, liquidatorAddr, lenderAddr, repayment, unrestrictedReward)
	s.Require().NoError(err)
	s.Require().Equal(repayment.Amount, repaid)
	s.Require().Equal(sdk.NewInt(11000000), reward)

	// verify lender's new loan amount is 80 umee (still over borrow limit)
	loanBalance := s.app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance.String(), sdk.NewInt64Coin(umeeapp.BondDenom, 80000000).String())

	// verify liquidator's new u/umee balance = 11 = (10 + liquidation incentive)
	uTokenBalance := app.BankKeeper.GetBalance(ctx, liquidatorAddr, rewardDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin(rewardDenom, 11000000))

	// verify liquidator's new umee balance (10k - 10) = 9990 umee
	tokenBalance := app.BankKeeper.GetBalance(ctx, liquidatorAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9990000000))

	// liquidator fully liquidates lender, receiving more collateral and reducing borrowed amount to zero
	repayment = sdk.NewInt64Coin(umeeapp.BondDenom, 300000000) // 300 umee
	repaid, reward, err = s.app.LeverageKeeper.LiquidateBorrow(ctx, liquidatorAddr, lenderAddr, repayment, unrestrictedReward)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(80000000), repaid)
	s.Require().Equal(sdk.NewInt(88000000), reward)

	// verify that repayment has not been modified
	s.Require().Equal(sdk.NewInt(300000000), repayment.Amount)

	// verify liquidator's new u/umee balance = 99 = (90 + liquidation incentive)
	uTokenBalance = app.BankKeeper.GetBalance(ctx, liquidatorAddr, rewardDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin(rewardDenom, 99000000))

	// verify lender's new loan amount is zero
	loanBalance = s.app.LeverageKeeper.GetBorrow(ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 0))

	// verify liquidator's new umee balance (10k - 90) = 9910 umee
	tokenBalance = app.BankKeeper.GetBalance(ctx, liquidatorAddr, umeeapp.BondDenom)
	s.Require().Equal(tokenBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 9910000000))
}

func (s *IntegrationTestSuite) TestRepayBadDebt() {
	// Creating a lender so module account has some uumee
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
	// and the uToken supply starts at 1000 due to lender account
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
	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral.
	lenderAddr, _ := s.initBorrowScenario()

	// lender borrows 40 umee
	err := s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 40000000))
	s.Require().NoError(err)

	// verify lender's loan amount (40 umee)
	loanBalance := s.app.LeverageKeeper.GetBorrow(s.ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 40000000))

	// Because no time has passed since genesis (due to test setup) this will not
	// increase borrowed amount.
	err = s.app.LeverageKeeper.AccrueAllInterest(s.ctx)
	s.Require().NoError(err)

	// verify lender's loan amount (40 umee)
	loanBalance = s.app.LeverageKeeper.GetBorrow(s.ctx, lenderAddr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 40000000))

	// borrow APY at utilization = 4%
	// when kink utilization = 80%, and base/kink APY are 0.02 and 0.22
	borrowAPY := s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(sdk.MustNewDecFromStr("0.03"), borrowAPY)

	// lend APY when borrow APY is 3%
	// and utilization is 4%, and reservefactor is 20%, and OracleRewardFactor is 1%
	// 0.03 * 0.04 * (1 - 0.21) = 0.000948
	lendAPY := s.app.LeverageKeeper.DeriveLendAPY(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("0.000948"), lendAPY)
}

func (s *IntegrationTestSuite) TestDynamicInterest() {
	// Init scenario is being used because the module account (lending pool)
	// already has 1000 umee.
	lenderAddr, _ := s.initBorrowScenario()

	umeeToken := types.Token{
		BaseDenom:            umeeapp.BondDenom,
		ReserveFactor:        sdk.MustNewDecFromStr("0.20"),
		CollateralWeight:     sdk.MustNewDecFromStr("1.0"), // to allow high utilization
		LiquidationThreshold: sdk.MustNewDecFromStr("1.0"), // to allow high utilization
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.22"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("1.52"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:          "UMEE",
		Exponent:             6,
		EnableLend:           true,
		EnableBorrow:         true,
		Blacklist:            false,
	}
	s.Require().NoError(s.app.LeverageKeeper.SetRegisteredToken(s.ctx, umeeToken))

	// Base interest rate (0% utilization)
	rate := s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.02"))

	// lender borrows 200 umee, utilization 200/1000
	err := s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	s.Require().NoError(err)

	// Between base interest and kink (20% utilization)
	rate = s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.07"))

	// lender borrows 600 more umee, utilization 800/1000
	err = s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 600000000))
	s.Require().NoError(err)

	// Kink interest rate (80% utilization)
	rate = s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.22"))

	// lender borrows 100 more umee, utilization 900/1000
	err = s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

	// Between kink interest and max (90% utilization)
	rate = s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.87"))

	// lender borrows 100 more umee, utilization 1000/1000
	err = s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
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

func (s *IntegrationTestSuite) TestSetCollateralSetting_Valid() {
	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee enabled as collateral.
	lenderAddr, _ := s.initBorrowScenario()

	// lender disables u/umee as collateral
	err := s.app.LeverageKeeper.SetCollateralSetting(s.ctx, lenderAddr, "u/uumee", false)
	s.Require().NoError(err)
	enabled := s.app.LeverageKeeper.GetCollateralSetting(s.ctx, lenderAddr, "u/uumee")
	s.Require().Equal(enabled, false)

	// verify lender's uToken balance is 1000 u/umee
	uTokenBalance := s.app.BankKeeper.GetBalance(s.ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))

	// verify lender's uToken collateral is 0 u/umee
	collateralBalance := s.app.LeverageKeeper.GetCollateralAmount(s.ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(collateralBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 0))

	// lender enables u/umee as collateral
	err = s.app.LeverageKeeper.SetCollateralSetting(s.ctx, lenderAddr, "u/uumee", true)
	s.Require().NoError(err)
	enabled = s.app.LeverageKeeper.GetCollateralSetting(s.ctx, lenderAddr, "u/uumee")
	s.Require().Equal(enabled, true)

	// verify lender's uToken balance is 0 u/umee
	uTokenBalance = s.app.BankKeeper.GetBalance(s.ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(uTokenBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 0))

	// verify lender's uToken collateral is 1000 u/umee
	collateralBalance = s.app.LeverageKeeper.GetCollateralAmount(s.ctx, lenderAddr, "u/"+umeeapp.BondDenom)
	s.Require().Equal(collateralBalance, sdk.NewInt64Coin("u/"+umeeapp.BondDenom, 1000000000))
}

func (s *IntegrationTestSuite) TestSetCollateralSetting_Invalid() {
	// Any user from the starting scenario can be used, since they are only toggling
	// collateral settings.
	lenderAddr, _ := s.initBorrowScenario()

	// lender disables u/abcd as collateral - fails because "u/abcd" is not a recognized uToken
	err := s.app.LeverageKeeper.SetCollateralSetting(s.ctx, lenderAddr, "u/abcd", false)
	s.Require().Error(err)

	// lender disables uumee as collateral - fails because "uumee" is an asset, not a uToken
	err = s.app.LeverageKeeper.SetCollateralSetting(s.ctx, lenderAddr, "uumee", false)
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestGetCollateralSetting_Invalid() {
	// Any user from the starting scenario can be used, since we are only viewing
	// collateral settings.
	lenderAddr, _ := s.initBorrowScenario()

	// Regular assets always return false, because only uTokens can be collateral
	enabled := s.app.LeverageKeeper.GetCollateralSetting(s.ctx, lenderAddr, "uumee")
	s.Require().Equal(enabled, false)

	// Invalid or unrecognized assets always return false
	enabled = s.app.LeverageKeeper.GetCollateralSetting(s.ctx, lenderAddr, "abcd")
	s.Require().Equal(enabled, false)
}

func (s *IntegrationTestSuite) TestGetEligibleLiquidationTargets_OneAddrOneAsset() {
	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee enabled as collateral.
	lenderAddr, _ := s.initBorrowScenario()

	// lender borrows 100 umee (max current allowed) lender amount enabled as collateral * CollateralWeight
	// = 1000 * 0.1
	// = 100
	err := s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

	zeroAddresses, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{}, zeroAddresses)

	// Note: Setting umee collateral weight to 0.05 to allow set the lender eligible to liquidation
	umeeToken := types.Token{
		BaseDenom:            umeeapp.BondDenom,
		ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:     sdk.MustNewDecFromStr("0.05"),
		LiquidationThreshold: sdk.MustNewDecFromStr("0.05"),
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.2"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:          "UMEE",
		Exponent:             6,
		EnableLend:           true,
		EnableBorrow:         true,
		Blacklist:            false,
	}
	s.Require().NoError(s.app.LeverageKeeper.SetRegisteredToken(s.ctx, umeeToken))

	lenderAddress, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{lenderAddr}, lenderAddress)

	// if it tries to borrow any other asset it should return an error
	err = s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(atomIBCDenom, 1))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestGetEligibleLiquidationTargets_OneAddrTwoAsset() {
	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee enabled as collateral.
	lenderAddr, _ := s.initBorrowScenario()

	// lender borrows 100 umee (max current allowed) lender amount enabled as collateral * CollateralWeight
	// = 1000 * 0.1
	// = 100
	err := s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

	zeroAddresses, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{}, zeroAddresses)

	mintAmountAtom := int64(100000000) // 100 atom
	lendAmountAtom := int64(50000000)  // 50 atom

	// mints and send to lender 100 atom and already
	// enable 50 u/atom as collateral.
	s.mintAndLendAtom(lenderAddr, mintAmountAtom, lendAmountAtom)

	// lender borrows 4 atom (max current allowed - 1) lender amount enabled as collateral * CollateralWeight
	// = (50 * 0.1) - 1
	// = 4
	err = s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(atomIBCDenom, 4000000)) // 4 atom
	s.Require().NoError(err)

	// Note: Setting umee collateral weight to 0.05 to allow set the lender eligible to liquidation
	umeeToken := types.Token{
		BaseDenom:            umeeapp.BondDenom,
		ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:     sdk.MustNewDecFromStr("0.05"),
		LiquidationThreshold: sdk.MustNewDecFromStr("0.05"),
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.2"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:          "UMEE",
		Exponent:             6,
		EnableLend:           true,
		EnableBorrow:         true,
		Blacklist:            false,
	}
	s.Require().NoError(s.app.LeverageKeeper.SetRegisteredToken(s.ctx, umeeToken))

	// Note: Setting atom collateral weight to 0.01 to the lender be eligible to liquidation in a second token
	atomIBCToken := types.Token{
		BaseDenom:            atomIBCDenom,
		ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:     sdk.MustNewDecFromStr("0.01"),
		LiquidationThreshold: sdk.MustNewDecFromStr("0.01"),
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.2"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:          "UMEE",
		Exponent:             6,
		EnableLend:           true,
		EnableBorrow:         true,
		Blacklist:            false,
	}
	s.Require().NoError(s.app.LeverageKeeper.SetRegisteredToken(s.ctx, atomIBCToken))

	lenderAddress, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{lenderAddr}, lenderAddress)
}

func (s *IntegrationTestSuite) TestGetEligibleLiquidationTargets_TwoAddr() {
	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee enabled as collateral.
	lenderAddr, anotherLender := s.initBorrowScenario()

	// lender borrows 100 umee (max current allowed) lender amount enabled as collateral * CollateralWeight
	// = 1000 * 0.1
	// = 100
	err := s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

	zeroAddresses, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{}, zeroAddresses)

	mintAmountAtom := int64(100000000) // 100 atom
	lendAmountAtom := int64(50000000)  // 50 atom

	// mints and send to anotherLender 100 atom and already
	// enable 50 u/atom as collateral.
	s.mintAndLendAtom(anotherLender, mintAmountAtom, lendAmountAtom)

	// anotherLender borrows 4 atom (max current allowed - 1) anotherLender amount enabled as collateral * CollateralWeight
	// = (50 * 0.1) - 1
	// = 4
	err = s.app.LeverageKeeper.BorrowAsset(s.ctx, anotherLender, sdk.NewInt64Coin(atomIBCDenom, 4000000)) // 4 atom
	s.Require().NoError(err)

	// Note: Setting umee collateral weight to 0.05 to allow set the lender eligible to liquidation
	umeeToken := types.Token{
		BaseDenom:            umeeapp.BondDenom,
		ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:     sdk.MustNewDecFromStr("0.05"),
		LiquidationThreshold: sdk.MustNewDecFromStr("0.05"),
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.2"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:          "UMEE",
		Exponent:             6,
		EnableLend:           true,
		EnableBorrow:         true,
		Blacklist:            false,
	}
	s.Require().NoError(s.app.LeverageKeeper.SetRegisteredToken(s.ctx, umeeToken))

	// Note: Setting atom collateral weight to 0.01 to the anotherLender also be eligible to liquidation
	atomIBCToken := types.Token{
		BaseDenom:            atomIBCDenom,
		ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:     sdk.MustNewDecFromStr("0.01"),
		LiquidationThreshold: sdk.MustNewDecFromStr("0.01"),
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.2"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:          "UMEE",
		Exponent:             6,
		EnableLend:           true,
		EnableBorrow:         true,
		Blacklist:            false,
	}
	s.Require().NoError(s.app.LeverageKeeper.SetRegisteredToken(s.ctx, atomIBCToken))

	lenderAddress, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{lenderAddr, anotherLender}, lenderAddress)
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
	lenderAddr, _ := s.initBorrowScenario()

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// check invariant
	_, broken := keeper.CollateralAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)

	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)

	// withdraw the lended umee in the initBorrowScenario
	err := s.app.LeverageKeeper.WithdrawAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(uTokenDenom, 1000000000))
	s.Require().NoError(err)

	// check invariant
	_, broken = keeper.CollateralAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)
}

func (s *IntegrationTestSuite) TestBorrowAmountInvariant() {
	lenderAddr, _ := s.initBorrowScenario()

	// The "lender" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// lender borrows 20 umee
	err := s.app.LeverageKeeper.BorrowAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000))
	s.Require().NoError(err)

	// check invariant
	_, broken := keeper.BorrowAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)

	// lender repays 30 umee, actually only 20 because is the min between
	// the amount borrowed and the amount repaid
	_, err = s.app.LeverageKeeper.RepayAsset(s.ctx, lenderAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 30000000))
	s.Require().NoError(err)

	// check invariant
	_, broken = keeper.BorrowAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)
}

func (s *IntegrationTestSuite) TestWithdrawAsset_InsufficientCollateral() {
	// Create a lender with 1 u/umee collateral by lending 1 umee
	lenderAddr := s.setupAccount(umeeapp.BondDenom, 1000000, 1000000, 0, true)

	// Create an additional lender so lending pool has extra umee
	_ = s.setupAccount(umeeapp.BondDenom, 1000000, 1000000, 0, true)

	// verify collateral amount and total supply of minted uTokens
	uTokenDenom := types.UTokenFromTokenDenom(umeeapp.BondDenom)
	collateral := s.app.LeverageKeeper.GetCollateralAmount(s.ctx, lenderAddr, uTokenDenom)
	s.Require().Equal(sdk.NewInt64Coin(uTokenDenom, 1000000), collateral) // 1 u/umee
	supply := s.app.LeverageKeeper.GetUTokenSupply(s.ctx, uTokenDenom)
	s.Require().Equal(sdk.NewInt64Coin(uTokenDenom, 2000000), supply) // 2 u/umee

	// withdraw more collateral than available
	uToken := collateral.Add(sdk.NewInt64Coin(uTokenDenom, 1))
	err := s.app.LeverageKeeper.WithdrawAsset(s.ctx, lenderAddr, uToken)
	s.Require().EqualError(err, "1000001u/uumee: insufficient balance")
}
