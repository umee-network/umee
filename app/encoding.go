package app

import (
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	sdkparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/umee-network/umee/v6/app/params"
)

// MakeEncodingConfig returns the application's encoding configuration with all
// types and interfaces registered.
func MakeEncodingConfig() testutil.TestEncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
