package tests

import (
	"testing"
)

func TestCLISuite(t *testing.T) {
	t.Parallel()

	// init the integration test and start the network
	s := NewCLITestSuite(t)
	s.SetupSuite()
	defer s.TearDownSuite()

	// test cli queries
	s.TestMinGasPrice(t)
	s.TestInflationParams(t)
}
