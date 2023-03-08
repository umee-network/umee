package keeper_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/leverage/fixtures"
	"github.com/umee-network/umee/v4/x/leverage/types"
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
			5,
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
			resp, err := s.queryClient.RegisteredTokens(ctx.Context(), &tc.req)
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

	resp, err := s.queryClient.Params(ctx.Context(), &types.QueryParams{})
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
		SymbolDenom:            "UMEE",
		Exponent:               6,
		OraclePrice:            &oracleSymbolPrice,
		OracleHistoricPrice:    &oracleSymbolPrice,
		UTokenExchangeRate:     sdk.OneDec(),
		Supply_APY:             sdk.MustNewDecFromStr("1.2008"),
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

func (s *IntegrationTestSuite) TestQuerier_AccountBalances() {
	ctx, require := s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 uumee
	addr := s.newAccount(coin.New(umeeDenom, 1000))
	s.supply(addr, coin.New(umeeDenom, 1000))
	s.collateralize(addr, coin.New("u/"+umeeDenom, 1000))

	resp, err := s.queryClient.AccountBalances(ctx.Context(), &types.QueryAccountBalances{Address: addr.String()})
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

	resp, err := s.queryClient.AccountSummary(ctx.Context(), &types.QueryAccountSummary{Address: addr.String()})
	require.NoError(err)

	lt := sdk.MustNewDecFromStr("1052.5")
	expected := types.QueryAccountSummaryResponse{
		// This result is umee's oracle exchange rate from
		// from .Reset() in x/leverage/keeper/oracle_test.go
		// times the amount of umee, then sometimes times params
		// from newToken in x/leverage/keeper/keeper_test.go
		// (1000) * 4.21 = 4210
		SuppliedValue: sdk.MustNewDecFromStr("4210"),
		// (1000) * 4.21 = 4210
		CollateralValue: sdk.MustNewDecFromStr("4210"),
		// Nothing borrowed
		BorrowedValue: sdk.ZeroDec(),
		// (1000) * 4.21 * 0.25 = 1052.5
		BorrowLimit: sdk.MustNewDecFromStr("1052.5"),
		// (1000) * 4.21 * 0.25 = 1052.5
		LiquidationThreshold: &lt,
	}

	require.Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_LiquidationTargets() {
	ctx, require := s.ctx, s.Require()

	resp, err := s.queryClient.LiquidationTargets(ctx.Context(), &types.QueryLiquidationTargets{})
	require.NoError(err)

	expected := types.QueryLiquidationTargetsResponse{
		Targets: nil,
	}

	require.Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_BadDebts() {
	ctx, require := s.ctx, s.Require()

	resp, err := s.queryClient.BadDebts(ctx.Context(), &types.QueryBadDebts{})
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

	resp, err := s.queryClient.MaxWithdraw(ctx.Context(), &types.QueryMaxWithdraw{
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
	resp, err = s.queryClient.MaxWithdraw(ctx.Context(), &types.QueryMaxWithdraw{
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

	resp, err = s.queryClient.MaxWithdraw(ctx.Context(), &types.QueryMaxWithdraw{
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
	resp, err = s.queryClient.MaxWithdraw(ctx.Context(), &types.QueryMaxWithdraw{
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

	resp, err := s.queryClient.MaxBorrow(ctx.Context(), &types.QueryMaxBorrow{
		Address: addr.String(),
		Denom:   umeeDenom,
	})
	require.NoError(err)

	expected := types.QueryMaxBorrowResponse{
		Tokens: sdk.NewCoins(sdk.NewCoin(umeeDenom, sdk.NewInt(250_000000))),
	}
	require.Equal(expected, *resp)

	// Also test all-denoms
	resp, err = s.queryClient.MaxBorrow(ctx.Context(), &types.QueryMaxBorrow{
		Address: addr.String(),
	})
	require.NoError(err)

	expected = types.QueryMaxBorrowResponse{
		Tokens: sdk.NewCoins(sdk.NewCoin(umeeDenom, sdk.NewInt(250_000000))),
	}
	require.Equal(expected, *resp)

	// borrow 100 UMEE for non-trivial query
	s.borrow(addr, coin.New(umeeDenom, 100_000000))

	resp, err = s.queryClient.MaxBorrow(ctx.Context(), &types.QueryMaxBorrow{
		Address: addr.String(),
		Denom:   umeeDenom,
	})
	require.NoError(err)

	expected = types.QueryMaxBorrowResponse{
		Tokens: sdk.NewCoins(sdk.NewCoin(umeeDenom, sdk.NewInt(150_000000))),
	}
	require.Equal(expected, *resp)

	// Also test all-denoms
	resp, err = s.queryClient.MaxBorrow(ctx.Context(), &types.QueryMaxBorrow{
		Address: addr.String(),
	})
	require.NoError(err)

	expected = types.QueryMaxBorrowResponse{
		Tokens: sdk.NewCoins(sdk.NewCoin(umeeDenom, sdk.NewInt(150_000000))),
	}
	require.Equal(expected, *resp)
}
