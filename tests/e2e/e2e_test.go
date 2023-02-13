package e2e

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/tests/grpc"
)

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	var ibcStakeDenom string

	valAddr, err := s.chain.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)

	s.Run("send_stake_to_umee", func() {
		recipient := valAddr.String()
		token := sdk.NewInt64Coin("stake", 3300000000) // 3300stake
		s.sendIBC(gaiaChainID, s.chain.id, recipient, token)

		umeeAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))

		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			balances sdk.Coins
			err      error
		)
		s.Require().Eventually(
			func() bool {
				balances, err = queryUmeeAllBalances(umeeAPIEndpoint, recipient)
				s.Require().NoError(err)

				return balances.Len() == 3
			},
			time.Minute,
			5*time.Second,
		)

		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				ibcStakeDenom = c.Denom
				s.Require().Equal(token.Amount.Int64(), c.Amount.Int64())
				break
			}
		}

		s.Require().NotEmpty(ibcStakeDenom)
	})
}

// TestMedians queries for the oracle params, collects historical
// prices based on those params, checks that the stored medians and
// medians deviations are correct, updates the oracle params with
// a gov prop, then checks the medians and median deviations again.
func (s *IntegrationTestSuite) TestMedians() {
	err := grpc.MedianCheck(s.umeeClient)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestUpdateOracleParams() {
	params, err := s.umeeClient.QueryClient.QueryParams()
	s.Require().NoError(err)

	s.Require().Equal(uint64(5), params.HistoricStampPeriod)
	s.Require().Equal(uint64(4), params.MaximumPriceStamps)
	s.Require().Equal(uint64(20), params.MedianStampPeriod)

	err = grpc.SubmitAndPassProposal(
		s.umeeClient,
		grpc.OracleParamChanges(10, 2, 20),
	)
	s.Require().NoError(err)

	params, err = s.umeeClient.QueryClient.QueryParams()
	s.Require().NoError(err)

	s.Require().Equal(uint64(10), params.HistoricStampPeriod)
	s.Require().Equal(uint64(2), params.MaximumPriceStamps)
	s.Require().Equal(uint64(20), params.MedianStampPeriod)

	s.Require().NoError(err)
}
