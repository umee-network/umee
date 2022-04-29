package app

import (
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

const (
	// DefaultUmeeWasmInstanceCost is initially set the same as in wasmd
	DefaultUmeeWasmInstanceCost uint64 = 60_000
	// DefaultUmeeWasmCompileCost cost per byte compiled
	DefaultUmeeWasmCompileCost uint64 = 100
)

var (
	// WasmProposalsEnabled together with EnableSpecificProposals defines the wasm proposals.
	// If EnabledSpecificProposals is "", and this is "true", then enable all x/wasm proposals.
	// If EnabledSpecificProposals is "", and this is not "true", then disable all x/wasm proposals.
	WasmProposalsEnabled = "true"
	// WasmEnableSpecificProposals If set to non-empty string it must be comma-separated
	// list of values that are all a subset of "EnableAllProposals" (takes precedence over ProposalsEnabled)
	// https://github.com/CosmWasm/wasmd/blob/02a54d33ff2c064f3539ae12d75d027d9c665f05/x/wasm/internal/types/proposal.go#L28-L34
	WasmEnableSpecificProposals = ""
)

// UmeeGasRegisterConfig is defaults plus a custom compile amount
func UmeeGasRegisterConfig() wasmkeeper.WasmGasRegisterConfig {
	gasConfig := wasmkeeper.DefaultGasRegisterConfig()
	gasConfig.InstanceCost = DefaultUmeeWasmInstanceCost
	gasConfig.CompileCost = DefaultUmeeWasmCompileCost

	return gasConfig
}

// NewUmeeWasmGasRegister returns a new Umee Gas Register for CosmWasm
func NewUmeeWasmGasRegister() wasmkeeper.WasmGasRegister {
	return wasmkeeper.NewWasmGasRegister(UmeeGasRegisterConfig())
}

// GetWasmEnabledProposals parses the ProposalsEnabled / EnableSpecificProposals values to
// produce a list of enabled proposals to pass into wasmd app.
func GetWasmEnabledProposals() []wasm.ProposalType {
	if WasmEnableSpecificProposals == "" {
		if WasmProposalsEnabled == "true" {
			return wasm.EnableAllProposals
		}
		return wasm.DisableAllProposals
	}
	chunks := strings.Split(WasmEnableSpecificProposals, ",")
	proposals, err := wasm.ConvertToProposals(chunks)
	if err != nil {
		panic(err)
	}
	return proposals
}

// GetWasmOpts returns the wasm options
func GetWasmOpts(appOpts servertypes.AppOptions) []wasm.Option {
	var wasmOpts []wasm.Option

	wasmOpts = append(wasmOpts, wasmkeeper.WithGasRegister(NewUmeeWasmGasRegister()))

	return wasmOpts
}

// GetWasmDefaultGenesisStateParams returns umee cosmwasm default params.
func GetWasmDefaultGenesisStateParams() wasmtypes.Params {
	return wasmtypes.Params{
		CodeUploadAccess:             wasmtypes.AllowNobody,
		InstantiateDefaultPermission: wasmtypes.AccessTypeEverybody,
		// DefaultMaxWasmCodeSize limit max bytes read to prevent gzip bombs
		// It is 1200 KB in x/wasm, update it later via governance if really needed
		MaxWasmCodeSize: wasmtypes.DefaultMaxWasmCodeSize,
	}
}

// SetWasmDefaultGenesisState sets the default genesis state if the access is
// bigger than AccessTypeNobody.
func SetWasmDefaultGenesisState(cdc codec.JSONCodec, genState GenesisState) {
	var wasmGenesisState wasm.GenesisState
	cdc.MustUnmarshalJSON(genState[wasm.ModuleName], &wasmGenesisState)

	if wasmGenesisState.Params.CodeUploadAccess.Permission <= wasmtypes.AccessTypeNobody {
		return
	}

	// here we override wasm config to make it permissioned by default
	wasmGen := wasm.GenesisState{
		Params:    GetWasmDefaultGenesisStateParams(),
		Codes:     wasmGenesisState.Codes,
		Contracts: wasmGenesisState.Contracts,
		Sequences: wasmGenesisState.Sequences,
		GenMsgs:   wasmGenesisState.GenMsgs,
	}
	genState[wasm.ModuleName] = cdc.MustMarshalJSON(&wasmGen)
}
