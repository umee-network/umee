package tests

import (
	"testing"

	itestsuite "github.com/umee-network/umee/v6/tests/cli"
)

type CLITests struct {
	*itestsuite.CLISuite
}

func NewCLITestSuite(t *testing.T) *CLITests {
	return &CLITests{
		&itestsuite.CLISuite{T: t},
	}
}
