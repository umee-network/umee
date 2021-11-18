//go:build norace
// +build norace

package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/umee-network/umee/app"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := app.IntegrationTestNetworkConfig()
	cfg.NumValidators = 2
	suite.Run(t, NewIntegrationTestSuite(cfg))
}
