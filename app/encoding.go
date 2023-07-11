package app

import (
	"github.com/umee-network/umee/v5/app/params"

	"github.com/cosmos/cosmos-sdk/types/module/testutil"
)

// MakeEncodingConfig returns the application's encoding configuration with all
// types and interfaces registered.
func MakeEncodingConfig() testutil.TestEncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
