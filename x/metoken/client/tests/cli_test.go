package tests

import (
	"testing"

	umeeapp "github.com/umee-network/umee/v6/app"
)

func TestIntegrationSuite(t *testing.T) {
	t.Parallel()
	cfg := umeeapp.IntegrationTestNetworkConfig()
	cfg.NumValidators = 2
	cfg.Mnemonics = []string{
		"empower ridge mystery shrimp predict alarm swear brick across funny vendor essay antique vote place lava proof gaze crush head east arch twin lady",
		"clean target advice dirt onion correct original vibrant actor upon waste eternal color barely shrimp aspect fall material wait repeat bench demise length seven",
	}

	// init the integration test and start the network
	s := NewIntegrationTestSuite(cfg, t)
	s.SetupSuite()
	defer s.TearDownSuite()

	// test cli queries
	s.TestInvalidQueries()
	s.TestValidQueries()

	//test cli transactions
	s.TestTransactions()
}
