package e2e

import (
	"github.com/umee-network/umee/v4/tests/grpc"
)

// TestMedians queries for the oracle params, collects historical
// prices based on those params, checks that the stored medians and
// medians deviations are correct, updates the oracle params with
// a gov prop, then checks the medians and median deviations again.
func (s *IntegrationTestSuite) TestMedians() {
	err := grpc.MedianCheck(s.umeeClient)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestUpdateOracleParams() {
	s.T().Skip("paused due to validator power threshold enforcing")
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
