package params

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
)

// MakeEncodingConfig creates an EncodingConfig for Amino-based tests.
func MakeEncodingConfig() params.EncodingConfig {
	return params.MakeTestEncodingConfig()
}
