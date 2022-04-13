package app

import (
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

const (
	// DefaultUmeeInstanceCost is initially set the same as in wasmd
	DefaultUmeeInstanceCost uint64 = 60_000
	// DefaultUmeeCompileCost set to a large number for testing
	DefaultUmeeCompileCost uint64 = 100
)

var (
	// If EnabledSpecificProposals is "", and this is "true", then enable all x/wasm proposals.
	// If EnabledSpecificProposals is "", and this is not "true", then disable all x/wasm proposals.
	ProposalsEnabled = "true"
	// If set to non-empty string it must be comma-separated list of values that are all a subset
	// of "EnableAllProposals" (takes precedence over ProposalsEnabled)
	// https://github.com/CosmWasm/wasmd/blob/02a54d33ff2c064f3539ae12d75d027d9c665f05/x/wasm/internal/types/proposal.go#L28-L34
	EnableSpecificProposals = ""
)

// UmeeGasRegisterConfig is defaults plus a custom compile amount
func UmeeGasRegisterConfig() wasmkeeper.WasmGasRegisterConfig {
	gasConfig := wasmkeeper.DefaultGasRegisterConfig()
	gasConfig.InstanceCost = DefaultUmeeInstanceCost
	gasConfig.CompileCost = DefaultUmeeCompileCost

	return gasConfig
}

// NewUmeeWasmGasRegister returns a new Umee Gas Register for CosmWasm
func NewUmeeWasmGasRegister() wasmkeeper.WasmGasRegister {
	return wasmkeeper.NewWasmGasRegister(UmeeGasRegisterConfig())
}

// GetEnabledProposals parses the ProposalsEnabled / EnableSpecificProposals values to
// produce a list of enabled proposals to pass into wasmd app.
func GetEnabledProposals() []wasm.ProposalType {
	if EnableSpecificProposals == "" {
		if ProposalsEnabled == "true" {
			return wasm.EnableAllProposals
		}
		return wasm.DisableAllProposals
	}
	chunks := strings.Split(EnableSpecificProposals, ",")
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
