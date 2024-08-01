package app

import (
	"github.com/cosmos/cosmos-sdk/runtime"
)

// CosmosApp defines the common methods for a Cosmos SDK-based application-specific
// blockchain.
type CosmosApp interface {
	runtime.AppI
	// All the registered module account addresses.
	ModuleAccountAddrs() map[string]bool
}
