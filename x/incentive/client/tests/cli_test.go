package tests

import (
	"testing"
)

func TestCLITestSuite(t *testing.T) {
	t.Parallel()

	s := NewCLITestSuite(t)
	s.SetupSuite()
	defer s.TearDownSuite()

	// queries
	s.TestInvalidQueries()
	s.TestIncentiveScenario()
}
