//go:build norace
// +build norace

package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	umeeapp "github.com/umee-network/umee/v4/app"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := umeeapp.IntegrationTestNetworkConfig()
	cfg.NumValidators = 2
	cfg.Mnemonics = []string{
		"empower ridge mystery shrimp predict alarm swear brick across funny vendor essay antique vote place lava proof gaze crush head east arch twin lady",
		"clean target advice dirt onion correct original vibrant actor upon waste eternal color barely shrimp aspect fall material wait repeat bench demise length seven",
	}

	suite.Run(t, NewIntegrationTestSuite(cfg))
}
