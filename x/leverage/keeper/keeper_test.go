package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/x/leverage/keeper"
	"github.com/umee-network/umee/v2/x/leverage/types"
)

func (s *IntegrationTestSuite) TestSupply() {
	type testCase struct {
		msg             string
		addr            sdk.AccAddress
		coin            sdk.Coin
		expectedUTokens sdk.Coin
		err             error
	}

	app, ctx, require := s.app, s.ctx, s.Require()

	// create and fund a supplier with 100 UMEE and 100 ATOM
	supplier := s.newAccount(coin(umeeDenom, 100_000000), coin(atomDenom, 100_000000))

	// create and modify a borrower to force the uToken exchange rate of ATOM from 1 to 1.5
	borrower := s.newAccount(coin(atomDenom, 100_000000))
	s.supply(borrower, coin(atomDenom, 100_000000))
	s.collateralize(borrower, coin("u/"+atomDenom, 100_000000))
	s.borrow(borrower, coin(atomDenom, 10_000000))
	s.tk.SetBorrow(ctx, borrower, coin(atomDenom, 60_000000))

	tcs := []testCase{
		{
			"unregistered denom",
			supplier,
			coin("abcd", 80_000000),
			sdk.Coin{},
			types.ErrNotRegisteredToken,
		},
		{
			"uToken",
			supplier,
			coin("u/"+umeeDenom, 80_000000),
			sdk.Coin{},
			types.ErrUToken,
		},
		{
			"insufficient balance",
			supplier,
			coin(umeeDenom, 120_000000),
			sdk.Coin{},
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"valid supply",
			supplier,
			coin(umeeDenom, 80_000000),
			coin("u/"+umeeDenom, 80_000000),
			nil,
		},
		{
			"high exchange rate",
			supplier,
			coin(atomDenom, 60_000000),
			coin("u/"+atomDenom, 40_000000),
			nil,
		},
	}

	for _, tc := range tcs {
		if tc.err != nil {
			_, err := app.LeverageKeeper.Supply(ctx, tc.addr, tc.coin)
			require.ErrorIs(err, tc.err, tc.msg)
		} else {
			denom := tc.coin.Denom
			uDenom := types.ToUTokenDenom(tc.coin.Denom)

			// initial state
			iBalance := app.BankKeeper.GetBalance(ctx, tc.addr, denom)
			iUTokens := app.BankKeeper.GetBalance(ctx, tc.addr, uDenom)
			iCollateral := app.LeverageKeeper.GetCollateralAmount(ctx, tc.addr, uDenom)
			iUTokenSupply := app.LeverageKeeper.GetUTokenSupply(ctx, uDenom)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)

			// verify the outputs of supply function
			uToken, err := app.LeverageKeeper.Supply(ctx, tc.addr, tc.coin)
			require.NoError(err)
			require.Equal(tc.expectedUTokens, uToken, tc.msg)

			// final state
			fBalance := app.BankKeeper.GetBalance(ctx, tc.addr, denom)
			fUTokens := app.BankKeeper.GetBalance(ctx, tc.addr, uDenom)
			fCollateral := app.LeverageKeeper.GetCollateralAmount(ctx, tc.addr, uDenom)
			fUTokenSupply := app.LeverageKeeper.GetUTokenSupply(ctx, uDenom)
			fExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)

			// verify token balance decreased by the expected amount
			require.Equal(iBalance.Sub(tc.coin), fBalance, tc.msg, "token balance")
			// verify uToken balance increased by the expected amount
			require.Equal(iUTokens.Add(tc.expectedUTokens), fUTokens, tc.msg, "uToken balance")
			// verify uToken collateral unchanged
			require.Equal(iCollateral.Amount.Int64(), fCollateral.Amount.Int64(), tc.msg, "uToken collateral")
			// verify uToken supply increased by the expected amount
			require.Equal(iUTokenSupply.Add(tc.expectedUTokens), fUTokenSupply, tc.msg, "uToken supply")
			// verify uToken exchange rate is unchanged
			require.Equal(iExchangeRate, fExchangeRate, tc.msg, "uToken exchange rate")
		}
	}
}

func (s *IntegrationTestSuite) TestWithdraw() {
	type testCase struct {
		msg                  string
		addr                 sdk.AccAddress
		uToken               sdk.Coin
		expectFromBalance    int64
		expectFromCollateral int64
		expectedTokens       sdk.Coin
		err                  error
	}

	app, ctx, require := s.app, s.ctx, s.Require()

	// create and fund a supplier with 100 UMEE and 100 ATOM, then supply 100 UMEE and 50 ATOM
	// also collateralize 75 of supplied UMEE
	supplier := s.newAccount(coin(umeeDenom, 100_000000), coin(atomDenom, 100_000000))
	s.supply(supplier, coin(umeeDenom, 100_000000))
	s.collateralize(supplier, coin("u/"+umeeDenom, 75_000000))
	s.supply(supplier, coin(atomDenom, 50_000000))

	// create and modify a borrower to force the uToken exchange rate of ATOM from 1 to 1.2
	borrower := s.newAccount(coin(atomDenom, 100_000000))
	s.supply(borrower, coin(atomDenom, 100_000000))
	s.collateralize(borrower, coin("u/"+atomDenom, 100_000000))
	s.borrow(borrower, coin(atomDenom, 10_000000))
	s.tk.SetBorrow(ctx, borrower, coin(atomDenom, 40_000000))

	// create an additional UMEE supplier
	other := s.newAccount(coin(umeeDenom, 100_000000))
	s.supply(other, coin(umeeDenom, 100_000000))

	tcs := []testCase{
		{
			"unregistered base token",
			supplier,
			coin("abcd", 80_000000),
			0,
			0,
			sdk.Coin{},
			types.ErrNotUToken,
		},
		{
			"base token",
			supplier,
			coin(umeeDenom, 80_000000),
			0,
			0,
			sdk.Coin{},
			types.ErrNotUToken,
		},
		{
			"insufficient uTokens",
			supplier,
			coin("u/"+umeeDenom, 120_000000),
			0,
			0,
			sdk.Coin{},
			types.ErrInsufficientBalance,
		},
		{
			"withdraw from balance",
			supplier,
			coin("u/"+umeeDenom, 10_000000),
			10_000000,
			0,
			coin(umeeDenom, 10_000000),
			nil,
		},
		{
			"withdraw from collateral",
			supplier,
			coin("u/"+umeeDenom, 90_000000),
			15_000000,
			75_000000,
			coin(umeeDenom, 90_000000),
			nil,
		},
		{
			"high exchange rate",
			supplier,
			coin("u/"+atomDenom, 50_000000),
			50_000000,
			0,
			coin(atomDenom, 60_000000),
			nil,
		},
		{
			"borrow limit",
			borrower,
			coin("u/"+atomDenom, 50_000000),
			0,
			0,
			sdk.Coin{},
			types.ErrUndercollaterized,
		},
	}

	for _, tc := range tcs {
		if tc.err != nil {
			_, err := app.LeverageKeeper.Withdraw(ctx, tc.addr, tc.uToken)
			require.ErrorIs(err, tc.err, tc.msg)
		} else {
			denom := types.ToTokenDenom(tc.uToken.Denom)
			uDenom := tc.uToken.Denom

			// initial state
			iBalance := app.BankKeeper.GetBalance(ctx, tc.addr, denom)
			iUTokens := app.BankKeeper.GetBalance(ctx, tc.addr, uDenom)
			iCollateral := app.LeverageKeeper.GetCollateralAmount(ctx, tc.addr, uDenom)
			iUTokenSupply := app.LeverageKeeper.GetUTokenSupply(ctx, uDenom)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)

			// verify the outputs of withdraw function
			token, err := app.LeverageKeeper.Withdraw(ctx, tc.addr, tc.uToken)

			require.NoError(err)
			require.Equal(tc.expectedTokens, token, tc.msg)

			// final state
			fBalance := app.BankKeeper.GetBalance(ctx, tc.addr, denom)
			fUTokens := app.BankKeeper.GetBalance(ctx, tc.addr, uDenom)
			fCollateral := app.LeverageKeeper.GetCollateralAmount(ctx, tc.addr, uDenom)
			fUTokenSupply := app.LeverageKeeper.GetUTokenSupply(ctx, uDenom)
			fExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)

			// verify token balance increased by the expected amount
			require.Equal(iBalance.Add(tc.expectedTokens), fBalance, tc.msg, "token balance")
			// verify uToken balance decreased by the expected amount
			require.Equal(iUTokens.Amount.Int64()-tc.expectFromBalance,
				fUTokens.Amount.Int64(), tc.msg, "uTokens from balance")
			// verify uToken collateral decreased by the expected amount
			require.Equal(iCollateral.Amount.Int64()-tc.expectFromCollateral,
				fCollateral.Amount.Int64(), tc.msg, "uToken collateral")
			// verify uToken supply decreased by the expected amount
			require.Equal(iUTokenSupply.Sub(tc.uToken).Amount.Int64(),
				fUTokenSupply.Amount.Int64(), tc.msg, "uToken supply")
			// verify uToken exchange rate is unchanged
			require.Equal(iExchangeRate, fExchangeRate, tc.msg, "uToken exchange rate")
		}
	}
}

func (s *IntegrationTestSuite) TestCollateralize() {
	type testCase struct {
		msg    string
		addr   sdk.AccAddress
		uToken sdk.Coin
		err    error
	}

	app, ctx, require := s.app, s.ctx, s.Require()

	// create and fund a supplier with 200 UMEE, then supply 100 UMEE
	supplier := s.newAccount(coin(umeeDenom, 200_000000))
	s.supply(supplier, coin(umeeDenom, 100_000000))

	tcs := []testCase{
		{
			"base token",
			supplier,
			coin(umeeDenom, 80_000000),
			types.ErrNotUToken,
		},
		{
			"unregistered uToken",
			supplier,
			coin("u/abcd", 80_000000),
			types.ErrNotRegisteredToken,
		},
		{
			"valid collateralize",
			supplier,
			coin("u/"+umeeDenom, 80_000000),
			nil,
		},
		{
			"insufficient balance",
			supplier,
			coin("u/"+umeeDenom, 40_000000),
			sdkerrors.ErrInsufficientFunds,
		},
	}

	for _, tc := range tcs {
		if tc.err != nil {
			err := app.LeverageKeeper.Collateralize(ctx, tc.addr, tc.uToken)
			require.ErrorIs(err, tc.err, tc.msg)
		} else {
			denom := types.ToTokenDenom(tc.uToken.Denom)
			uDenom := tc.uToken.Denom

			// initial state
			iBalance := app.BankKeeper.GetBalance(ctx, tc.addr, denom)
			iUTokens := app.BankKeeper.GetBalance(ctx, tc.addr, uDenom)
			iCollateral := app.LeverageKeeper.GetCollateralAmount(ctx, tc.addr, uDenom)
			iUTokenSupply := app.LeverageKeeper.GetUTokenSupply(ctx, uDenom)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)

			// verify the output of collateralize function
			err := app.LeverageKeeper.Collateralize(ctx, tc.addr, tc.uToken)
			require.NoError(err)

			// final state
			fBalance := app.BankKeeper.GetBalance(ctx, tc.addr, denom)
			fUTokens := app.BankKeeper.GetBalance(ctx, tc.addr, uDenom)
			fCollateral := app.LeverageKeeper.GetCollateralAmount(ctx, tc.addr, uDenom)
			fUTokenSupply := app.LeverageKeeper.GetUTokenSupply(ctx, uDenom)
			fExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)

			// verify token balance is unchanged
			require.Equal(iBalance, fBalance, tc.msg, "token balance")
			// verify uToken balance decreased by the expected amount
			require.Equal(iUTokens.Sub(tc.uToken), fUTokens, tc.msg, "uToken balance")
			// verify uToken collateral increased by the expected amount
			require.Equal(iCollateral.Add(tc.uToken), fCollateral, tc.msg, "uToken collateral")
			// verify uToken supply is unchanged
			require.Equal(iUTokenSupply, fUTokenSupply, tc.msg, "uToken supply")
			// verify uToken exchange rate is unchanged
			require.Equal(iExchangeRate, fExchangeRate, tc.msg, "uToken exchange rate")
		}
	}
}

func (s *IntegrationTestSuite) TestDecollateralize() {
	type testCase struct {
		msg    string
		addr   sdk.AccAddress
		uToken sdk.Coin
		err    error
	}

	app, ctx, require := s.app, s.ctx, s.Require()

	// create and fund a supplier with 200 UMEE, then supply and collateralize 100 UMEE
	supplier := s.newAccount(coin(umeeDenom, 200_000000))
	s.supply(supplier, coin(umeeDenom, 100_000000))
	s.collateralize(supplier, coin("u/"+umeeDenom, 100_000000))

	// create a borrower which supplies, collateralizes, then borrows ATOM
	borrower := s.newAccount(coin(atomDenom, 100_000000))
	s.supply(borrower, coin(atomDenom, 100_000000))
	s.collateralize(borrower, coin("u/"+atomDenom, 100_000000))
	s.borrow(borrower, coin(atomDenom, 10_000000))

	tcs := []testCase{
		{
			"base token",
			supplier,
			coin(umeeDenom, 80_000000),
			types.ErrNotUToken,
		},
		{
			"valid decollateralize",
			supplier,
			coin("u/"+umeeDenom, 80_000000),
			nil,
		},
		{
			"insufficient collateral",
			supplier,
			coin("u/"+umeeDenom, 40_000000),
			types.ErrInsufficientCollateral,
		},
		{
			"borrow limit",
			borrower,
			coin("u/"+atomDenom, 100_000000),
			types.ErrUndercollaterized,
		},
	}

	for _, tc := range tcs {
		if tc.err != nil {
			err := app.LeverageKeeper.Decollateralize(ctx, tc.addr, tc.uToken)
			require.ErrorIs(err, tc.err, tc.msg)
		} else {
			denom := types.ToTokenDenom(tc.uToken.Denom)
			uDenom := tc.uToken.Denom

			// initial state
			iBalance := app.BankKeeper.GetBalance(ctx, tc.addr, denom)
			iUTokens := app.BankKeeper.GetBalance(ctx, tc.addr, uDenom)
			iCollateral := app.LeverageKeeper.GetCollateralAmount(ctx, tc.addr, uDenom)
			iUTokenSupply := app.LeverageKeeper.GetUTokenSupply(ctx, uDenom)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)

			// verify the output of decollateralize function
			err := app.LeverageKeeper.Decollateralize(ctx, tc.addr, tc.uToken)
			require.NoError(err)

			// final state
			fBalance := app.BankKeeper.GetBalance(ctx, tc.addr, denom)
			fUTokens := app.BankKeeper.GetBalance(ctx, tc.addr, uDenom)
			fCollateral := app.LeverageKeeper.GetCollateralAmount(ctx, tc.addr, uDenom)
			fUTokenSupply := app.LeverageKeeper.GetUTokenSupply(ctx, uDenom)
			fExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)

			// verify token balance is unchanged
			require.Equal(iBalance, fBalance, tc.msg, "token balance")
			// verify uToken balance increased by the expected amount
			require.Equal(iUTokens.Add(tc.uToken), fUTokens, tc.msg, "uToken balance")
			// verify uToken collateral decreased by the expected amount
			require.Equal(iCollateral.Sub(tc.uToken), fCollateral, tc.msg, "uToken collateral")
			// verify uToken supply is unchanged
			require.Equal(iUTokenSupply, fUTokenSupply, tc.msg, "uToken supply")
			// verify uToken exchange rate is unchanged
			require.Equal(iExchangeRate, fExchangeRate, tc.msg, "uToken exchange rate")
		}
	}
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
	uDenom := types.ToUTokenDenom(umeeapp.BondDenom)
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
	_, err = s.app.LeverageKeeper.Withdraw(s.ctx, addr, sdk.NewCoin(uDenom, collateral.Amount))
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
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 8000000), repaid)

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
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 12000000), repaid)

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
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 20000000), repaid)

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
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(atomDenom, 1))
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

	// mints and send to addr 100 atom and already
	// enable 50 u/atom as collateral.
	s.fundAccount(addr, coin(atomDenom, 100_000000))
	s.supply(addr, coin(atomDenom, 50_000000))
	s.collateralize(addr, coin("u/"+atomDenom, 50_000000))

	// user borrows 4 atom (max current allowed - 1) user amount enabled as collateral * CollateralWeight
	// = (50 * 0.1) - 1
	// = 4
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(atomDenom, 4000000)) // 4 atom
	s.Require().NoError(err)

	// Note: Setting umee liquidation threshold to 0.05 to make the user eligible for liquidation
	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("0.05")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.05")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// Note: Setting atom collateral weight to 0.01 to make the user eligible for liquidation
	atomIBCToken := newToken(atomDenom, "ATOM")
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

	// mints and send to anotherSupplier 100 atom and already
	// enable 50 u/atom as collateral.
	s.fundAccount(anotherSupplier, coin(atomDenom, 100_000000))
	s.supply(anotherSupplier, coin(atomDenom, 50_000000))
	s.collateralize(anotherSupplier, coin("u/"+atomDenom, 50_000000))

	// anotherSupplier borrows 4 atom (max current allowed - 1) anotherSupplier amount enabled as collateral * CollateralWeight
	// = (50 * 0.1) - 1
	// = 4
	err = s.app.LeverageKeeper.Borrow(s.ctx, anotherSupplier, sdk.NewInt64Coin(atomDenom, 4000000)) // 4 atom
	s.Require().NoError(err)

	// Note: Setting umee liquidation threshold to 0.05 to make the supplier eligible for liquidation
	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("0.05")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.05")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// Note: Setting atom collateral weight to 0.01 to make the supplier eligible for liquidation
	atomIBCToken := newToken(atomDenom, "ATOM")
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

	uTokenDenom := types.ToUTokenDenom(umeeapp.BondDenom)

	// withdraw the supplyed umee in the initBorrowScenario
	_, err := s.app.LeverageKeeper.Withdraw(s.ctx, addr, sdk.NewInt64Coin(uTokenDenom, 1000000000))
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

func (s *IntegrationTestSuite) TestTotalCollateral() {
	// Test zero collateral
	uDenom := types.ToUTokenDenom(umeeDenom)
	collateral := s.app.LeverageKeeper.GetTotalCollateral(s.ctx, uDenom)
	s.Require().Equal(sdk.ZeroInt(), collateral)

	// Uses borrow scenario, because supplier possesses collateral
	_, _ = s.initBorrowScenario()

	// Test nonzero collateral
	collateral = s.app.LeverageKeeper.GetTotalCollateral(s.ctx, uDenom)
	s.Require().Equal(sdk.NewInt(1000000000), collateral)
}

func (s *IntegrationTestSuite) TestLiqudate() {
	addr, _ := s.initBorrowScenario()

	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral.

	// user borrows 90 umee
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 90000000))
	s.Require().NoError(err)

	// create an account and address which will represent a liquidator
	liquidatorAddr := sdk.AccAddress([]byte("addr______________03"))
	liquidatorAcc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, liquidatorAddr)
	s.app.AccountKeeper.SetAccount(s.ctx, liquidatorAcc)

	// mint and send 10k umee to liquidator
	s.Require().NoError(s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName,
		sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 10000000000)), // 10k umee
	))
	s.Require().NoError(s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, liquidatorAddr,
		sdk.NewCoins(sdk.NewInt64Coin(umeeapp.BondDenom, 10000000000)), // 10k umee,
	))

	// liquidator attempts to liquidate user, but user is ineligible (not over borrow limit)
	repayment := sdk.NewInt64Coin(umeeapp.BondDenom, 30000000) // 30 umee
	rewardDenom := types.ToUTokenDenom(umeeapp.BondDenom)
	_, _, _, err = s.app.LeverageKeeper.Liquidate(s.ctx, liquidatorAddr, addr, repayment, rewardDenom)
	s.Require().Error(err)

	// set umee liquidation threshold to 0.01 to allow liquidation
	umeeToken, err := s.app.LeverageKeeper.GetTokenSettings(s.ctx, umeeDenom)
	s.Require().NoError(err)
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("0.01")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.01")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// liquidator partially liquidates user, receiving some uTokens
	repayment = sdk.NewInt64Coin(umeeapp.BondDenom, 10000000) // 10 umee
	repaid, liquidated, reward, err := s.app.LeverageKeeper.Liquidate(
		s.ctx, liquidatorAddr, addr, repayment, types.ToUTokenDenom(umeeDenom),
	)
	s.Require().NoError(err)
	s.Require().Equal(repayment, repaid)                                                      // 10 umee
	s.Require().Equal(sdk.NewInt64Coin(types.ToUTokenDenom(umeeDenom), 11000000), liquidated) // 11 u/umee
	s.Require().Equal(sdk.NewInt64Coin(types.ToUTokenDenom(umeeDenom), 11000000), reward)     // 11 u/umee

	// verify borrower's new borrowed amount is 80 umee (still over borrow limit)
	borrowed := s.app.LeverageKeeper.GetBorrow(s.ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeapp.BondDenom, 80000000), borrowed)

	// verify borrower's new collateral amount (1k - 11) = 989 u/umee
	collateral := s.app.LeverageKeeper.GetCollateralAmount(s.ctx, addr, types.ToUTokenDenom(umeeDenom))
	s.Require().Equal(sdk.NewInt64Coin(types.ToUTokenDenom(umeeDenom), 989000000), collateral)

	// verify liquidator's new u/umee balance = 11 = (10 + liquidation incentive)
	uTokenBalance := s.app.BankKeeper.GetBalance(s.ctx, liquidatorAddr, rewardDenom)
	s.Require().Equal(sdk.NewInt64Coin(rewardDenom, 11000000), uTokenBalance)

	// verify liquidator's new umee balance (10k - 11) = 9990 umee
	tokenBalance := s.app.BankKeeper.GetBalance(s.ctx, liquidatorAddr, umeeapp.BondDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeapp.BondDenom, 9990000000), tokenBalance)

	// liquidator partially liquidates user, receiving base tokens directly at slightly reduced incentive
	repaid, liquidated, reward, err = s.app.LeverageKeeper.Liquidate(
		s.ctx, liquidatorAddr, addr, repayment, umeeDenom,
	)
	s.Require().NoError(err)
	s.Require().Equal(repayment, repaid)                                                      // 10 umee
	s.Require().Equal(sdk.NewInt64Coin(types.ToUTokenDenom(umeeDenom), 10900000), liquidated) // 10.9 u/umee
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 10900000), reward)                          // 10.9 umee

	// verify borrower's new borrow amount is 70 umee (still over borrow limit)
	borrowed = s.app.LeverageKeeper.GetBorrow(s.ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeapp.BondDenom, 70000000), borrowed)

	// verify borrower's new collateral amount (989 - 10.9) = 978.1 u/umee
	collateral = s.app.LeverageKeeper.GetCollateralAmount(s.ctx, addr, types.ToUTokenDenom(umeeDenom))
	s.Require().Equal(sdk.NewInt64Coin(types.ToUTokenDenom(umeeDenom), 978100000), collateral)

	// verify liquidator's u/umee balance = 11 (unchanged)
	uTokenBalance = s.app.BankKeeper.GetBalance(s.ctx, liquidatorAddr, rewardDenom)
	s.Require().Equal(sdk.NewInt64Coin(rewardDenom, 11000000), uTokenBalance)

	// verify liquidator's new umee balance (9990 - 10 + 10.9) = 9990.9 umee
	tokenBalance = s.app.BankKeeper.GetBalance(s.ctx, liquidatorAddr, umeeapp.BondDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeapp.BondDenom, 9990900000), tokenBalance)

	// liquidator fully liquidates user, receiving more collateral and reducing borrowed amount to zero
	repayment = sdk.NewInt64Coin(umeeapp.BondDenom, 300000000) // 300 umee
	repaid, liquidated, reward, err = s.app.LeverageKeeper.Liquidate(
		s.ctx, liquidatorAddr, addr, repayment, types.ToUTokenDenom(umeeDenom),
	)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 70000000), repaid)                          // 70 umee
	s.Require().Equal(sdk.NewInt64Coin(types.ToUTokenDenom(umeeDenom), 77000000), liquidated) // 77 u/umee
	s.Require().Equal(sdk.NewInt64Coin(types.ToUTokenDenom(umeeDenom), 77000000), reward)     // 77 u/umee

	// verify that repayment has not been modified
	s.Require().Equal(sdk.NewInt(300000000), repayment.Amount)

	// verify liquidator's new u/umee balance = 88 = (11 + 77)
	uTokenBalance = s.app.BankKeeper.GetBalance(s.ctx, liquidatorAddr, rewardDenom)
	s.Require().Equal(sdk.NewInt64Coin(rewardDenom, 88000000), uTokenBalance)

	// verify borrower's new borrowed amount is zero
	borrowed = s.app.LeverageKeeper.GetBorrow(s.ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeapp.BondDenom, 0), borrowed)

	// verify borrower's new collateral amount (978.1 - 77) = 901.1 u/umee
	collateral = s.app.LeverageKeeper.GetCollateralAmount(s.ctx, addr, types.ToUTokenDenom(umeeDenom))
	s.Require().Equal(sdk.NewInt64Coin(types.ToUTokenDenom(umeeDenom), 901100000), collateral)

	// verify liquidator's new umee balance (9990.9 - 70) = 9920.9 umee
	tokenBalance = s.app.BankKeeper.GetBalance(s.ctx, liquidatorAddr, umeeapp.BondDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeapp.BondDenom, 9920900000), tokenBalance)
}
