package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	setup "github.com/umee-network/umee/v6/tests/e2e/setup"
	"github.com/umee-network/umee/v6/tests/grpc"
)

type E2ETest struct {
	setup.E2ETestSuite
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETest))
}

// mustEventuallySucceedTx executes an sdk.Msg, retrying several times if receiving any error,
// and requires that the transaction eventually succeeded with nil error. Since this function
// retries for 5 seconds and ignores errors, it is suitable for scenario setup transaction or
// those which might require a few blocks elapsing before they succeed.
func (s *E2ETest) mustEventuallySucceedTx(msg sdk.Msg) {
	s.Require().Eventually(
		func() bool {
			return s.BroadcastTxWithRetry(msg) == nil
		},
		5*time.Second,
		500*time.Millisecond,
	)
}

// mustSucceedTx executes an sdk.Msg (retrying several times if receiving incorrect account sequence) and
// requires that the error returned is nil.
func (s *E2ETest) mustSucceedTx(msg sdk.Msg) {
	s.Require().NoError(s.BroadcastTxWithRetry(msg))
}

// mustFailTx executes an sdk.Msg (retrying several times if receiving incorrect account sequence) and
// requires that the error returned contains a given substring. If the substring is empty, simply requires
// non-nil error.
func (s *E2ETest) mustFailTx(msg sdk.Msg, errSubstring string) {
	s.Require().ErrorContains(
		s.BroadcastTxWithRetry(msg),
		errSubstring,
	)
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

	// simple retry loop to submit and pass a proposal
	for i := 0; i < 3; i++ {
		err = grpc.SubmitAndPassProposal(
			s.Umee,
			grpc.OracleParamChanges(10, 2, 20),
		)
		if err == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}

	s.Require().NoError(err, "submit and pass proposal")

	params, err = s.Umee.QueryOracleParams()
	s.Require().NoError(err)

	s.Require().Equal(uint64(10), params.HistoricStampPeriod)
	s.Require().Equal(uint64(2), params.MaximumPriceStamps)
	s.Require().Equal(uint64(20), params.MedianStampPeriod)

	s.Require().NoError(err)
}
