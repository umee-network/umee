package keeper_test

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/leverage/fixtures"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

func (s *IntegrationTestSuite) TestQuerier_RegisteredTokens() {
	ctx := s.ctx

	tests := []struct {
		name           string
		errMsg         string
		req            types.QueryRegisteredTokens
		expectedTokens int
	}{
		{
			"valid: get the all registered tokens",
			"",
			types.QueryRegisteredTokens{},
			leverage_initial_registry_length,
		},
		{
			"valid: get the registered token info by base_denom",
			"",
			types.QueryRegisteredTokens{BaseDenom: appparams.BondDenom},
			1,
		},
		{
			"invalid: get the not registered token info by base_denom",
			"not a registered Token",
			types.QueryRegisteredTokens{BaseDenom: "not_reg_token"},
			0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := s.queryClient.RegisteredTokens(ctx, &tc.req)
			if tc.errMsg == "" {
				assert.NilError(s.T(), err)
				assert.Equal(s.T(), tc.expectedTokens, len(resp.Registry))
			} else {
				assert.ErrorContains(s.T(), err, "not a registered Token")
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQuerier_Params() {
	ctx, require := s.ctx, s.Require()

	resp, err := s.queryClient.Params(ctx, &types.QueryParams{})
	require.NoError(err)
	require.Equal(fixtures.Params(), resp.Params)
}

func (s *IntegrationTestSuite) TestQuerier_MarketSummary() {
	require := s.Require()

	req := &types.QueryMarketSummary{}
	_, err := s.queryClient.MarketSummary(context.Background(), req)
	require.ErrorContains(err, "empty denom")

	req = &types.QueryMarketSummary{Denom: "uumee"}
	resp, err := s.queryClient.MarketSummary(context.Background(), req)
	require.NoError(err)

	oracleSymbolPrice := sdk.MustNewDecFromStr("4.21")

	expected := types.QueryMarketSummaryResponse{
		SymbolDenom:         "UMEE",
		Exponent:            6,
		OraclePrice:         &oracleSymbolPrice,
		OracleHistoricPrice: &oracleSymbolPrice,
		UTokenExchangeRate:  sdk.OneDec(),
		// see cli/tests "query market summary - zero supply"
		Supply_APY:             sdk.MustNewDecFromStr("1.1704"),
		Borrow_APY:             sdk.MustNewDecFromStr("1.52"),
		Supplied:               sdk.ZeroInt(),
		Reserved:               sdk.ZeroInt(),
		Collateral:             sdk.ZeroInt(),
		Borrowed:               sdk.ZeroInt(),
		Liquidity:              sdk.ZeroInt(),
		MaximumBorrow:          sdk.ZeroInt(),
		MaximumCollateral:      sdk.ZeroInt(),
		MinimumLiquidity:       sdk.ZeroInt(),
		UTokenSupply:           sdk.ZeroInt(),
		AvailableBorrow:        sdk.ZeroInt(),
		AvailableWithdraw:      sdk.ZeroInt(),
		AvailableCollateralize: sdk.ZeroInt(),
	}
	require.Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_TokenMarkets() {
	require := s.Require()

	req := &types.QueryRegisteredTokensWithMarkets{}
	resp, err := s.queryClient.RegisteredTokensWithMarkets(context.Background(), req)
	require.NoError(err)

	expected := types.QueryRegisteredTokensWithMarketsResponse{
		Markets: []types.TokenMarket{},
	}
	tokens := s.tk.GetAllRegisteredTokens(s.ctx)
	for _, token := range tokens {
		ms, err := s.queryClient.MarketSummary(context.Background(), &types.QueryMarketSummary{Denom: token.BaseDenom})
		require.NoError(err)
		expected.Markets = append(expected.Markets, types.TokenMarket{
			Token:  token,
			Market: *ms,
		})
	}

	require.Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_AccountBalances() {
	ctx, require := s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 uumee
	addr := s.newAccount(coin.New(umeeDenom, 1000))
	s.supply(addr, coin.New(umeeDenom, 1000))
	s.collateralize(addr, coin.New("u/"+umeeDenom, 1000))

	resp, err := s.queryClient.AccountBalances(ctx, &types.QueryAccountBalances{Address: addr.String()})
	require.NoError(err)

	expected := types.QueryAccountBalancesResponse{
		Supplied: sdk.NewCoins(
			coin.New(umeeDenom, 1000),
		),
		Collateral: sdk.NewCoins(
			coin.New("u/"+umeeDenom, 1000),
		),
		Borrowed: nil,
	}

	require.Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_AccountSummary() {
	ctx, require := s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(coin.New(umeeDenom, 1000_000000))
	s.supply(addr, coin.New(umeeDenom, 1000_000000))
	s.collateralize(addr, coin.New("u/"+umeeDenom, 1000_000000))

	resp, err := s.queryClient.AccountSummary(ctx, &types.QueryAccountSummary{Address: addr.String()})
	require.NoError(err)

	lt := sdk.MustNewDecFromStr("1094.6")
	bl := sdk.MustNewDecFromStr("1052.5")
	expected := types.QueryAccountSummaryResponse{
		// This result is umee's oracle exchange rate from
		// from .Reset() in x/leverage/keeper/oracle_test.go
		// times the amount of umee, then sometimes times params
		// from newToken in x/leverage/keeper/keeper_test.go
		// (1000) * 4.21 = 4210
		SuppliedValue:     sdk.MustNewDecFromStr("4210"),
		SpotSuppliedValue: sdk.MustNewDecFromStr("4210"),
		// (1000) * 4.21 = 4210
		CollateralValue:     sdk.MustNewDecFromStr("4210"),
		SpotCollateralValue: sdk.MustNewDecFromStr("4210"),
		// Nothing borrowed
		BorrowedValue:     sdk.ZeroDec(),
		SpotBorrowedValue: sdk.ZeroDec(),
		// (1000) * 4.21 * 0.25 = 1052.5
		BorrowLimit: &bl,
		// (1000) * 4.21 * 0.26 = 1094.6
		LiquidationThreshold: &lt,
	}
	require.Equal(expected, *resp)

	// creates account which has supplied and collateralized 1000 OUTAGE and 500 PAIRED (these are $1, 6 exponent tokens)
	addr = s.newAccount(coin.New(outageDenom, 1000_000000), coin.New(pairedDenom, 500_000000))
	s.supply(addr, coin.New(outageDenom, 1000_000000), coin.New(pairedDenom, 500_000000))
	s.collateralize(addr, coin.Utoken(outageDenom, 1000_000000), coin.Utoken(pairedDenom, 500_000000))

	resp, err = s.queryClient.AccountSummary(ctx, &types.QueryAccountSummary{Address: addr.String()})
	require.NoError(err)
	bl = sdk.MustNewDecFromStr("125")
	expected = types.QueryAccountSummaryResponse{
		// Price outage should have no effect on value fields
		SuppliedValue:       sdk.MustNewDecFromStr("1500"),
		SpotSuppliedValue:   sdk.MustNewDecFromStr("1500"),
		CollateralValue:     sdk.MustNewDecFromStr("1500"),
		SpotCollateralValue: sdk.MustNewDecFromStr("1500"),
		// Nothing borrowed
		BorrowedValue:        sdk.ZeroDec(),
		SpotBorrowedValue:    sdk.ZeroDec(),
		BorrowLimit:          &bl,
		LiquidationThreshold: nil, // missing collateral price: no threshold can be displayed
	}
	require.Equal(expected, *resp)

	// creates account which has supplied and collateralized 1000 OUTAGE and 500 PAIRED (these are $1, 6 exponent tokens)
	addr = s.newAccount(coin.New(outageDenom, 1000_000000), coin.New(pairedDenom, 500_000000))
	s.supply(addr, coin.New(outageDenom, 1000_000000), coin.New(pairedDenom, 500_000000))
	s.collateralize(addr, coin.Utoken(outageDenom, 1000_000000), coin.Utoken(pairedDenom, 500_000000))
	// also borrow some PAIRED normally
	s.borrow(addr, coin.New(pairedDenom, 100_000000))
	// and force-borrow (cannot normally because due to missing price) some OUTAGE
	s.forceBorrow(addr, coin.New(outageDenom, 200_000000))

	resp, err = s.queryClient.AccountSummary(ctx, &types.QueryAccountSummary{Address: addr.String()})
	require.NoError(err)
	expected = types.QueryAccountSummaryResponse{
		// Both prices should show up in spot fields and query fields.
		SuppliedValue:       sdk.MustNewDecFromStr("1500"),
		SpotSuppliedValue:   sdk.MustNewDecFromStr("1500"),
		CollateralValue:     sdk.MustNewDecFromStr("1500"),
		SpotCollateralValue: sdk.MustNewDecFromStr("1500"),
		// Borrowed 1/5 of collateral values
		BorrowedValue:        sdk.MustNewDecFromStr("300"),
		SpotBorrowedValue:    sdk.MustNewDecFromStr("300"),
		BorrowLimit:          nil, // missing borrow price: no borrow limit can be displayed
		LiquidationThreshold: nil, // missing collateral price: no threshold can be displayed
	}
	require.Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_Inspect() {
	ctx, require := s.ctx, s.Require()

	// creates accounts
	addr1 := s.newAccount(coin.New(umeeDenom, 1000_000000))
	s.supply(addr1, coin.New(umeeDenom, 1000_000000))
	s.collateralize(addr1, coin.New("u/"+umeeDenom, 1000_000000))
	s.borrow(addr1, coin.New(umeeDenom, 10_500000))
	addr2 := s.newAccount(coin.New(umeeDenom, 1000_000000))
	s.supply(addr2, coin.New(umeeDenom, 1000_000000))
	s.collateralize(addr2, coin.New("u/"+umeeDenom, 60_000000))
	s.borrow(addr2, coin.New(umeeDenom, 1_500000))
	addr3 := s.newAccount(coin.New(umeeDenom, 1000_000000))
	s.supply(addr3, coin.New(umeeDenom, 1000_000000))
	s.collateralize(addr3, coin.New("u/"+umeeDenom, 600_000000))
	s.borrow(addr3, coin.New(umeeDenom, 15_000000))

	resp, err := s.queryClient.Inspect(ctx, &types.QueryInspect{})
	require.NoError(err)

	convertToPositionBalances := func(c sdk.DecCoins, baseAmount sdkmath.Int) []types.PositionBalance {
		res := make([]types.PositionBalance, 0)
		for _, c := range c {
			res = append(res, types.PositionBalance{
				Amount:     c.Amount,
				Denom:      c.Denom,
				BaseDenom:  umeeDenom,
				BaseAmount: baseAmount,
			})
		}
		return res
	}

	expected := types.QueryInspectResponse{
		Borrowers: []types.InspectAccount{
			{
				Address: addr3.String(),
				Analysis: &types.RiskInfo{
					Borrowed:    63.15,
					Liquidation: 656,
					Value:       2526,
				},
				Position: &types.DecBalances{
					Collateral: convertToPositionBalances(sdk.NewDecCoins(coin.Dec("UMEE", "600")), sdkmath.NewInt(600_000000)),
					Borrowed:   convertToPositionBalances(sdk.NewDecCoins(coin.Dec("UMEE", "15")), sdkmath.NewInt(15_000000)),
				},
			},
			{
				Address: addr1.String(),
				Analysis: &types.RiskInfo{
					Borrowed:    44.2,
					Liquidation: 1094,
					Value:       4210,
				},
				Position: &types.DecBalances{
					Collateral: convertToPositionBalances(sdk.NewDecCoins(coin.Dec("UMEE", "1000")), sdkmath.NewInt(1000_000000)),
					Borrowed:   convertToPositionBalances(sdk.NewDecCoins(coin.Dec("UMEE", "10.5")), sdkmath.NewInt(10_500000)),
				},
			},
			{
				Address: addr2.String(),
				Analysis: &types.RiskInfo{
					Borrowed:    6.31,
					Liquidation: 65.67,
					Value:       252,
				},
				Position: &types.DecBalances{
					Collateral: convertToPositionBalances(sdk.NewDecCoins(coin.Dec("UMEE", "60")), sdkmath.NewInt(60_000000)),
					Borrowed:   convertToPositionBalances(sdk.NewDecCoins(coin.Dec("UMEE", "1.5")), sdkmath.NewInt(1_500000)),
				},
			},
		},
	}
	require.Equal(expected, *resp)

	req := &types.QueryInspect{}
	req.Symbol = "UMEE"
	resp, err = s.queryClient.Inspect(ctx, req)
	require.NoError(err)
	require.Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_LiquidationTargets() {
	ctx, require := s.ctx, s.Require()

	resp, err := s.queryClient.LiquidationTargets(ctx, &types.QueryLiquidationTargets{})
	require.NoError(err)

	expected := types.QueryLiquidationTargetsResponse{
		Targets: nil,
	}

	require.Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_BadDebts() {
	ctx, require := s.ctx, s.Require()

	resp, err := s.queryClient.BadDebts(ctx, &types.QueryBadDebts{})
	require.NoError(err)

	expected := types.QueryBadDebtsResponse{
		Targets: nil,
	}

	require.Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_MaxWithdraw() {
	ctx, require := s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(coin.New(umeeDenom, 1000_000000))
	s.supply(addr, coin.New(umeeDenom, 1000_000000))
	s.collateralize(addr, coin.New("u/"+umeeDenom, 1000_000000))

	resp, err := s.queryClient.MaxWithdraw(ctx, &types.QueryMaxWithdraw{
		Address: addr.String(),
		Denom:   umeeDenom,
	})
	require.NoError(err)

	expected := types.QueryMaxWithdrawResponse{
		Tokens:  sdk.NewCoins(sdk.NewCoin(umeeDenom, sdk.NewInt(1000_000000))),
		UTokens: sdk.NewCoins(sdk.NewCoin("u/"+umeeDenom, sdk.NewInt(1000_000000))),
	}
	require.Equal(expected, *resp)

	// Also test all-denoms
	resp, err = s.queryClient.MaxWithdraw(ctx, &types.QueryMaxWithdraw{
		Address: addr.String(),
	})
	require.NoError(err)

	expected = types.QueryMaxWithdrawResponse{
		Tokens:  sdk.NewCoins(sdk.NewCoin(umeeDenom, sdk.NewInt(1000_000000))),
		UTokens: sdk.NewCoins(sdk.NewCoin("u/"+umeeDenom, sdk.NewInt(1000_000000))),
	}
	require.Equal(expected, *resp)

	// borrow 100 UMEE for non-trivial query
	s.borrow(addr, coin.New(umeeDenom, 100_000000))

	resp, err = s.queryClient.MaxWithdraw(ctx, &types.QueryMaxWithdraw{
		Address: addr.String(),
		Denom:   umeeDenom,
	})
	require.NoError(err)

	expected = types.QueryMaxWithdrawResponse{
		Tokens:  sdk.NewCoins(sdk.NewCoin(umeeDenom, sdk.NewInt(600_000000))),
		UTokens: sdk.NewCoins(sdk.NewCoin("u/"+umeeDenom, sdk.NewInt(600_000000))),
	}
	require.Equal(expected, *resp)

	// Also test all-denoms
	resp, err = s.queryClient.MaxWithdraw(ctx, &types.QueryMaxWithdraw{
		Address: addr.String(),
	})
	require.NoError(err)

	expected = types.QueryMaxWithdrawResponse{
		Tokens:  sdk.NewCoins(sdk.NewCoin(umeeDenom, sdk.NewInt(600_000000))),
		UTokens: sdk.NewCoins(sdk.NewCoin("u/"+umeeDenom, sdk.NewInt(600_000000))),
	}
	require.Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_MaxBorrow() {
	ctx, require := s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(coin.New(umeeDenom, 1000_000000))
	s.supply(addr, coin.New(umeeDenom, 1000_000000))
	s.collateralize(addr, coin.New("u/"+umeeDenom, 1000_000000))

	resp, err := s.queryClient.MaxBorrow(ctx, &types.QueryMaxBorrow{
		Address: addr.String(),
		Denom:   umeeDenom,
	})
	require.NoError(err)

	expected := types.QueryMaxBorrowResponse{
		Tokens: sdk.NewCoins(sdk.NewCoin(umeeDenom, sdk.NewInt(250_000000))),
	}
	require.Equal(expected, *resp)

	// Also test all-denoms
	resp, err = s.queryClient.MaxBorrow(ctx, &types.QueryMaxBorrow{
		Address: addr.String(),
	})
	require.NoError(err)

	expected = types.QueryMaxBorrowResponse{
		Tokens: sdk.NewCoins(sdk.NewCoin(umeeDenom, sdk.NewInt(250_000000))),
	}
	require.Equal(expected, *resp)

	// borrow 100 UMEE for non-trivial query
	s.borrow(addr, coin.New(umeeDenom, 100_000000))

	resp, err = s.queryClient.MaxBorrow(ctx, &types.QueryMaxBorrow{
		Address: addr.String(),
		Denom:   umeeDenom,
	})
	require.NoError(err)

	expected = types.QueryMaxBorrowResponse{
		Tokens: sdk.NewCoins(sdk.NewCoin(umeeDenom, sdk.NewInt(150_000000))),
	}
	require.Equal(expected, *resp)

	// Also test all-denoms
	resp, err = s.queryClient.MaxBorrow(ctx, &types.QueryMaxBorrow{
		Address: addr.String(),
	})
	require.NoError(err)

	expected = types.QueryMaxBorrowResponse{
		Tokens: sdk.NewCoins(sdk.NewCoin(umeeDenom, sdk.NewInt(150_000000))),
	}
	require.Equal(expected, *resp)
}
