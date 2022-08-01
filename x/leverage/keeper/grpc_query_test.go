package keeper_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

func (s *IntegrationTestSuite) TestQuerier_RegisteredTokens() {
	testCases := []struct {
		name         string
		req          *types.QueryRegisteredTokens
		registrySize int
		expectErr    bool
	}{
		{
			name:         "valid request",
			req:          &types.QueryRegisteredTokens{},
			registrySize: 2,
			expectErr:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			resp, err := s.queryClient.RegisteredTokens(context.Background(), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Len(resp.Registry, tc.registrySize)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQuerier_Params() {
	resp, err := s.queryClient.Params(context.Background(), &types.QueryParams{})
	s.Require().NoError(err)
	s.Require().NotZero(resp.Params.MinimumCloseFactor)
}

func (s *IntegrationTestSuite) TestQuerier_Borrowed() {
	s.Run("get_all_borrowed", func() {
		// TODO: We need to setup borrowing first prior to testing this out.
		//
		// Ref: https://github.com/umee-network/umee/issues/93
	})

	s.Run("get_denom_borrowed", func() {
		// TODO: We need to setup borrowing first prior to testing this out.
		//
		// Ref: https://github.com/umee-network/umee/issues/93
	})
}

func (s *IntegrationTestSuite) TestQuerier_MarketSummary() {
	s.Run("missing_denom", func() {
		req := &types.QueryMarketSummary{}
		_, err := s.queryClient.MarketSummary(context.Background(), req)
		s.Require().Error(err)
	})

	s.Run("valid denom", func() {
		req := &types.QueryMarketSummary{Denom: "uumee"}
		summ, err := s.queryClient.MarketSummary(context.Background(), req)
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
		s.Require().Equal(expected, *summ)
	})
}
