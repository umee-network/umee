package tests

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	"gotest.tools/v3/assert"
)

type IntegrationTestSuite struct {
	cfg     network.Config
	network *network.Network
}

func initIntegrationTestSuite(cfg network.Config, t *testing.T) *IntegrationTestSuite {
	var err error
	t.Log("setting up integration test suite")
	network, err := network.New(t, t.TempDir(), cfg)
	assert.NilError(t, err)
	_, err = network.WaitForHeight(1)
	assert.NilError(t, err)
	return &IntegrationTestSuite{cfg: cfg, network: network}
}

func tearDownSuite(s *IntegrationTestSuite, t *testing.T) {
	t.Log("tearing down integration test suite")
	s.network.Cleanup()
}
