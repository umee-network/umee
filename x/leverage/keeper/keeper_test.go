package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v3/x/leverage/types"
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

	// create a supplier that will exceed token's default MaxSupply
	whale := s.newAccount(coin(umeeDenom, 1_000_000_000000))

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
			"no balance",
			borrower,
			coin(umeeDenom, 20_000000),
			sdk.Coin{},
			sdkerrors.ErrInsufficientFunds,
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
			"additional supply",
			supplier,
			coin(umeeDenom, 20_000000),
			coin("u/"+umeeDenom, 20_000000),
			nil,
		},
		{
			"high exchange rate",
			supplier,
			coin(atomDenom, 60_000000),
			coin("u/"+atomDenom, 40_000000),
			nil,
		},
		{
			"max supply",
			whale,
			coin(umeeDenom, 1_000_000_000000),
			sdk.Coin{},
			types.ErrMaxSupply,
		},
	}

	for _, tc := range tcs {
		if tc.err != nil {
			_, err := app.LeverageKeeper.Supply(ctx, tc.addr, tc.coin)
			require.ErrorIs(err, tc.err, tc.msg)
		} else {
			denom := tc.coin.Denom

			// initial state
			iBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)
			iBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify the outputs of supply function
			uToken, err := app.LeverageKeeper.Supply(ctx, tc.addr, tc.coin)
			require.NoError(err, tc.msg)
			require.Equal(tc.expectedUTokens, uToken, tc.msg)

			// final state
			fBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)
			fBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify token balance decreased and uToken balance increased by the expected amounts
			require.Equal(iBalance.Sub(tc.coin).Add(tc.expectedUTokens), fBalance, tc.msg, "token balance")
			// verify uToken collateral unchanged
			require.Equal(iCollateral, fCollateral, tc.msg, "uToken collateral")
			// verify uToken supply increased by the expected amount
			require.Equal(iUTokenSupply.Add(tc.expectedUTokens), fUTokenSupply, tc.msg, "uToken supply")
			// verify uToken exchange rate is unchanged
			require.Equal(iExchangeRate, fExchangeRate, tc.msg, "uToken exchange rate")
			// verify borrowed coins are unchanged
			require.Equal(iBorrowed, fBorrowed, tc.msg, "borrowed coins")

			// check all available invariants
			s.checkInvariants(tc.msg)
		}
	}
}

func (s *IntegrationTestSuite) TestWithdraw() {
	type testCase struct {
		msg                  string
		addr                 sdk.AccAddress
		uToken               sdk.Coin
		expectFromBalance    sdk.Coins
		expectFromCollateral sdk.Coins
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
			nil,
			nil,
			sdk.Coin{},
			types.ErrNotUToken,
		},
		{
			"base token",
			supplier,
			coin(umeeDenom, 80_000000),
			nil,
			nil,
			sdk.Coin{},
			types.ErrNotUToken,
		},
		{
			"insufficient uTokens",
			supplier,
			coin("u/"+umeeDenom, 120_000000),
			nil,
			nil,
			sdk.Coin{},
			types.ErrInsufficientBalance,
		},
		{
			"withdraw from balance",
			supplier,
			coin("u/"+umeeDenom, 10_000000),
			sdk.NewCoins(coin("u/"+umeeDenom, 10_000000)),
			nil,
			coin(umeeDenom, 10_000000),
			nil,
		},
		{
			"some from collateral",
			supplier,
			coin("u/"+umeeDenom, 80_000000),
			sdk.NewCoins(coin("u/"+umeeDenom, 15_000000)),
			sdk.NewCoins(coin("u/"+umeeDenom, 65_000000)),
			coin(umeeDenom, 80_000000),
			nil,
		},
		{
			"only from collateral",
			supplier,
			coin("u/"+umeeDenom, 10_000000),
			nil,
			sdk.NewCoins(coin("u/"+umeeDenom, 10_000000)),
			coin(umeeDenom, 10_000000),
			nil,
		},
		{
			"high exchange rate",
			supplier,
			coin("u/"+atomDenom, 50_000000),
			sdk.NewCoins(coin("u/"+atomDenom, 50_000000)),
			nil,
			coin(atomDenom, 60_000000),
			nil,
		},
		{
			"borrow limit",
			borrower,
			coin("u/"+atomDenom, 50_000000),
			nil,
			nil,
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

			// initial state
			iBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)
			iBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify the outputs of withdraw function
			token, err := app.LeverageKeeper.Withdraw(ctx, tc.addr, tc.uToken)

			require.NoError(err, tc.msg)
			require.Equal(tc.expectedTokens, token, tc.msg)

			// final state
			fBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)
			fBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify token balance increased by the expected amount
			require.Equal(iBalance.Add(tc.expectedTokens).Sub(tc.expectFromBalance...),
				fBalance, tc.msg, "token balance")
			// verify uToken collateral decreased by the expected amount
			s.requireEqualCoins(iCollateral.Sub(tc.expectFromCollateral...), fCollateral, tc.msg, "uToken collateral")
			// verify uToken supply decreased by the expected amount
			require.Equal(iUTokenSupply.Sub(tc.uToken), fUTokenSupply, tc.msg, "uToken supply")
			// verify uToken exchange rate is unchanged
			require.Equal(iExchangeRate, fExchangeRate, tc.msg, "uToken exchange rate")
			// verify borrowed coins are unchanged
			require.Equal(iBorrowed, fBorrowed, tc.msg, "borrowed coins")

			// check all available invariants
			s.checkInvariants(tc.msg)
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

	// create and fund another supplier
	otherSupplier := s.newAccount(coin(umeeDenom, 200_000000), coin(atomDenom, 200_000000))
	s.supply(otherSupplier, coin(umeeDenom, 200_000000), coin(atomDenom, 200_000000))

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
			"wrong balance",
			supplier,
			coin("u/"+atomDenom, 10_000000),
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"valid collateralize",
			supplier,
			coin("u/"+umeeDenom, 80_000000),
			nil,
		},
		{
			"additional collateralize",
			supplier,
			coin("u/"+umeeDenom, 10_000000),
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

			// initial state
			iBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)
			iBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify the output of collateralize function
			err := app.LeverageKeeper.Collateralize(ctx, tc.addr, tc.uToken)
			require.NoError(err, tc.msg)

			// final state
			fBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)
			fBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify uToken balance decreased by the expected amount
			require.Equal(iBalance.Sub(tc.uToken), fBalance, tc.msg, "uToken balance")
			// verify uToken collateral increased by the expected amount
			require.Equal(iCollateral.Add(tc.uToken), fCollateral, tc.msg, "uToken collateral")
			// verify uToken supply is unchanged
			require.Equal(iUTokenSupply, fUTokenSupply, tc.msg, "uToken supply")
			// verify uToken exchange rate is unchanged
			require.Equal(iExchangeRate, fExchangeRate, tc.msg, "uToken exchange rate")
			// verify borrowed coins are unchanged
			require.Equal(iBorrowed, fBorrowed, tc.msg, "borrowed coins")

			// check all available invariants
			s.checkInvariants(tc.msg)
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
			"no collateral",
			supplier,
			coin("u/"+atomDenom, 40_000000),
			types.ErrInsufficientCollateral,
		},
		{
			"valid decollateralize",
			supplier,
			coin("u/"+umeeDenom, 80_000000),
			nil,
		},
		{
			"additional decollateralize",
			supplier,
			coin("u/"+umeeDenom, 10_000000),
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

			// initial state
			iBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)
			iBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify the output of decollateralize function
			err := app.LeverageKeeper.Decollateralize(ctx, tc.addr, tc.uToken)
			require.NoError(err, tc.msg)

			// final state
			fBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)
			fBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify uToken balance increased by the expected amount
			require.Equal(iBalance.Add(tc.uToken), fBalance, tc.msg, "uToken balance")
			// verify uToken collateral decreased by the expected amount
			require.Equal(iCollateral.Sub(tc.uToken), fCollateral, tc.msg, "uToken collateral")
			// verify uToken supply is unchanged
			require.Equal(iUTokenSupply, fUTokenSupply, tc.msg, "uToken supply")
			// verify uToken exchange rate is unchanged
			require.Equal(iExchangeRate, fExchangeRate, tc.msg, "uToken exchange rate")
			// verify borrowed coins are unchanged
			require.Equal(iBorrowed, fBorrowed, tc.msg, "borrowed coins")

			// check all available invariants
			s.checkInvariants(tc.msg)
		}
	}
}

func (s *IntegrationTestSuite) TestBorrow() {
	type testCase struct {
		msg  string
		addr sdk.AccAddress
		coin sdk.Coin
		err  error
	}

	app, ctx, require := s.app, s.ctx, s.Require()

	// create and fund a supplier which supplies UMEE and ATOM
	supplier := s.newAccount(coin(umeeDenom, 100_000000), coin(atomDenom, 100_000000))
	s.supply(supplier, coin(umeeDenom, 100_000000), coin(atomDenom, 100_000000))

	// create a borrower which supplies and collateralizes 100 ATOM
	borrower := s.newAccount(coin(atomDenom, 100_000000))
	s.supply(borrower, coin(atomDenom, 100_000000))
	s.collateralize(borrower, coin("u/"+atomDenom, 100_000000))

	tcs := []testCase{
		{
			"uToken",
			borrower,
			coin("u/"+umeeDenom, 100_000000),
			types.ErrUToken,
		},
		{
			"unregistered token",
			borrower,
			coin("abcd", 100_000000),
			types.ErrNotRegisteredToken,
		},
		{
			"lending pool insufficient",
			borrower,
			coin(umeeDenom, 200_000000),
			types.ErrLendingPoolInsufficient,
		},
		{
			"valid borrow",
			borrower,
			coin(umeeDenom, 70_000000),
			nil,
		},
		{
			"additional borrow",
			borrower,
			coin(umeeDenom, 20_000000),
			nil,
		},
		{
			"max supply utilization",
			borrower,
			coin(umeeDenom, 10_000000),
			types.ErrMaxSupplyUtilization,
		},
		{
			"atom borrow",
			borrower,
			coin(atomDenom, 1_000000),
			nil,
		},
		{
			"borrow limit",
			borrower,
			coin(atomDenom, 100_000000),
			types.ErrUndercollaterized,
		},
		{
			"zero collateral",
			supplier,
			coin(atomDenom, 100_000000),
			types.ErrUndercollaterized,
		},
	}

	for _, tc := range tcs {
		if tc.err != nil {
			err := app.LeverageKeeper.Borrow(ctx, tc.addr, tc.coin)
			require.ErrorIs(err, tc.err, tc.msg)
		} else {
			// initial state
			iBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, tc.coin.Denom)
			iBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify the output of borrow function
			err := app.LeverageKeeper.Borrow(ctx, tc.addr, tc.coin)
			require.NoError(err, tc.msg)

			// final state
			fBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, tc.coin.Denom)
			fBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify token balance is increased by expected amount
			require.Equal(iBalance.Add(tc.coin), fBalance, tc.msg, "balances")
			// verify uToken collateral unchanged
			require.Equal(iCollateral, fCollateral, tc.msg, "collateral")
			// verify uToken supply is unchanged
			require.Equal(iUTokenSupply, fUTokenSupply, tc.msg, "uToken supply")
			// verify uToken exchange rate is unchanged
			require.Equal(iExchangeRate, fExchangeRate, tc.msg, "uToken exchange rate")
			// verify borrowed coins increased by expected amount
			require.Equal(iBorrowed.Add(tc.coin), fBorrowed, "borrowed coins")

			// check all available invariants
			s.checkInvariants(tc.msg)
		}
	}
}

func (s *IntegrationTestSuite) TestRepay() {
	type testCase struct {
		msg           string
		addr          sdk.AccAddress
		coin          sdk.Coin
		expectedRepay sdk.Coin
		err           error
	}

	app, ctx, require := s.app, s.ctx, s.Require()

	// create and fund a borrower which supplies and collateralizes UMEE, then borrows 10 UMEE
	borrower := s.newAccount(coin(umeeDenom, 200_000000))
	s.supply(borrower, coin(umeeDenom, 150_000000))
	s.collateralize(borrower, coin("u/"+umeeDenom, 120_000000))
	s.borrow(borrower, coin(umeeDenom, 10_000000))

	// create and fund a borrower which engages in a supply->borrow->supply loop
	looper := s.newAccount(coin(umeeDenom, 50_000000))
	s.supply(looper, coin(umeeDenom, 50_000000))
	s.collateralize(looper, coin("u/"+umeeDenom, 50_000000))
	s.borrow(looper, coin(umeeDenom, 5_000000))
	s.supply(looper, coin(umeeDenom, 5_000000))

	tcs := []testCase{
		{
			"uToken",
			borrower,
			coin("u/"+umeeDenom, 100_000000),
			sdk.Coin{},
			types.ErrUToken,
		},
		{
			"unregistered token",
			borrower,
			coin("abcd", 100_000000),
			sdk.Coin{},
			types.ErrDenomNotBorrowed,
		},
		{
			"not borrowed",
			borrower,
			coin(atomDenom, 100_000000),
			sdk.Coin{},
			types.ErrDenomNotBorrowed,
		},
		{
			"valid repay",
			borrower,
			coin(umeeDenom, 1_000000),
			coin(umeeDenom, 1_000000),
			nil,
		},
		{
			"additional repay",
			borrower,
			coin(umeeDenom, 3_000000),
			coin(umeeDenom, 3_000000),
			nil,
		},
		{
			"overpay",
			borrower,
			coin(umeeDenom, 30_000000),
			coin(umeeDenom, 6_000000),
			nil,
		},
		{
			"insufficient balance",
			looper,
			coin(umeeDenom, 1_000000),
			sdk.Coin{},
			sdkerrors.ErrInsufficientFunds,
		},
	}

	for _, tc := range tcs {
		if tc.err != nil {
			_, err := app.LeverageKeeper.Repay(ctx, tc.addr, tc.coin)
			require.ErrorIs(err, tc.err, tc.msg)
		} else {
			// initial state
			iBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, tc.coin.Denom)
			iBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify the output of repay function
			repaid, err := app.LeverageKeeper.Repay(ctx, tc.addr, tc.coin)
			require.NoError(err, tc.msg)
			require.Equal(tc.expectedRepay, repaid, tc.msg)

			// final state
			fBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, tc.coin.Denom)
			fBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify token balance is decreased by expected amount
			require.Equal(iBalance.Sub(tc.expectedRepay), fBalance, tc.msg, "balances")
			// verify uToken collateral unchanged
			require.Equal(iCollateral, fCollateral, tc.msg, "collateral")
			// verify uToken supply is unchanged
			require.Equal(iUTokenSupply, fUTokenSupply, tc.msg, "uToken supply")
			// verify uToken exchange rate is unchanged
			require.Equal(iExchangeRate, fExchangeRate, tc.msg, "uToken exchange rate")
			// verify borrowed coins decreased by expected amount
			s.requireEqualCoins(iBorrowed.Sub(tc.expectedRepay), fBorrowed, "borrowed coins")

			// check all available invariants
			s.checkInvariants(tc.msg)
		}
	}
}

func (s *IntegrationTestSuite) TestLiquidate() {
	type testCase struct {
		msg               string
		liquidator        sdk.AccAddress
		borrower          sdk.AccAddress
		attemptedRepay    sdk.Coin
		rewardDenom       string
		expectedRepay     sdk.Coin
		expectedLiquidate sdk.Coin
		expectedReward    sdk.Coin
		err               error
	}

	app, ctx, require := s.app, s.ctx, s.Require()

	// create and fund a liquidator which supplies plenty of UMEE and ATOM to the module
	supplier := s.newAccount(coin(umeeDenom, 1000_000000), coin(atomDenom, 1000_000000))
	s.supply(supplier, coin(umeeDenom, 1000_000000), coin(atomDenom, 1000_000000))

	// create and fund a liquidator which has 1000 UMEE and 1000 ATOM
	liquidator := s.newAccount(coin(umeeDenom, 1000_000000), coin(atomDenom, 1000_000000))

	// create a healthy borrower
	healthyBorrower := s.newAccount(coin(umeeDenom, 100_000000))
	s.supply(healthyBorrower, coin(umeeDenom, 100_000000))
	s.collateralize(healthyBorrower, coin("u/"+umeeDenom, 100_000000))
	s.borrow(healthyBorrower, coin(umeeDenom, 10_000000))

	// create a borrower which supplies and collateralizes 1000 ATOM
	atomBorrower := s.newAccount(coin(atomDenom, 1000_000000))
	s.supply(atomBorrower, coin(atomDenom, 1000_000000))
	s.collateralize(atomBorrower, coin("u/"+atomDenom, 1000_000000))
	// artificially borrow 500 ATOM - this can be liquidated without bad debt
	s.forceBorrow(atomBorrower, coin(atomDenom, 500_000000))

	// create a borrower which collateralizes 110 UMEE
	umeeBorrower := s.newAccount(coin(umeeDenom, 300_000000))
	s.supply(umeeBorrower, coin(umeeDenom, 200_000000))
	s.collateralize(umeeBorrower, coin("u/"+umeeDenom, 110_000000))
	// artificially borrow 200 UMEE - this will create a bad debt when liquidated
	s.forceBorrow(umeeBorrower, coin(umeeDenom, 200_000000))

	// creates a complex borrower with multiple denoms active
	complexBorrower := s.newAccount(coin(umeeDenom, 100_000000), coin(atomDenom, 100_000000))
	s.supply(complexBorrower, coin(umeeDenom, 100_000000), coin(atomDenom, 100_000000))
	s.collateralize(complexBorrower, coin("u/"+umeeDenom, 100_000000), coin("u/"+atomDenom, 100_000000))
	// artificially borrow multiple denoms
	s.forceBorrow(complexBorrower, coin(atomDenom, 30_000000), coin(umeeDenom, 30_000000))

	// creates a realistic borrower with 400 UMEE collateral which will have a close factor < 1
	closeBorrower := s.newAccount(coin(umeeDenom, 400_000000))
	s.supply(closeBorrower, coin(umeeDenom, 400_000000))
	s.collateralize(closeBorrower, coin("u/"+umeeDenom, 400_000000))
	// artificially borrow just barely above liquidation threshold to simulate interest accruing
	s.forceBorrow(closeBorrower, coin(umeeDenom, 102_000000))

	tcs := []testCase{
		{
			"healthy borrower",
			liquidator,
			healthyBorrower,
			coin(atomDenom, 1_000000),
			atomDenom,
			sdk.Coin{},
			sdk.Coin{},
			sdk.Coin{},
			types.ErrLiquidationIneligible,
		},
		{
			"not borrowed denom",
			liquidator,
			umeeBorrower,
			coin(atomDenom, 1_000000),
			atomDenom,
			sdk.Coin{},
			sdk.Coin{},
			sdk.Coin{},
			types.ErrLiquidationRepayZero,
		},
		{
			"direct atom liquidation",
			liquidator,
			atomBorrower,
			coin(atomDenom, 100_000000),
			atomDenom,
			coin(atomDenom, 100_000000),
			coin("u/"+atomDenom, 109_000000),
			coin(atomDenom, 109_000000),
			nil,
		},
		{
			"u/atom liquidation",
			liquidator,
			atomBorrower,
			coin(atomDenom, 100_000000),
			"u/" + atomDenom,
			coin(atomDenom, 100_000000),
			coin("u/"+atomDenom, 110_000000),
			coin("u/"+atomDenom, 110_000000),
			nil,
		},
		{
			"complete u/atom liquidation",
			liquidator,
			atomBorrower,
			coin(atomDenom, 500_000000),
			"u/" + atomDenom,
			coin(atomDenom, 300_000000),
			coin("u/"+atomDenom, 330_000000),
			coin("u/"+atomDenom, 330_000000),
			nil,
		},
		{
			"bad debt u/umee liquidation",
			liquidator,
			umeeBorrower,
			coin(umeeDenom, 200_000000),
			"u/" + umeeDenom,
			coin(umeeDenom, 100_000000),
			coin("u/"+umeeDenom, 110_000000),
			coin("u/"+umeeDenom, 110_000000),
			nil,
		},
		{
			"complex borrower",
			liquidator,
			complexBorrower,
			coin(umeeDenom, 200_000000),
			"u/" + atomDenom,
			coin(umeeDenom, 30_000000),
			coin("u/"+atomDenom, 3_527932),
			coin("u/"+atomDenom, 3_527932),
			nil,
		},
		{
			"close factor < 1",
			liquidator,
			closeBorrower,
			coin(umeeDenom, 200_000000),
			"u/" + umeeDenom,
			coin(umeeDenom, 7_752000),
			coin("u/"+umeeDenom, 8_527200),
			coin("u/"+umeeDenom, 8_527200),
			nil,
		},
	}

	for _, tc := range tcs {
		if tc.err != nil {
			_, _, _, err := app.LeverageKeeper.Liquidate(
				ctx, tc.liquidator, tc.borrower, tc.attemptedRepay, tc.rewardDenom,
			)
			require.ErrorIs(err, tc.err, tc.msg)
		} else {
			baseRewardDenom := types.ToTokenDenom(tc.expectedLiquidate.Denom)

			// initial state (borrowed denom)
			biUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			biExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, tc.attemptedRepay.Denom)

			// initial state (liquidated denom)
			liUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			liExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, baseRewardDenom)

			// borrower initial state
			biBalance := app.BankKeeper.GetAllBalances(ctx, tc.borrower)
			biCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.borrower)
			biBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.borrower)

			// liquidator initial state
			liBalance := app.BankKeeper.GetAllBalances(ctx, tc.liquidator)
			liCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.liquidator)
			liBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.liquidator)

			// verify the output of liquidate function
			repaid, liquidated, reward, err := app.LeverageKeeper.Liquidate(
				ctx, tc.liquidator, tc.borrower, tc.attemptedRepay, tc.rewardDenom,
			)
			require.NoError(err, tc.msg)
			require.Equal(tc.expectedRepay, repaid, tc.msg, "repaid")
			require.Equal(tc.expectedLiquidate, liquidated, tc.msg, "liquidated")
			require.Equal(tc.expectedReward, reward, tc.msg, "reward")

			// final state (liquidated denom)
			lfUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			lfExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, baseRewardDenom)

			// borrower final state
			bfBalance := app.BankKeeper.GetAllBalances(ctx, tc.borrower)
			bfCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.borrower)
			bfBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.borrower)

			// liquidator final state
			lfBalance := app.BankKeeper.GetAllBalances(ctx, tc.liquidator)
			lfCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.liquidator)
			lfBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.liquidator)

			// if borrowed denom and reward denom are different, then borrowed denom uTokens should be unaffected
			if tc.rewardDenom != tc.attemptedRepay.Denom {
				// final state (borrowed denom)
				bfUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
				bfExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, tc.attemptedRepay.Denom)

				// verify borrowed denom uToken supply is unchanged
				require.Equal(biUTokenSupply, bfUTokenSupply, tc.msg, "uToken supply (borrowed denom")
				// verify borrowed denom uToken exchange rate is unchanged
				require.Equal(biExchangeRate, bfExchangeRate, tc.msg, "uToken exchange rate (borrowed denom")
			}

			// verify liquidated denom uToken supply is unchanged if indirect liquidation, or reduced if direct
			expectedLiquidatedUTokenSupply := liUTokenSupply
			if !types.HasUTokenPrefix(tc.rewardDenom) {
				expectedLiquidatedUTokenSupply = expectedLiquidatedUTokenSupply.Sub(tc.expectedLiquidate)
			}
			require.Equal(expectedLiquidatedUTokenSupply, lfUTokenSupply, tc.msg, "uToken supply (liquidated denom")
			// verify liquidated denom uToken exchange rate is unchanged
			require.Equal(liExchangeRate, lfExchangeRate, tc.msg, "uToken exchange rate (liquidated denom")

			// verify borrower balances unchanged
			require.Equal(biBalance, bfBalance, tc.msg, "borrower balances")
			// verify borrower collateral reduced by the expected amount
			s.requireEqualCoins(biCollateral.Sub(tc.expectedLiquidate), bfCollateral, tc.msg, "borrower collateral")
			// verify borrowed coins decreased by expected amount
			s.requireEqualCoins(biBorrowed.Sub(tc.expectedRepay), bfBorrowed, "borrowed coins")

			// verify liquidator balance changes by expected amounts
			require.Equal(liBalance.Sub(tc.expectedRepay).Add(tc.expectedReward), lfBalance,
				tc.msg, "liquidator balances")
			// verify liquidator collateral unchanged
			require.Equal(liCollateral, lfCollateral, tc.msg, "liquidator collateral")
			// verify liquidator borrowed coins unchanged
			s.requireEqualCoins(liBorrowed, lfBorrowed, "liquidator borrowed coins")

			// check all available invariants
			s.checkInvariants(tc.msg)
		}
	}
}

func (s *IntegrationTestSuite) TestMaxCollateralShare() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// update initial ATOM to have a limited MaxCollateralShare
	atom, err := app.LeverageKeeper.GetTokenSettings(ctx, atomDenom)
	require.NoError(err)
	atom.MaxCollateralShare = sdk.MustNewDecFromStr("0.1")
	s.registerToken(atom)

	// Mock oracle prices:
	// UMEE $4.21
	// ATOM $39.38

	// create a supplier to collateralize 100 UMEE, worth $421.00
	umeeSupplier := s.newAccount(coin(umeeDenom, 100_000000))
	s.supply(umeeSupplier, coin(umeeDenom, 100_000000))
	s.collateralize(umeeSupplier, coin("u/"+umeeDenom, 100_000000))

	// create an ATOM supplier
	atomSupplier := s.newAccount(coin(atomDenom, 100_000000))
	s.supply(atomSupplier, coin(atomDenom, 100_000000))

	// collateralize 1.18 ATOM, worth $46.46, with no error.
	// total collateral value (across all denoms) will be $467.46
	// so ATOM's collateral share ($46.46 / $467.46) is barely below 10%
	s.collateralize(atomSupplier, coin("u/"+atomDenom, 1_180000))

	// attempt to collateralize another 0.01 ATOM, which would result in too much collateral share for ATOM
	err = app.LeverageKeeper.Collateralize(ctx, atomSupplier, coin("u/"+atomDenom, 10000))
	require.ErrorIs(err, types.ErrMaxCollateralShare)
}

func (s *IntegrationTestSuite) TestMinCollateralLiquidity() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// update initial UMEE to have a limited MinCollateralLiquidity
	umee, err := app.LeverageKeeper.GetTokenSettings(ctx, umeeDenom)
	require.NoError(err)
	umee.MinCollateralLiquidity = sdk.MustNewDecFromStr("0.5")
	s.registerToken(umee)

	// create a supplier to collateralize 100 UMEE
	umeeSupplier := s.newAccount(coin(umeeDenom, 100_000000))
	s.supply(umeeSupplier, coin(umeeDenom, 100_000000))
	s.collateralize(umeeSupplier, coin("u/"+umeeDenom, 100_000000))

	// create an ATOM supplier and borrow 49 UMEE
	atomSupplier := s.newAccount(coin(atomDenom, 100_000000))
	s.supply(atomSupplier, coin(atomDenom, 100_000000))
	s.collateralize(atomSupplier, coin("u/"+atomDenom, 100_000000))
	s.borrow(atomSupplier, coin(umeeDenom, 49_000000))

	// collateral liquidity (liquidity / collateral) of UMEE is 51/100

	// withdrawal would reduce collateral liquidity to 41/90
	_, err = app.LeverageKeeper.Withdraw(ctx, umeeSupplier, coin("u/"+umeeDenom, 10_000000))
	require.ErrorIs(err, types.ErrMinCollateralLiquidity, "withdraw")

	// borrow would reduce collateral liquidity to 41/100
	err = app.LeverageKeeper.Borrow(ctx, umeeSupplier, coin(umeeDenom, 10_000000))
	require.ErrorIs(err, types.ErrMinCollateralLiquidity, "borrow")
}

func (s *IntegrationTestSuite) TestMinCollateralLiquidity_Collateralize() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// update initial UMEE to have a limited MinCollateralLiquidity
	umee, err := app.LeverageKeeper.GetTokenSettings(ctx, umeeDenom)
	require.NoError(err)
	umee.MinCollateralLiquidity = sdk.MustNewDecFromStr("0.5")
	s.registerToken(umee)

	// create a supplier to supply 200 UMEE, and collateralize 100 UMEE
	umeeSupplier := s.newAccount(coin(umeeDenom, 200))
	s.supply(umeeSupplier, coin(umeeDenom, 200))
	s.collateralize(umeeSupplier, coin("u/"+umeeDenom, 100))

	// create an ATOM supplier and borrow 149 UMEE
	atomSupplier := s.newAccount(coin(atomDenom, 100))
	s.supply(atomSupplier, coin(atomDenom, 100))
	s.collateralize(atomSupplier, coin("u/"+atomDenom, 100))
	s.borrow(atomSupplier, coin(umeeDenom, 149))

	// collateral liquidity (liquidity / collateral) of UMEE is 51/100

	// collateralize would reduce collateral liquidity to 51/200
	err = app.LeverageKeeper.Collateralize(ctx, umeeSupplier, coin("u/"+umeeDenom, 100))
	require.ErrorIs(err, types.ErrMinCollateralLiquidity, "collateralize")
}
