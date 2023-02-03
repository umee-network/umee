package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/leverage/fixtures"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

func (s *IntegrationTestSuite) TestAddTokensToRegistry() {
	govAccAddr := s.app.GovKeeper.GetGovernanceAccount(s.ctx).GetAddress().String()
	registeredUmee := fixtures.Token("uumee", "UMEE", 6)
	newTokens := fixtures.Token("uabcd", "ABCD", 6)

	testCases := []struct {
		name      string
		req       *types.MsgGovUpdateRegistry
		expectErr bool
		errMsg    string
	}{
		{
			"invalid token data",
			&types.MsgGovUpdateRegistry{
				Authority:   govAccAddr,
				Title:       "test",
				Description: "test",
				AddTokens: []types.Token{
					fixtures.Token("uosmo", "", 6), // empty denom is invalid
				},
			},
			true,
			"invalid denom",
		}, {
			"unauthorized authority address",
			&types.MsgGovUpdateRegistry{
				Authority:   s.addrs[0].String(),
				Title:       "test",
				Description: "test",
				AddTokens: []types.Token{
					newTokens,
				},
			},
			true,
			"invalid authority",
		}, {
			"already registered token",
			&types.MsgGovUpdateRegistry{
				Authority:   govAccAddr,
				Title:       "test",
				Description: "test",
				AddTokens: []types.Token{
					registeredUmee,
				},
			},
			true,
			fmt.Sprintf("token %s is already registered", registeredUmee.BaseDenom),
		}, {
			"valid authority and valid token for registry",
			&types.MsgGovUpdateRegistry{
				Authority:   govAccAddr,
				Title:       "test",
				Description: "test",
				AddTokens: []types.Token{
					newTokens,
				},
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := tc.req.ValidateBasic()
			if err == nil {
				_, err = s.msgSrvr.GovUpdateRegistry(s.ctx, tc.req)
			}
			if tc.expectErr {
				s.Require().ErrorContains(err, tc.errMsg)
			} else {
				s.Require().NoError(err)
				// no tokens should have been deleted
				tokens := s.app.LeverageKeeper.GetAllRegisteredTokens(s.ctx)
				s.Require().Len(tokens, 6)

				token, err := s.app.LeverageKeeper.GetTokenSettings(s.ctx, "uabcd")
				s.Require().NoError(err)
				s.Require().Equal(token.BaseDenom, "uabcd")
			}
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateRegistry() {
	govAccAddr := s.app.GovKeeper.GetGovernanceAccount(s.ctx).GetAddress().String()
	modifiedUmee := fixtures.Token("uumee", "UMEE", 6)
	modifiedUmee.ReserveFactor = sdk.MustNewDecFromStr("0.69")

	testCases := []struct {
		name      string
		req       *types.MsgGovUpdateRegistry
		expectErr bool
		errMsg    string
	}{
		{
			"invalid token data",
			&types.MsgGovUpdateRegistry{
				Authority:   govAccAddr,
				Title:       "test",
				Description: "test",
				UpdateTokens: []types.Token{
					fixtures.Token("uosmo", "", 6), // empty denom is invalid
				},
			},
			true,
			"invalid denom",
		}, {
			"unauthorized authority address",
			&types.MsgGovUpdateRegistry{
				Authority:   s.addrs[0].String(),
				Title:       "test",
				Description: "test",
				UpdateTokens: []types.Token{
					fixtures.Token("uosmo", "", 6), // empty denom is invalid
				},
			},
			true,
			"invalid authority",
		}, {
			"valid authority and valid update token registry",
			&types.MsgGovUpdateRegistry{
				Authority:   govAccAddr,
				Title:       "test",
				Description: "test",
				UpdateTokens: []types.Token{
					modifiedUmee,
				},
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := tc.req.ValidateBasic()
			if err == nil {
				_, err = s.msgSrvr.GovUpdateRegistry(s.ctx, tc.req)
			}
			if tc.expectErr {
				s.Require().ErrorContains(err, tc.errMsg)
			} else {
				s.Require().NoError(err)
				// no tokens should have been deleted
				tokens := s.app.LeverageKeeper.GetAllRegisteredTokens(s.ctx)
				s.Require().Len(tokens, 5)

				token, err := s.app.LeverageKeeper.GetTokenSettings(s.ctx, "uumee")
				s.Require().NoError(err)
				s.Require().Equal("0.690000000000000000", token.ReserveFactor.String(),
					"reserve factor is correctly set")

				_, err = s.app.LeverageKeeper.GetTokenSettings(s.ctx, fixtures.AtomDenom)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestMsgSupply() {
	type testCase struct {
		msg             string
		addr            sdk.AccAddress
		coin            sdk.Coin
		expectedUTokens sdk.Coin
		err             error
	}

	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// create and fund a supplier with 100 UMEE and 100 ATOM
	supplier := s.newAccount(mkCoin(umeeDenom, 100_000000), mkCoin(atomDenom, 100_000000))

	// create and modify a borrower to force the uToken exchange rate of ATOM from 1 to 1.5
	borrower := s.newAccount(mkCoin(atomDenom, 100_000000))
	s.supply(borrower, mkCoin(atomDenom, 100_000000))
	s.collateralize(borrower, mkCoin("u/"+atomDenom, 100_000000))
	s.borrow(borrower, mkCoin(atomDenom, 10_000000))
	s.tk.SetBorrow(ctx, borrower, mkCoin(atomDenom, 60_000000))

	// create a supplier that will exceed token's default MaxSupply
	whale := s.newAccount(mkCoin(umeeDenom, 1_000_000_000000))

	tcs := []testCase{
		{
			"unregistered denom",
			supplier,
			mkCoin("abcd", 80_000000),
			sdk.Coin{},
			types.ErrNotRegisteredToken,
		}, {
			"uToken",
			supplier,
			mkCoin("u/"+umeeDenom, 80_000000),
			sdk.Coin{},
			types.ErrUToken,
		}, {
			"no balance",
			borrower,
			mkCoin(umeeDenom, 20_000000),
			sdk.Coin{},
			sdkerrors.ErrInsufficientFunds,
		}, {
			"insufficient balance",
			supplier,
			mkCoin(umeeDenom, 120_000000),
			sdk.Coin{},
			sdkerrors.ErrInsufficientFunds,
		}, {
			"valid supply",
			supplier,
			mkCoin(umeeDenom, 80_000000),
			mkCoin("u/"+umeeDenom, 80_000000),
			nil,
		}, {
			"additional supply",
			supplier,
			mkCoin(umeeDenom, 20_000000),
			mkCoin("u/"+umeeDenom, 20_000000),
			nil,
		}, {
			"high exchange rate",
			supplier,
			mkCoin(atomDenom, 60_000000),
			mkCoin("u/"+atomDenom, 40_000000),
			nil,
		}, {
			"max supply",
			whale,
			mkCoin(umeeDenom, 1_000_000_000000),
			sdk.Coin{},
			types.ErrMaxSupply,
		},
	}

	for _, tc := range tcs {
		msg := &types.MsgSupply{
			Supplier: tc.addr.String(),
			Asset:    tc.coin,
		}
		if tc.err != nil {
			_, err := srv.Supply(ctx, msg)
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
			resp, err := srv.Supply(ctx, msg)
			require.NoError(err, tc.msg)
			require.Equal(tc.expectedUTokens, resp.Received, tc.msg)

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

func (s *IntegrationTestSuite) TestMsgWithdraw() {
	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// create and fund a supplier with 100 UMEE and 100 ATOM, then supply 100 UMEE and 50 ATOM
	// also collateralize 75 of supplied UMEE
	supplier := s.newAccount(mkCoin(umeeDenom, 100_000000), mkCoin(atomDenom, 100_000000))
	s.supply(supplier, mkCoin(umeeDenom, 100_000000))
	s.collateralize(supplier, mkCoin("u/"+umeeDenom, 75_000000))
	s.supply(supplier, mkCoin(atomDenom, 50_000000))

	// create and modify a borrower to force the uToken exchange rate of ATOM from 1 to 1.2
	borrower := s.newAccount(mkCoin(atomDenom, 100_000000))
	s.supply(borrower, mkCoin(atomDenom, 100_000000))
	s.collateralize(borrower, mkCoin("u/"+atomDenom, 100_000000))
	s.borrow(borrower, mkCoin(atomDenom, 10_000000))
	s.tk.SetBorrow(ctx, borrower, mkCoin(atomDenom, 40_000000))

	// create an additional supplier (UMEE, DUMP, PUMP tokens)
	other := s.newAccount(mkCoin(umeeDenom, 100_000000), mkCoin(dumpDenom, 100_000000), mkCoin(pumpDenom, 100_000000))
	s.supply(other, mkCoin(umeeDenom, 100_000000))
	s.supply(other, mkCoin(pumpDenom, 100_000000))
	s.supply(other, mkCoin(dumpDenom, 100_000000))

	// create a DUMP (historic price 1.00, current price 0.50) borrower
	// using PUMP (historic price 1.00, current price 2.00) collateral
	dumpborrower := s.newAccount(mkCoin(pumpDenom, 100_000000))
	s.supply(dumpborrower, mkCoin(pumpDenom, 100_000000))
	s.collateralize(dumpborrower, mkCoin("u/"+pumpDenom, 100_000000))
	s.borrow(dumpborrower, mkCoin(dumpDenom, 20_000000))
	// collateral value is $200 (current) or $100 (historic)
	// borrowed value is $10 (current) or $20 (historic)
	// collateral weights are always 0.25 in testing

	// create a PUMP (historic price 1.00, current price 2.00) borrower
	// using DUMP (historic price 1.00, current price 0.50) collateral
	pumpborrower := s.newAccount(mkCoin(dumpDenom, 100_000000))
	s.supply(pumpborrower, mkCoin(dumpDenom, 100_000000))
	s.collateralize(pumpborrower, mkCoin("u/"+dumpDenom, 100_000000))
	s.borrow(pumpborrower, mkCoin(pumpDenom, 5_000000))
	// collateral value is $50 (current) or $100 (historic)
	// borrowed value is $10 (current) or $5 (historic)
	// collateral weights are always 0.25 in testing

	tcs := []struct {
		msg                  string
		addr                 sdk.AccAddress
		uToken               sdk.Coin
		expectFromBalance    sdk.Coins
		expectFromCollateral sdk.Coins
		expectedTokens       sdk.Coin
		err                  error
	}{
		{
			"unregistered base token",
			supplier,
			mkCoin("abcd", 80_000000),
			nil,
			nil,
			sdk.Coin{},
			types.ErrNotUToken,
		}, {
			"only uToken can be withdrawn",
			supplier,
			mkCoin(umeeDenom, 80_000000),
			nil,
			nil,
			sdk.Coin{},
			types.ErrNotUToken,
		}, {
			"insufficient uTokens",
			supplier,
			mkCoin("u/"+umeeDenom, 120_000000),
			nil,
			nil,
			sdk.Coin{},
			types.ErrInsufficientBalance,
		}, {
			"withdraw from balance",
			supplier,
			mkCoin("u/"+umeeDenom, 10_000000),
			sdk.NewCoins(mkCoin("u/"+umeeDenom, 10_000000)),
			nil,
			mkCoin(umeeDenom, 10_000000),
			nil,
		}, {
			"some from collateral",
			supplier,
			mkCoin("u/"+umeeDenom, 80_000000),
			sdk.NewCoins(mkCoin("u/"+umeeDenom, 15_000000)),
			sdk.NewCoins(mkCoin("u/"+umeeDenom, 65_000000)),
			mkCoin(umeeDenom, 80_000000),
			nil,
		}, {
			"only from collateral",
			supplier,
			mkCoin("u/"+umeeDenom, 10_000000),
			nil,
			sdk.NewCoins(mkCoin("u/"+umeeDenom, 10_000000)),
			mkCoin(umeeDenom, 10_000000),
			nil,
		}, {
			"high exchange rate",
			supplier,
			mkCoin("u/"+atomDenom, 50_000000),
			sdk.NewCoins(mkCoin("u/"+atomDenom, 50_000000)),
			nil,
			mkCoin(atomDenom, 60_000000),
			nil,
		}, {
			"borrow limit",
			borrower,
			mkCoin("u/"+atomDenom, 50_000000),
			nil,
			nil,
			sdk.Coin{},
			types.ErrUndercollaterized,
		}, {
			"acceptable withdrawal (dump borrower)",
			dumpborrower,
			mkCoin("u/"+pumpDenom, 20_000000),
			nil,
			sdk.NewCoins(mkCoin("u/"+pumpDenom, 20_000000)),
			mkCoin(pumpDenom, 20_000000),
			nil,
		}, {
			"borrow limit (undercollateralized under historic prices but ok with current prices)",
			dumpborrower,
			mkCoin("u/"+pumpDenom, 20_000000),
			nil,
			nil,
			sdk.Coin{},
			types.ErrUndercollaterized,
		}, {
			"acceptable withdrawal (pump borrower)",
			pumpborrower,
			mkCoin("u/"+dumpDenom, 20_000000),
			nil,
			sdk.NewCoins(mkCoin("u/"+dumpDenom, 20_000000)),
			mkCoin(dumpDenom, 20_000000),
			nil,
		}, {
			"borrow limit (undercollateralized under current prices but ok with historic prices)",
			pumpborrower,
			mkCoin("u/"+dumpDenom, 20_000000),
			nil,
			nil,
			sdk.Coin{},
			types.ErrUndercollaterized,
		},
	}

	for _, tc := range tcs {
		msg := &types.MsgWithdraw{
			Supplier: tc.addr.String(),
			Asset:    tc.uToken,
		}
		if tc.err != nil {
			_, err := srv.Withdraw(ctx, msg)
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
			resp, err := srv.Withdraw(ctx, msg)
			require.NoError(err, tc.msg)
			require.Equal(tc.expectedTokens, resp.Received, tc.msg)

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

func (s *IntegrationTestSuite) TestMsgMaxWithdraw() {
	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// create and fund a supplier with 100 UMEE and 100 ATOM, then supply 100 UMEE and 50 ATOM
	// also collateralize 75 of supplied UMEE
	supplier := s.newAccount(mkCoin(umeeDenom, 100_000000), mkCoin(atomDenom, 100_000000))
	s.supply(supplier, mkCoin(umeeDenom, 100_000000))
	s.collateralize(supplier, mkCoin("u/"+umeeDenom, 75_000000))
	s.supply(supplier, mkCoin(atomDenom, 50_000000))

	// create and modify a borrower to force the uToken exchange rate of ATOM from 1 to 1.2
	borrower := s.newAccount(mkCoin(atomDenom, 100_000000))
	s.supply(borrower, mkCoin(atomDenom, 100_000000))
	s.collateralize(borrower, mkCoin("u/"+atomDenom, 100_000000))
	s.borrow(borrower, mkCoin(atomDenom, 10_000000))
	s.tk.SetBorrow(ctx, borrower, mkCoin(atomDenom, 40_000000))

	// create an additional UMEE supplier with a small borrow
	other := s.newAccount(mkCoin(umeeDenom, 100_000000))
	s.supply(other, mkCoin(umeeDenom, 100_000000))
	s.collateralize(other, mkCoin("u/"+umeeDenom, 100_000000))
	s.borrow(other, mkCoin(umeeDenom, 10_000000))

	// create an additional supplier (UMEE, DUMP, PUMP tokens)
	surplus := s.newAccount(mkCoin(umeeDenom, 100_000000), mkCoin(dumpDenom, 100_000000), mkCoin(pumpDenom, 100_000000))
	s.supply(surplus, mkCoin(umeeDenom, 100_000000))
	s.supply(surplus, mkCoin(pumpDenom, 100_000000))
	s.supply(surplus, mkCoin(dumpDenom, 100_000000))

	// create a DUMP (historic price 1.00, current price 0.50) borrower
	// using PUMP (historic price 1.00, current price 2.00) collateral
	dumpborrower := s.newAccount(mkCoin(pumpDenom, 100_000000))
	s.supply(dumpborrower, mkCoin(pumpDenom, 100_000000))
	s.collateralize(dumpborrower, mkCoin("u/"+pumpDenom, 100_000000))
	s.borrow(dumpborrower, mkCoin(dumpDenom, 20_000000))
	// collateral value is $200 (current) or $100 (historic)
	// borrowed value is $10 (current) or $20 (historic)
	// collateral weights are always 0.25 in testing

	// create a PUMP (historic price 1.00, current price 2.00) borrower
	// using DUMP (historic price 1.00, current price 0.50) collateral
	pumpborrower := s.newAccount(mkCoin(dumpDenom, 100_000000))
	s.supply(pumpborrower, mkCoin(dumpDenom, 100_000000))
	s.collateralize(pumpborrower, mkCoin("u/"+dumpDenom, 100_000000))
	s.borrow(pumpborrower, mkCoin(pumpDenom, 5_000000))
	// collateral value is $50 (current) or $100 (historic)
	// borrowed value is $10 (current) or $5 (historic)
	// collateral weights are always 0.25 in testing

	zeroUmee := coin.ZeroCoin(umeeDenom)
	zeroUUmee := mkCoin("u/"+umeeDenom, 0)
	tcs := []struct {
		msg                  string
		addr                 sdk.AccAddress
		denom                string
		expectedWithdraw     sdk.Coin
		expectFromCollateral sdk.Coin
		expectedTokens       sdk.Coin
		err                  error
	}{
		{
			"unregistered base token",
			supplier,
			"abcd",
			sdk.Coin{},
			sdk.Coin{},
			sdk.Coin{},
			types.ErrNotRegisteredToken,
		}, {
			"can't borrow uToken",
			supplier,
			"u/" + umeeDenom,
			sdk.Coin{},
			sdk.Coin{},
			sdk.Coin{},
			types.ErrUToken,
		}, {
			"max withdraw umee",
			supplier,
			umeeDenom,
			mkCoin("u/"+umeeDenom, 100_000000),
			mkCoin("u/"+umeeDenom, 75_000000),
			mkCoin(umeeDenom, 100_000000),
			nil,
		}, {
			"duplicate max withdraw umee",
			supplier,
			umeeDenom,
			zeroUUmee,
			zeroUUmee,
			zeroUmee,
			nil,
		}, {
			"max withdraw with borrow",
			other,
			umeeDenom,
			mkCoin("u/"+umeeDenom, 60_000000),
			mkCoin("u/"+umeeDenom, 60_000000),
			mkCoin(umeeDenom, 60_000000),
			nil,
		}, {
			"max withdrawal (dump borrower)",
			dumpborrower,
			pumpDenom,
			mkCoin("u/"+pumpDenom, 20_000000),
			mkCoin("u/"+pumpDenom, 20_000000),
			mkCoin(pumpDenom, 20_000000),
			nil,
		}, {
			"max withdrawal (pump borrower)",
			pumpborrower,
			dumpDenom,
			mkCoin("u/"+dumpDenom, 20_000000),
			mkCoin("u/"+dumpDenom, 20_000000),
			mkCoin(dumpDenom, 20_000000),
			nil,
		},
	}

	for _, tc := range tcs {
		msg := &types.MsgMaxWithdraw{
			Supplier: tc.addr.String(),
			Denom:    tc.denom,
		}
		if tc.err != nil {
			_, err := srv.MaxWithdraw(ctx, msg)
			require.ErrorIs(err, tc.err, tc.msg)
			continue
		}
		expectFromBalance := tc.expectedWithdraw.Sub(tc.expectFromCollateral)

		// initial state
		iBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
		iCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
		iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
		iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, tc.denom)
		iBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

		// verify the outputs of withdraw function
		resp, err := srv.MaxWithdraw(ctx, msg)
		require.NoError(err, tc.msg)
		require.Equal(tc.expectedWithdraw, resp.Withdrawn, tc.msg)
		require.Equal(tc.expectedTokens, resp.Received, tc.msg)

		// final state
		fBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
		fCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
		fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
		fExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, tc.denom)
		fBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

		// verify token balance increased by the expected amount
		require.Equal(iBalance.Add(tc.expectedTokens).Sub(expectFromBalance),
			fBalance, tc.msg, "token balance")
		// verify uToken collateral decreased by the expected amount
		s.requireEqualCoins(iCollateral.Sub(tc.expectFromCollateral), fCollateral,
			tc.msg, "uToken collateral")
		// verify uToken supply decreased by the expected amount
		require.Equal(iUTokenSupply.Sub(tc.expectedWithdraw), fUTokenSupply, tc.msg, "uToken supply")
		// verify uToken exchange rate is unchanged
		require.Equal(iExchangeRate, fExchangeRate, tc.msg, "uToken exchange rate")
		// verify borrowed coins are unchanged
		require.Equal(iBorrowed, fBorrowed, tc.msg, "borrowed coins")

		// check all available invariants
		s.checkInvariants(tc.msg)
	}
}

func (s *IntegrationTestSuite) TestMsgCollateralize() {
	type testCase struct {
		msg    string
		addr   sdk.AccAddress
		uToken sdk.Coin
		err    error
	}

	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// create and fund a supplier with 200 UMEE, then supply 100 UMEE
	supplier := s.newAccount(mkCoin(umeeDenom, 200_000000))
	s.supply(supplier, mkCoin(umeeDenom, 100_000000))

	// create and fund another supplier
	otherSupplier := s.newAccount(mkCoin(umeeDenom, 200_000000), mkCoin(atomDenom, 200_000000))
	s.supply(otherSupplier, mkCoin(umeeDenom, 200_000000), mkCoin(atomDenom, 200_000000))

	tcs := []testCase{
		{
			"base token",
			supplier,
			mkCoin(umeeDenom, 80_000000),
			types.ErrNotUToken,
		}, {
			"unregistered uToken",
			supplier,
			mkCoin("u/abcd", 80_000000),
			types.ErrNotRegisteredToken,
		}, {
			"wrong balance",
			supplier,
			mkCoin("u/"+atomDenom, 10_000000),
			sdkerrors.ErrInsufficientFunds,
		}, {
			"valid collateralize",
			supplier,
			mkCoin("u/"+umeeDenom, 80_000000),
			nil,
		}, {
			"additional collateralize",
			supplier,
			mkCoin("u/"+umeeDenom, 10_000000),
			nil,
		}, {
			"insufficient balance",
			supplier,
			mkCoin("u/"+umeeDenom, 40_000000),
			sdkerrors.ErrInsufficientFunds,
		},
	}

	for _, tc := range tcs {
		msg := &types.MsgCollateralize{
			Borrower: tc.addr.String(),
			Asset:    tc.uToken,
		}
		if tc.err != nil {
			_, err := srv.Collateralize(ctx, msg)
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
			resp, err := srv.Collateralize(ctx, msg)
			require.NoError(err, tc.msg)
			require.Equal(&types.MsgCollateralizeResponse{}, resp, tc.msg)

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

func (s *IntegrationTestSuite) TestMsgDecollateralize() {
	type testCase struct {
		msg    string
		addr   sdk.AccAddress
		uToken sdk.Coin
		err    error
	}

	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// create and fund a supplier with 200 UMEE, then supply and collateralize 100 UMEE
	supplier := s.newAccount(mkCoin(umeeDenom, 200_000000))
	s.supply(supplier, mkCoin(umeeDenom, 100_000000))
	s.collateralize(supplier, mkCoin("u/"+umeeDenom, 100_000000))

	// create a borrower which supplies, collateralizes, then borrows ATOM
	borrower := s.newAccount(mkCoin(atomDenom, 100_000000))
	s.supply(borrower, mkCoin(atomDenom, 100_000000))
	s.collateralize(borrower, mkCoin("u/"+atomDenom, 100_000000))
	s.borrow(borrower, mkCoin(atomDenom, 10_000000))

	// create an additional supplier (UMEE, DUMP, PUMP tokens)
	surplus := s.newAccount(mkCoin(umeeDenom, 100_000000), mkCoin(dumpDenom, 100_000000), mkCoin(pumpDenom, 100_000000))
	s.supply(surplus, mkCoin(umeeDenom, 100_000000))
	s.supply(surplus, mkCoin(pumpDenom, 100_000000))
	s.supply(surplus, mkCoin(dumpDenom, 100_000000))

	// create a DUMP (historic price 1.00, current price 0.50) borrower
	// using PUMP (historic price 1.00, current price 2.00) collateral
	dumpborrower := s.newAccount(mkCoin(pumpDenom, 100_000000))
	s.supply(dumpborrower, mkCoin(pumpDenom, 100_000000))
	s.collateralize(dumpborrower, mkCoin("u/"+pumpDenom, 100_000000))
	s.borrow(dumpborrower, mkCoin(dumpDenom, 20_000000))
	// collateral value is $200 (current) or $100 (historic)
	// borrowed value is $10 (current) or $20 (historic)
	// collateral weights are always 0.25 in testing

	// create a PUMP (historic price 1.00, current price 2.00) borrower
	// using DUMP (historic price 1.00, current price 0.50) collateral
	pumpborrower := s.newAccount(mkCoin(dumpDenom, 100_000000))
	s.supply(pumpborrower, mkCoin(dumpDenom, 100_000000))
	s.collateralize(pumpborrower, mkCoin("u/"+dumpDenom, 100_000000))
	s.borrow(pumpborrower, mkCoin(pumpDenom, 5_000000))
	// collateral value is $50 (current) or $100 (historic)
	// borrowed value is $10 (current) or $5 (historic)
	// collateral weights are always 0.25 in testing

	tcs := []testCase{
		{
			"base token",
			supplier,
			mkCoin(umeeDenom, 80_000000),
			types.ErrNotUToken,
		}, {
			"no collateral",
			supplier,
			mkCoin("u/"+atomDenom, 40_000000),
			types.ErrInsufficientCollateral,
		}, {
			"valid decollateralize",
			supplier,
			mkCoin("u/"+umeeDenom, 80_000000),
			nil,
		}, {
			"additional decollateralize",
			supplier,
			mkCoin("u/"+umeeDenom, 10_000000),
			nil,
		}, {
			"insufficient collateral",
			supplier,
			mkCoin("u/"+umeeDenom, 40_000000),
			types.ErrInsufficientCollateral,
		}, {
			"above borrow limit",
			borrower,
			mkCoin("u/"+atomDenom, 100_000000),
			types.ErrUndercollaterized,
		},

		{
			"acceptable decollateralize (dump borrower)",
			dumpborrower,
			mkCoin("u/"+pumpDenom, 20_000000),
			nil,
		}, {
			"above borrow limit (undercollateralized under historic prices but ok with current prices)",
			dumpborrower,
			mkCoin("u/"+pumpDenom, 20_000000),
			types.ErrUndercollaterized,
		}, {
			"acceptable decollateralize (pump borrower)",
			pumpborrower,
			mkCoin("u/"+dumpDenom, 20_000000),
			nil,
		}, {
			"above borrow limit (undercollateralized under current prices but ok with historic prices)",
			pumpborrower,
			mkCoin("u/"+dumpDenom, 20_000000),
			types.ErrUndercollaterized,
		},
	}

	for _, tc := range tcs {
		msg := &types.MsgDecollateralize{
			Borrower: tc.addr.String(),
			Asset:    tc.uToken,
		}
		if tc.err != nil {
			_, err := srv.Decollateralize(ctx, msg)
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
			resp, err := srv.Decollateralize(ctx, msg)
			require.NoError(err, tc.msg)
			require.Equal(&types.MsgDecollateralizeResponse{}, resp, tc.msg)

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

func (s *IntegrationTestSuite) TestMsgSupplyCollateral() {
	type testCase struct {
		msg             string
		addr            sdk.AccAddress
		coin            sdk.Coin
		expectedUTokens sdk.Coin
		err             error
	}

	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// create and fund a supplier with 100 UMEE and 100 ATOM
	supplier := s.newAccount(mkCoin(umeeDenom, 100_000000), mkCoin(atomDenom, 100_000000))

	// create and modify a borrower to force the uToken exchange rate of ATOM from 1 to 1.5
	borrower := s.newAccount(mkCoin(atomDenom, 100_000000))
	s.supply(borrower, mkCoin(atomDenom, 100_000000))
	s.collateralize(borrower, mkCoin("u/"+atomDenom, 100_000000))
	s.borrow(borrower, mkCoin(atomDenom, 10_000000))
	s.tk.SetBorrow(ctx, borrower, mkCoin(atomDenom, 60_000000))

	// create a supplier that will exceed token's default MaxSupply
	whale := s.newAccount(mkCoin(umeeDenom, 1_000_000_000000))

	tcs := []testCase{
		{
			"unregistered denom",
			supplier,
			mkCoin("abcd", 80_000000),
			sdk.Coin{},
			types.ErrNotRegisteredToken,
		}, {
			"uToken",
			supplier,
			mkCoin("u/"+umeeDenom, 80_000000),
			sdk.Coin{},
			types.ErrUToken,
		}, {
			"no balance",
			borrower,
			mkCoin(umeeDenom, 20_000000),
			sdk.Coin{},
			sdkerrors.ErrInsufficientFunds,
		}, {
			"insufficient balance",
			supplier,
			mkCoin(umeeDenom, 120_000000),
			sdk.Coin{},
			sdkerrors.ErrInsufficientFunds,
		}, {
			"valid supply",
			supplier,
			mkCoin(umeeDenom, 80_000000),
			mkCoin("u/"+umeeDenom, 80_000000),
			nil,
		}, {
			"additional supply",
			supplier,
			mkCoin(umeeDenom, 20_000000),
			mkCoin("u/"+umeeDenom, 20_000000),
			nil,
		}, {
			"high exchange rate",
			supplier,
			mkCoin(atomDenom, 60_000000),
			mkCoin("u/"+atomDenom, 40_000000),
			nil,
		}, {
			"max supply",
			whale,
			mkCoin(umeeDenom, 1_000_000_000000),
			sdk.Coin{},
			types.ErrMaxSupply,
		},
	}

	for _, tc := range tcs {
		msg := &types.MsgSupplyCollateral{
			Supplier: tc.addr.String(),
			Asset:    tc.coin,
		}
		if tc.err != nil {
			_, err := srv.SupplyCollateral(ctx, msg)
			require.ErrorIs(err, tc.err, tc.msg)
		} else {
			denom := tc.coin.Denom

			// initial state
			iBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)
			iBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify the outputs of supply collateral function
			resp, err := srv.SupplyCollateral(ctx, msg)
			require.NoError(err, tc.msg)
			require.Equal(tc.expectedUTokens, resp.Collateralized, tc.msg)

			// final state
			fBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, denom)
			fBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify token balance decreased and uToken balance unchanged
			require.Equal(iBalance.Sub(tc.coin), fBalance, tc.msg, "token balance")
			// verify uToken collateral increaaed
			require.Equal(iCollateral.Add(tc.expectedUTokens), fCollateral, tc.msg, "uToken collateral")
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

func (s *IntegrationTestSuite) TestMsgBorrow() {
	type testCase struct {
		msg  string
		addr sdk.AccAddress
		coin sdk.Coin
		err  error
	}

	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// create and fund a supplier which supplies 100 UMEE and 100 ATOM
	supplier := s.newAccount(mkCoin(umeeDenom, 100_000000), mkCoin(atomDenom, 100_000000))
	s.supply(supplier, mkCoin(umeeDenom, 100_000000), mkCoin(atomDenom, 100_000000))

	// create a borrower which supplies and collateralizes 100 ATOM
	borrower := s.newAccount(mkCoin(atomDenom, 100_000000))
	s.supply(borrower, mkCoin(atomDenom, 100_000000))
	s.collateralize(borrower, mkCoin("u/"+atomDenom, 100_000000))

	// create an additional supplier (DUMP, PUMP tokens)
	surplus := s.newAccount(mkCoin(dumpDenom, 100_000000), mkCoin(pumpDenom, 100_000000))
	s.supply(surplus, mkCoin(pumpDenom, 100_000000))
	s.supply(surplus, mkCoin(dumpDenom, 100_000000))

	// this will be a DUMP (historic price 1.00, current price 0.50) borrower
	// using PUMP (historic price 1.00, current price 2.00) collateral
	dumpborrower := s.newAccount(mkCoin(pumpDenom, 100_000000))
	s.supply(dumpborrower, mkCoin(pumpDenom, 100_000000))
	s.collateralize(dumpborrower, mkCoin("u/"+pumpDenom, 100_000000))
	// collateral value is $200 (current) or $100 (historic)
	// collateral weights are always 0.25 in testing

	// this will be a PUMP (historic price 1.00, current price 2.00) borrower
	// using DUMP (historic price 1.00, current price 0.50) collateral
	pumpborrower := s.newAccount(mkCoin(dumpDenom, 100_000000))
	s.supply(pumpborrower, mkCoin(dumpDenom, 100_000000))
	s.collateralize(pumpborrower, mkCoin("u/"+dumpDenom, 100_000000))
	// collateral value is $50 (current) or $100 (historic)
	// collateral weights are always 0.25 in testing

	tcs := []testCase{
		{
			"uToken",
			borrower,
			mkCoin("u/"+umeeDenom, 100_000000),
			types.ErrUToken,
		}, {
			"unregistered token",
			borrower,
			mkCoin("abcd", 100_000000),
			types.ErrNotRegisteredToken,
		}, {
			"lending pool insufficient",
			borrower,
			mkCoin(umeeDenom, 200_000000),
			types.ErrLendingPoolInsufficient,
		}, {
			"valid borrow",
			borrower,
			mkCoin(umeeDenom, 70_000000),
			nil,
		}, {
			"additional borrow",
			borrower,
			mkCoin(umeeDenom, 20_000000),
			nil,
		}, {
			"max supply utilization",
			borrower,
			mkCoin(umeeDenom, 10_000000),
			types.ErrMaxSupplyUtilization,
		}, {
			"atom borrow",
			borrower,
			mkCoin(atomDenom, 1_000000),
			nil,
		}, {
			"borrow limit",
			borrower,
			mkCoin(atomDenom, 100_000000),
			types.ErrUndercollaterized,
		}, {
			"zero collateral",
			supplier,
			mkCoin(atomDenom, 1_000000),
			types.ErrUndercollaterized,
		}, {
			"dump borrower (acceptable)",
			dumpborrower,
			mkCoin(dumpDenom, 20_000000),
			nil,
		}, {
			"dump borrower (borrow limit)",
			dumpborrower,
			mkCoin(dumpDenom, 10_000000),
			types.ErrUndercollaterized,
		}, {
			"pump borrower (acceptable)",
			pumpborrower,
			mkCoin(pumpDenom, 5_000000),
			nil,
		}, {
			"pump borrower (borrow limit)",
			pumpborrower,
			mkCoin(pumpDenom, 2_000000),
			types.ErrUndercollaterized,
		},
	}

	for _, tc := range tcs {
		msg := &types.MsgBorrow{
			Borrower: tc.addr.String(),
			Asset:    tc.coin,
		}
		if tc.err != nil {
			_, err := srv.Borrow(ctx, msg)
			require.ErrorIs(err, tc.err, tc.msg)
		} else {
			// initial state
			iBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, tc.coin.Denom)
			iBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify the output of borrow function
			resp, err := srv.Borrow(ctx, msg)
			require.NoError(err, tc.msg)
			require.Equal(&types.MsgBorrowResponse{}, resp, tc.msg)

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

func (s *IntegrationTestSuite) TestMsgMaxBorrow() {
	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// create and fund a supplier which supplies 100 UMEE and 100 ATOM
	supplier := s.newAccount(mkCoin(umeeDenom, 100_000000), mkCoin(atomDenom, 100_000000))
	s.supply(supplier, mkCoin(umeeDenom, 100_000000), mkCoin(atomDenom, 100_000000))

	// create a borrower which supplies and collateralizes 100 ATOM
	borrower := s.newAccount(mkCoin(atomDenom, 100_000000))
	s.supply(borrower, mkCoin(atomDenom, 100_000000))
	s.collateralize(borrower, mkCoin("u/"+atomDenom, 100_000000))

	// create an additional supplier (DUMP, PUMP tokens)
	surplus := s.newAccount(mkCoin(dumpDenom, 100_000000), mkCoin(pumpDenom, 100_000000))
	s.supply(surplus, mkCoin(pumpDenom, 100_000000))
	s.supply(surplus, mkCoin(dumpDenom, 100_000000))

	// this will be a DUMP (historic price 1.00, current price 0.50) borrower
	// using PUMP (historic price 1.00, current price 2.00) collateral
	dumpborrower := s.newAccount(mkCoin(pumpDenom, 100_000000))
	s.supply(dumpborrower, mkCoin(pumpDenom, 100_000000))
	s.collateralize(dumpborrower, mkCoin("u/"+pumpDenom, 100_000000))
	// collateral value is $200 (current) or $100 (historic)
	// collateral weights are always 0.25 in testing

	// this will be a PUMP (historic price 1.00, current price 2.00) borrower
	// using DUMP (historic price 1.00, current price 0.50) collateral
	pumpborrower := s.newAccount(mkCoin(dumpDenom, 100_000000))
	s.supply(pumpborrower, mkCoin(dumpDenom, 100_000000))
	s.collateralize(pumpborrower, mkCoin("u/"+dumpDenom, 100_000000))
	// collateral value is $50 (current) or $100 (historic)
	// collateral weights are always 0.25 in testing

	tcs := []struct {
		msg  string
		addr sdk.AccAddress
		coin sdk.Coin
		err  error
	}{
		{
			"uToken",
			borrower,
			mkCoin("u/"+umeeDenom, 0),
			types.ErrUToken,
		}, {
			"unregistered token",
			borrower,
			mkCoin("abcd", 0),
			types.ErrNotRegisteredToken,
		}, {
			"zero collateral - should return zero",
			supplier,
			mkCoin(atomDenom, 0),
			nil,
		}, {
			"atom borrow",
			borrower,
			mkCoin(atomDenom, 25_000000),
			nil,
		}, {
			"already borrowed max - should return zero",
			borrower,
			mkCoin(atomDenom, 0),
			nil,
		}, {
			"dump borrower",
			dumpborrower,
			mkCoin(dumpDenom, 25_000000),
			nil,
		}, {
			"pump borrower",
			pumpborrower,
			mkCoin(pumpDenom, 6_250000),
			nil,
		},
	}

	for _, tc := range tcs {
		msg := &types.MsgMaxBorrow{
			Borrower: tc.addr.String(),
			Denom:    tc.coin.Denom,
		}
		if tc.err != nil {
			_, err := srv.MaxBorrow(ctx, msg)
			require.ErrorIs(err, tc.err, tc.msg)
			continue
		}
		// initial state
		iBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
		iCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
		iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
		iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, tc.coin.Denom)
		iBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

		// verify the output of borrow function
		resp, err := srv.MaxBorrow(ctx, msg)
		require.NoError(err, tc.msg)
		require.Equal(&types.MsgMaxBorrowResponse{Borrowed: tc.coin}, resp, tc.msg)

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
		require.Equal(coin.NormalizeCoins(iBorrowed.Add(tc.coin)), fBorrowed, tc.msg, "borrowed coins")

		// check all available invariants
		s.checkInvariants(tc.msg)
	}
}

func (s *IntegrationTestSuite) TestMsgRepay() {
	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// create and fund a borrower which supplies and collateralizes UMEE, then borrows 10 UMEE
	borrower := s.newAccount(mkCoin(umeeDenom, 200_000000))
	s.supply(borrower, mkCoin(umeeDenom, 150_000000))
	s.collateralize(borrower, mkCoin("u/"+umeeDenom, 120_000000))
	s.borrow(borrower, mkCoin(umeeDenom, 10_000000))

	// create and fund a borrower which engages in a supply->borrow->supply loop
	looper := s.newAccount(mkCoin(umeeDenom, 50_000000))
	s.supply(looper, mkCoin(umeeDenom, 50_000000))
	s.collateralize(looper, mkCoin("u/"+umeeDenom, 50_000000))
	s.borrow(looper, mkCoin(umeeDenom, 5_000000))
	s.supply(looper, mkCoin(umeeDenom, 5_000000))

	tcs := []struct {
		msg           string
		addr          sdk.AccAddress
		coin          sdk.Coin
		expectedRepay sdk.Coin
		err           error
	}{
		{
			"should not accept uToken repay",
			borrower,
			mkCoin("u/"+umeeDenom, 100_000000),
			sdk.Coin{},
			types.ErrUToken,
		}, {
			"unregistered token",
			borrower,
			mkCoin("abcd", 100_000000),
			coin.ZeroCoin("abcd"),
			nil,
		}, {
			"not borrowed",
			borrower,
			mkCoin(atomDenom, 100_000000),
			coin.ZeroCoin(atomDenom),
			nil,
		}, {
			"valid repay",
			borrower,
			mkCoin(umeeDenom, 1_000000),
			mkCoin(umeeDenom, 1_000000),
			nil,
		}, {
			"additional repay",
			borrower,
			mkCoin(umeeDenom, 3_000000),
			mkCoin(umeeDenom, 3_000000),
			nil,
		}, {
			"overpay",
			borrower,
			mkCoin(umeeDenom, 30_000000),
			mkCoin(umeeDenom, 6_000000),
			nil,
		}, {
			"insufficient balance",
			looper,
			mkCoin(umeeDenom, 1_000000),
			sdk.Coin{},
			sdkerrors.ErrInsufficientFunds,
		},
	}

	for _, tc := range tcs {
		msg := &types.MsgRepay{
			Borrower: tc.addr.String(),
			Asset:    tc.coin,
		}
		if tc.err != nil {
			_, err := srv.Repay(ctx, msg)
			require.ErrorIs(err, tc.err, tc.msg)
		} else {
			// initial state
			iBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iCollateral := app.LeverageKeeper.GetBorrowerCollateral(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iExchangeRate := app.LeverageKeeper.DeriveExchangeRate(ctx, tc.coin.Denom)
			iBorrowed := app.LeverageKeeper.GetBorrowerBorrows(ctx, tc.addr)

			// verify the output of repay function
			resp, err := srv.Repay(ctx, msg)
			require.NoError(err, tc.msg)
			require.Equal(tc.expectedRepay, resp.Repaid, tc.msg)

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

func (s *IntegrationTestSuite) TestMsgLiquidate() {
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

	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// create and fund a liquidator which supplies plenty of UMEE and ATOM to the module
	supplier := s.newAccount(mkCoin(umeeDenom, 1000_000000), mkCoin(atomDenom, 1000_000000))
	s.supply(supplier, mkCoin(umeeDenom, 1000_000000), mkCoin(atomDenom, 1000_000000))

	// create and fund a liquidator which has 1000 UMEE and 1000 ATOM
	liquidator := s.newAccount(mkCoin(umeeDenom, 1000_000000), mkCoin(atomDenom, 1000_000000))

	// create a healthy borrower
	healthyBorrower := s.newAccount(mkCoin(umeeDenom, 100_000000))
	s.supply(healthyBorrower, mkCoin(umeeDenom, 100_000000))
	s.collateralize(healthyBorrower, mkCoin("u/"+umeeDenom, 100_000000))
	s.borrow(healthyBorrower, mkCoin(umeeDenom, 10_000000))

	// create a borrower which supplies and collateralizes 1000 ATOM
	atomBorrower := s.newAccount(mkCoin(atomDenom, 1000_000000))
	s.supply(atomBorrower, mkCoin(atomDenom, 1000_000000))
	s.collateralize(atomBorrower, mkCoin("u/"+atomDenom, 1000_000000))
	// artificially borrow 500 ATOM - this can be liquidated without bad debt
	s.forceBorrow(atomBorrower, mkCoin(atomDenom, 500_000000))

	// create a borrower which collateralizes 110 UMEE
	umeeBorrower := s.newAccount(mkCoin(umeeDenom, 300_000000))
	s.supply(umeeBorrower, mkCoin(umeeDenom, 200_000000))
	s.collateralize(umeeBorrower, mkCoin("u/"+umeeDenom, 110_000000))
	// artificially borrow 200 UMEE - this will create a bad debt when liquidated
	s.forceBorrow(umeeBorrower, mkCoin(umeeDenom, 200_000000))

	// creates a complex borrower with multiple denoms active
	complexBorrower := s.newAccount(mkCoin(umeeDenom, 100_000000), mkCoin(atomDenom, 100_000000))
	s.supply(complexBorrower, mkCoin(umeeDenom, 100_000000), mkCoin(atomDenom, 100_000000))
	s.collateralize(complexBorrower, mkCoin("u/"+umeeDenom, 100_000000), mkCoin("u/"+atomDenom, 100_000000))
	// artificially borrow multiple denoms
	s.forceBorrow(complexBorrower, mkCoin(atomDenom, 30_000000), mkCoin(umeeDenom, 30_000000))

	// creates a realistic borrower with 400 UMEE collateral which will have a close factor < 1
	closeBorrower := s.newAccount(mkCoin(umeeDenom, 400_000000))
	s.supply(closeBorrower, mkCoin(umeeDenom, 400_000000))
	s.collateralize(closeBorrower, mkCoin("u/"+umeeDenom, 400_000000))
	// artificially borrow just barely above liquidation threshold to simulate interest accruing
	s.forceBorrow(closeBorrower, mkCoin(umeeDenom, 102_000000))

	tcs := []testCase{
		{
			"healthy borrower",
			liquidator,
			healthyBorrower,
			mkCoin(atomDenom, 1_000000),
			atomDenom,
			sdk.Coin{},
			sdk.Coin{},
			sdk.Coin{},
			types.ErrLiquidationIneligible,
		}, {
			"not borrowed denom",
			liquidator,
			umeeBorrower,
			mkCoin(atomDenom, 1_000000),
			atomDenom,
			sdk.Coin{},
			sdk.Coin{},
			sdk.Coin{},
			types.ErrLiquidationRepayZero,
		}, {
			"direct atom liquidation",
			liquidator,
			atomBorrower,
			mkCoin(atomDenom, 100_000000),
			atomDenom,
			mkCoin(atomDenom, 100_000000),
			mkCoin("u/"+atomDenom, 109_000000),
			mkCoin(atomDenom, 109_000000),
			nil,
		}, {
			"u/atom liquidation",
			liquidator,
			atomBorrower,
			mkCoin(atomDenom, 100_000000),
			"u/" + atomDenom,
			mkCoin(atomDenom, 100_000000),
			mkCoin("u/"+atomDenom, 110_000000),
			mkCoin("u/"+atomDenom, 110_000000),
			nil,
		}, {
			"complete u/atom liquidation",
			liquidator,
			atomBorrower,
			mkCoin(atomDenom, 500_000000),
			"u/" + atomDenom,
			mkCoin(atomDenom, 300_000000),
			mkCoin("u/"+atomDenom, 330_000000),
			mkCoin("u/"+atomDenom, 330_000000),
			nil,
		}, {
			"bad debt u/umee liquidation",
			liquidator,
			umeeBorrower,
			mkCoin(umeeDenom, 200_000000),
			"u/" + umeeDenom,
			mkCoin(umeeDenom, 100_000000),
			mkCoin("u/"+umeeDenom, 110_000000),
			mkCoin("u/"+umeeDenom, 110_000000),
			nil,
		}, {
			"complex borrower",
			liquidator,
			complexBorrower,
			mkCoin(umeeDenom, 200_000000),
			"u/" + atomDenom,
			mkCoin(umeeDenom, 30_000000),
			mkCoin("u/"+atomDenom, 3_527932),
			mkCoin("u/"+atomDenom, 3_527932),
			nil,
		}, {
			"close factor < 1",
			liquidator,
			closeBorrower,
			mkCoin(umeeDenom, 200_000000),
			"u/" + umeeDenom,
			mkCoin(umeeDenom, 7_752000),
			mkCoin("u/"+umeeDenom, 8_527200),
			mkCoin("u/"+umeeDenom, 8_527200),
			nil,
		},
	}

	for _, tc := range tcs {
		msg := &types.MsgLiquidate{
			Liquidator:  tc.liquidator.String(),
			Borrower:    tc.borrower.String(),
			Repayment:   tc.attemptedRepay,
			RewardDenom: tc.rewardDenom,
		}
		if tc.err != nil {
			_, err := srv.Liquidate(ctx, msg)
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
			resp, err := srv.Liquidate(ctx, msg)
			require.NoError(err, tc.msg)
			require.Equal(tc.expectedRepay, resp.Repaid, tc.msg, "repaid")
			require.Equal(tc.expectedLiquidate, resp.Collateral, tc.msg, "liquidated")
			require.Equal(tc.expectedReward, resp.Reward, tc.msg, "reward")

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
	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// update initial ATOM to have a limited MaxCollateralShare
	atom, err := app.LeverageKeeper.GetTokenSettings(ctx, atomDenom)
	require.NoError(err)
	atom.MaxCollateralShare = sdk.MustNewDecFromStr("0.1")
	s.registerToken(atom)

	// Mock oracle prices:
	// UMEE $4.21
	// ATOM $39.38

	// create a supplier to collateralize 100 UMEE, worth $421.00
	umeeSupplier := s.newAccount(mkCoin(umeeDenom, 100_000000))
	s.supply(umeeSupplier, mkCoin(umeeDenom, 100_000000))
	s.collateralize(umeeSupplier, mkCoin("u/"+umeeDenom, 100_000000))

	// create an ATOM supplier
	atomSupplier := s.newAccount(mkCoin(atomDenom, 100_000000))
	s.supply(atomSupplier, mkCoin(atomDenom, 100_000000))

	// collateralize 1.18 ATOM, worth $46.46, with no error.
	// total collateral value (across all denoms) will be $467.46
	// so ATOM's collateral share ($46.46 / $467.46) is barely below 10%
	s.collateralize(atomSupplier, mkCoin("u/"+atomDenom, 1_180000))

	// attempt to collateralize another 0.01 ATOM, which would result in too much collateral share for ATOM
	msg := &types.MsgCollateralize{
		Borrower: atomSupplier.String(),
		Asset:    mkCoin("u/"+atomDenom, 10000),
	}
	_, err = srv.Collateralize(ctx, msg)
	require.ErrorIs(err, types.ErrMaxCollateralShare)
}

func (s *IntegrationTestSuite) TestMinCollateralLiquidity() {
	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// update initial UMEE to have a limited MinCollateralLiquidity
	umee, err := app.LeverageKeeper.GetTokenSettings(ctx, umeeDenom)
	require.NoError(err)
	umee.MinCollateralLiquidity = sdk.MustNewDecFromStr("0.5")
	s.registerToken(umee)

	// create a supplier to collateralize 100 UMEE
	umeeSupplier := s.newAccount(mkCoin(umeeDenom, 100_000000))
	s.supply(umeeSupplier, mkCoin(umeeDenom, 100_000000))
	s.collateralize(umeeSupplier, mkCoin("u/"+umeeDenom, 100_000000))

	// create an ATOM supplier and borrow 49 UMEE
	atomSupplier := s.newAccount(mkCoin(atomDenom, 100_000000))
	s.supply(atomSupplier, mkCoin(atomDenom, 100_000000))
	s.collateralize(atomSupplier, mkCoin("u/"+atomDenom, 100_000000))
	s.borrow(atomSupplier, mkCoin(umeeDenom, 49_000000))

	// collateral liquidity (liquidity / collateral) of UMEE is 51/100

	// withdrawal would reduce collateral liquidity to 41/90
	msg1 := &types.MsgWithdraw{
		Supplier: umeeSupplier.String(),
		Asset:    mkCoin("u/"+umeeDenom, 10_000000),
	}
	_, err = srv.Withdraw(ctx, msg1)
	require.ErrorIs(err, types.ErrMinCollateralLiquidity, "withdraw")

	// borrow would reduce collateral liquidity to 41/100
	msg2 := &types.MsgBorrow{
		Borrower: umeeSupplier.String(),
		Asset:    mkCoin(umeeDenom, 10_000000),
	}
	_, err = srv.Borrow(ctx, msg2)
	require.ErrorIs(err, types.ErrMinCollateralLiquidity, "borrow")
}

func (s *IntegrationTestSuite) TestMinCollateralLiquidity_Collateralize() {
	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// update initial UMEE to have a limited MinCollateralLiquidity
	umee, err := app.LeverageKeeper.GetTokenSettings(ctx, umeeDenom)
	require.NoError(err)
	umee.MinCollateralLiquidity = sdk.MustNewDecFromStr("0.5")
	s.registerToken(umee)

	// create a supplier to supply 200 UMEE, and collateralize 100 UMEE
	umeeSupplier := s.newAccount(mkCoin(umeeDenom, 200))
	s.supply(umeeSupplier, mkCoin(umeeDenom, 200))
	s.collateralize(umeeSupplier, mkCoin("u/"+umeeDenom, 100))

	// create an ATOM supplier and borrow 149 UMEE
	atomSupplier := s.newAccount(mkCoin(atomDenom, 100))
	s.supply(atomSupplier, mkCoin(atomDenom, 100))
	s.collateralize(atomSupplier, mkCoin("u/"+atomDenom, 100))
	s.borrow(atomSupplier, mkCoin(umeeDenom, 149))

	// collateral liquidity (liquidity / collateral) of UMEE is 51/100

	// collateralize would reduce collateral liquidity to 51/200
	msg := &types.MsgCollateralize{
		Borrower: umeeSupplier.String(),
		Asset:    mkCoin("u/"+umeeDenom, 100),
	}
	_, err = srv.Collateralize(ctx, msg)
	require.ErrorIs(err, types.ErrMinCollateralLiquidity, "collateralize")
}
