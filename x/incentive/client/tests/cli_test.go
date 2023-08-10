package tests

import (
	"testing"

	umeeapp "github.com/umee-network/umee/v6/app"
)

func TestIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	cfg := umeeapp.IntegrationTestNetworkConfig()
	cfg.NumValidators = 2
	cfg.Mnemonics = []string{
		"empower ridge mystery shrimp predict alarm swear brick across funny vendor essay antique vote place lava proof gaze crush head east arch twin lady",
		"clean target advice dirt onion correct original vibrant actor upon waste eternal color barely shrimp aspect fall material wait repeat bench demise length seven",
	}

	s := NewIntegrationTestSuite(cfg, t)
	s.SetupSuite()
	defer s.TearDownSuite()

	// queries
	s.TestInvalidQueries()
	s.TestIncentiveScenario()
}
