package tests

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"

	itestsuite "github.com/umee-network/umee/v5/tests/cli"
)

type IntegrationTests struct {
	*itestsuite.E2ESuite
}

func NewIntegrationTestSuite(cfg network.Config, t *testing.T) *IntegrationTests {
	return &IntegrationTests{
		&itestsuite.E2ESuite{Cfg: cfg, T: t},
	}
}
