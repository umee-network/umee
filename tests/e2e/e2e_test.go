package e2e

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	setup "github.com/umee-network/umee/v6/tests/e2e/setup"
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
