package keeper_test

import (
	"context"

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
	resp, err := s.queryClient.Params(context.Background(), &types.QueryParamsRequest{})
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

func (s *IntegrationTestSuite) TestQuerier_ReserveAmount() {
	// TODO: We need to setup borrowing first prior to testing this out.
	//
	// Ref: https://github.com/umee-network/umee/issues/93
	s.Run("missing_denom", func() {
		req := &types.QueryReserveAmountRequest{}
		_, err := s.queryClient.ReserveAmount(context.Background(), req)
		s.Require().Error(err)
	})
}
