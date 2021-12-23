//go:build norace
// +build norace

package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"

	umeeappbeta "github.com/umee-network/umee/app/beta"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := umeeappbeta.IntegrationTestNetworkConfig()
	cfg.NumValidators = 2
	suite.Run(t, NewIntegrationTestSuite(cfg))
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")

	//ref: https://pkg.go.dev/github.com/cosmos/cosmos-sdk/testutil/network
	s.network.Cleanup()
}
