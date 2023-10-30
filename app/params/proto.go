package params

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
)

// MakeEncodingConfig creates an EncodingConfig for Amino-based tests.
func MakeEncodingConfig(modules ...module.AppModuleBasic) testutil.TestEncodingConfig {
	return testutil.MakeTestEncodingConfig(modules...)
}
