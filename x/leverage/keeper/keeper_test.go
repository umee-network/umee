package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	umeeapp "github.com/umee-network/umee/v2/app"
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
			coin(umeeDenom, 30_000000),
			nil,
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

	// create and fund a liquidator which has 1000 UMEE and 40 ATOM
	liquidator := s.newAccount(coin(umeeDenom, 1000_000000), coin(atomDenom, 40_000000))

	// create a healthy borrower
	healthyBorrower := s.newAccount(coin(umeeDenom, 100_000000))
	s.supply(healthyBorrower, coin(umeeDenom, 100_000000))
	s.collateralize(healthyBorrower, coin("u/"+umeeDenom, 100_000000))
	s.borrow(healthyBorrower, coin(umeeDenom, 10_000000))

	// create a borrower which supplies and collateralizes 100 ATOM
	atomBorrower := s.newAccount(coin(atomDenom, 100_000000))
	s.supply(atomBorrower, coin(atomDenom, 100_000000))
	s.collateralize(atomBorrower, coin("u/"+atomDenom, 100_000000))
	// artificially borrow 50 ATOM
	s.forceBorrow(atomBorrower, coin(atomDenom, 50_000000))

	// create a borrower which supplies and collateralizes 100 UMEE
	umeeBorrower := s.newAccount(coin(umeeDenom, 100_000000))
	s.supply(umeeBorrower, coin(umeeDenom, 100_000000))
	s.collateralize(umeeBorrower, coin("u/"+umeeDenom, 100_000000))
	// artificially borrow 200 UMEE
	s.forceBorrow(umeeBorrower, coin(umeeDenom, 200_000000))

	// TODO: complex liquidations
	// TODO: partial liquidations
	// TODO: zero liquidations
	// TODO: expect bad debt repayment
	// TODO: expect reserve exhaustion
	// TODO: direct and indirect liquidation
	// TODO: close factor limited (use atom borrower)

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
			"direct atom liquidation",
			liquidator,
			atomBorrower,
			coin(atomDenom, 1_000000),
			atomDenom,
			coin(atomDenom, 1_000000),
			coin("u/"+atomDenom, 1_090000),
			coin(atomDenom, 1_090000),
			nil,
		},
		{
			"u/atom liquidation",
			liquidator,
			atomBorrower,
			coin(atomDenom, 1_000000),
			"u/" + atomDenom,
			coin(atomDenom, 1_000000),
			coin("u/"+atomDenom, 1_100000),
			coin("u/"+atomDenom, 1_100000),
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
			require.Equal(biCollateral.Sub(tc.expectedLiquidate), bfCollateral, tc.msg, "borrower collateral")
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

func (s *IntegrationTestSuite) TestLiquidate_Old() {
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
