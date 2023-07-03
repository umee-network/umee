package e2e

import (
	"testing"

	"github.com/stretchr/testify/suite"

	setup "github.com/umee-network/umee/v5/tests/e2e/setup"
	"github.com/umee-network/umee/v5/tests/grpc"
)

type E2ETest struct {
	setup.E2ETestSuite
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETest))
}

// TestMedians queries for the oracle params, collects historical
// prices based on those params, checks that the stored medians and
// medians deviations are correct, updates the oracle params with
// a gov prop, then checks the medians and median deviations again.
func (s *E2ETest) TestMedians() {
	err := grpc.MedianCheck(s.Umee)
	s.Require().NoError(err)
}

func (s *E2ETest) TestUpdateOracleParams() {
	params, err := s.Umee.QueryOracleParams()
	s.Require().NoError(err)

	s.Require().Equal(uint64(5), params.HistoricStampPeriod)
	s.Require().Equal(uint64(4), params.MaximumPriceStamps)
	s.Require().Equal(uint64(20), params.MedianStampPeriod)

	err = grpc.SubmitAndPassProposal(
		s.Umee,
		grpc.OracleParamChanges(10, 2, 20),
	)
	s.Require().NoError(err)

	params, err = s.Umee.QueryOracleParams()
	s.Require().NoError(err)

	s.Require().Equal(uint64(10), params.HistoricStampPeriod)
	s.Require().Equal(uint64(2), params.MaximumPriceStamps)
	s.Require().Equal(uint64(20), params.MedianStampPeriod)

	s.Require().NoError(err)
}
