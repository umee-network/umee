package e2e

import (
	"time"

	"github.com/umee-network/umee/v6/tests/grpc"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

func (s *E2ETest) TestMetokenSwapAndRedeem() {
	tokens := []ltypes.Token{
		mocks.ValidToken(mocks.USDTBaseDenom, mocks.USDTSymbolDenom, 6),
		mocks.ValidToken(mocks.USDCBaseDenom, mocks.USDCSymbolDenom, 6),
		mocks.ValidToken(mocks.ISTBaseDenom, mocks.ISTSymbolDenom, 6),
	}

	err := grpc.LeverageRegistryUpdate(s.Umee, tokens, nil)
	s.Require().NoError(err)

	index := mocks.StableIndex(mocks.MeUSDDenom)
	err = grpc.MetokenRegistryUpdate(s.Umee, []metoken.Index{index}, nil)
	s.Require().NoError(err)

	umeeAPIEndpoint := s.UmeeREST()
	s.checkMetokenBalance(umeeAPIEndpoint, index.Denom, mocks.EmptyUSDIndexBalances(mocks.MeUSDDenom))
}

func (s *E2ETest) checkMetokenBalance(umeeAPIEndpoint, denom string, expectedBalance metoken.IndexBalances) {
	s.Require().Eventually(
		func() bool {
			resp, err := s.QueryMetokenBalances(umeeAPIEndpoint, denom)
			if err != nil {
				return false
			}

			var exist bool
			for _, balance := range resp.IndexBalances {
				if balance.MetokenSupply.Denom == expectedBalance.MetokenSupply.Denom {
					exist = true
					s.Require().Equal(expectedBalance, balance)
					break
				}
			}

			s.Require().True(exist)
			return true
		},
		30*time.Second,
		500*time.Millisecond,
	)
}
