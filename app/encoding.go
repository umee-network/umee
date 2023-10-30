package app

import (
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/umee-network/umee/v6/app/params"
)

// MakeEncodingConfig returns the application's encoding configuration with all
// types and interfaces registered.
func MakeEncodingConfig() testutil.TestEncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	// std.RegisterLegacyAminoCodec(encodingConfig.Amino) // removed because it gets called on init of codec/legacy
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
