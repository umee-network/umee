package keeper_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v3/x/leverage/types"
)

func (s *IntegrationTestSuite) TestQuerier_RegisteredTokens() {
	resp, err := s.queryClient.RegisteredTokens(s.ctx.Context(), &types.QueryRegisteredTokens{})
	s.Require().NoError(err)
	s.Require().Len(resp.Registry, 2, "token registry length")
}

func (s *IntegrationTestSuite) TestQuerier_Params() {
	resp, err := s.queryClient.Params(s.ctx.Context(), &types.QueryParams{})
	s.Require().NoError(err)
	s.Require().Equal(types.DefaultParams(), resp.Params)
}

func (s *IntegrationTestSuite) TestQuerier_MarketSummary() {
	req := &types.QueryMarketSummary{}
	_, err := s.queryClient.MarketSummary(context.Background(), req)
	s.Require().Error(err)

	req = &types.QueryMarketSummary{Denom: "uumee"}
	resp, err := s.queryClient.MarketSummary(context.Background(), req)
	s.Require().NoError(err)

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
	s.Require().Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_AccountBalances() {
	addr, _ := s.initBorrowScenario()

	resp, err := s.queryClient.AccountBalances(s.ctx.Context(), &types.QueryAccountBalances{Address: addr.String()})
	s.Require().NoError(err)

	expected := types.QueryAccountBalancesResponse{
		Supplied: sdk.NewCoins(
			sdk.NewCoin(umeeDenom, sdk.NewInt(1000000000)),
		),
		Collateral: sdk.NewCoins(
			sdk.NewCoin(types.ToUTokenDenom(umeeDenom), sdk.NewInt(1000000000)),
		),
		Borrowed: nil,
	}

	s.Require().Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_AccountSummary() {
	addr, _ := s.initBorrowScenario()

	resp, err := s.queryClient.AccountSummary(s.ctx.Context(), &types.QueryAccountSummary{Address: addr.String()})
	s.Require().NoError(err)

	expected := types.QueryAccountSummaryResponse{
		// This result is umee's oracle exchange rate from
		// from .Reset() in x/leverage/keeper/oracle_test.go
		// times the amount of umee, then sometimes times params
		// from newToken in x/leverage/keeper/keeper_test.go
		// (1000 / 1000000) * 4.21 = 4210
		SuppliedValue: sdk.MustNewDecFromStr("4210"),
		// (1000 / 1000000) * 4.21 = 4210
		CollateralValue: sdk.MustNewDecFromStr("4210"),
		// Nothing borrowed
		BorrowedValue: sdk.ZeroDec(),
		// (1000 / 1000000) * 4.21 * 0.25 = 1052.5
		BorrowLimit: sdk.MustNewDecFromStr("1052.5"),
		// (1000 / 1000000) * 4.21 * 0.25 = 1052.5
		LiquidationThreshold: sdk.MustNewDecFromStr("1052.5"),
	}

	s.Require().Equal(expected, *resp)
}

func (s *IntegrationTestSuite) TestQuerier_LiquidationTargets() {
	resp, err := s.queryClient.LiquidationTargets(s.ctx.Context(), &types.QueryLiquidationTargets{})
	s.Require().NoError(err)

	expected := types.QueryLiquidationTargetsResponse{
		Targets: nil,
	}

	s.Require().Equal(expected, *resp)
}
