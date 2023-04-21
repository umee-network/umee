package tests

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	itestsuite "github.com/umee-network/umee/v4/tests/cli"
)

type IntegrationTests struct {
	*itestsuite.IntegrationTestSuite
}

func NewIntegrationTestSuite(cfg network.Config, t *testing.T) *IntegrationTests {
	return &IntegrationTests{
		&itestsuite.IntegrationTestSuite{Cfg: cfg, T: t},
	}
}
