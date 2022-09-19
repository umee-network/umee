package keeper_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v3/x/leverage/fixtures"
	"github.com/umee-network/umee/v3/x/leverage/keeper"
	"github.com/umee-network/umee/v3/x/leverage/types"
)

func (s *IntegrationTestSuite) TestQuerier_RegisteredTokens() {
	ctx, require := s.ctx, s.Require()

	resp, err := s.queryClient.RegisteredTokens(ctx.Context(), &types.QueryRegisteredTokens{})
	require.NoError(err)
	require.Len(resp.Registry, 2, "token registry length")
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

	oraclePrice := sdk.MustNewDecFromStr("0.00000421")

	expected := types.QueryMarketSummaryResponse{
		SymbolDenom:            "UMEE",
		Exponent:               6,
		OraclePrice:            &oraclePrice,
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
	addr := s.newAccount(coin(umeeDenom, 1000))
	s.supply(addr, coin(umeeDenom, 1000))
	s.collateralize(addr, coin("u/"+umeeDenom, 1000))

	resp, err := s.queryClient.AccountBalances(ctx.Context(), &types.QueryAccountBalances{Address: addr.String()})
	require.NoError(err)

	expected := types.QueryAccountBalancesResponse{
		Supplied: sdk.NewCoins(
			coin(umeeDenom, 1000),
		),
		Collateral: sdk.NewCoins(
			coin("u/"+umeeDenom, 1000),
		),
		Borrowed: nil,
	}

	require.Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_AccountSummary() {
	ctx, require := s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(coin(umeeDenom, 1000_000000))
	s.supply(addr, coin(umeeDenom, 1000_000000))
	s.collateralize(addr, coin("u/"+umeeDenom, 1000_000000))

	resp, err := s.queryClient.AccountSummary(ctx.Context(), &types.QueryAccountSummary{Address: addr.String()})
	require.NoError(err)

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
		LiquidationThreshold: sdk.MustNewDecFromStr("1052.5"),
	}

	require.Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_LiquidationTargets() {
	ctx, require := s.ctx, s.Require()

	keeper.EnableLiquidator = "false"

	_, err := s.queryClient.LiquidationTargets(ctx.Context(), &types.QueryLiquidationTargets{})
	require.ErrorIs(err, types.ErrNotLiquidatorNode)

	keeper.EnableLiquidator = "true"

	resp, err := s.queryClient.LiquidationTargets(ctx.Context(), &types.QueryLiquidationTargets{})
	require.NoError(err)

	expected := types.QueryLiquidationTargetsResponse{
		Targets: nil,
	}

	require.Equal(expected, *resp)
}
