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
