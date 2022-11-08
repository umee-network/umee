//go:build !experimental
// +build !experimental

// DONTCOVER

package app

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// Experimental is a flag which determines expermiental features.
// It's set via build flag.
const Experimental = false

var (
	WasmProposalsEnabled        = "false"
	EnableSpecificWasmProposals = ""
	EmptyWasmOpts               []wasm.Option
)

// GetWasmEnabledProposals parses the WasmProposalsEnabled and
// EnableSpecificWasmProposals values to produce a list of enabled proposals to
// pass into the application.
func GetWasmEnabledProposals() []wasm.ProposalType { return []wasm.ProposalType{} }

func (app *UmeeApp) registerCustomExtensions() {}

func (app *UmeeApp) customKeepers(
	_ *baseapp.BaseApp,
	_ map[string]*storetypes.KVStoreKey,
	_ codec.Codec, _ govv1beta1.Router, _ string,
	_ servertypes.AppOptions,
	_ []wasm.Option) {
}

func (app *UmeeApp) initializeCustomScopedKeepers() {}
