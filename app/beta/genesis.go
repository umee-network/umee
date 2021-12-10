package beta

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/umee-network/umee/app"
)

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) app.GenesisState {
	return ModuleBasics.DefaultGenesis(cdc)
}
