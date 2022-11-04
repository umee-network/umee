//go:build experimental
// +build experimental

// DONTCOVER

package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmclient "github.com/CosmWasm/wasmd/x/wasm/client"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	ibcclientclient "github.com/cosmos/ibc-go/v5/modules/core/02-client/client"
	customante "github.com/umee-network/umee/v3/ante"
	leverageclient "github.com/umee-network/umee/v3/x/leverage/client"
)

// Experimental is a flag which determines expermiental features.
// It's set via build flag.
const Experimental = false

var (
	// WasmProposalsEnabled enables all x/wasm proposals when it's value is "true"
	// and EnableSpecificWasmProposals is empty. Otherwise, all x/wasm proposals
	// are disabled.
	WasmProposalsEnabled = "true"

	// EnableSpecificWasmProposals, if set, must be comma-separated list of values
	// that are all a subset of "EnableAllProposals", which takes precedence over
	// WasmProposalsEnabled.
	//
	// See: https://github.com/CosmWasm/wasmd/blob/02a54d33ff2c064f3539ae12d75d027d9c665f05/x/wasm/internal/types/proposal.go#L28-L34
	EnableSpecificWasmProposals = ""

	// EmptyWasmOpts defines a type alias for a list of wasm options.
	EmptyWasmOpts []wasm.Option
)

// GetWasmEnabledProposals parses the WasmProposalsEnabled and
// EnableSpecificWasmProposals values to produce a list of enabled proposals to
// pass into the application.
func GetWasmEnabledProposals() []wasm.ProposalType {
	if EnableSpecificWasmProposals == "" {
		if WasmProposalsEnabled == "true" {
			return wasm.EnableAllProposals
		}

		return wasm.DisableAllProposals
	}

	chunks := strings.Split(EnableSpecificWasmProposals, ",")

	proposals, err := wasm.ConvertToProposals(chunks)
	if err != nil {
		panic(err)
	}

	return proposals
}

func setCustomKVStoreKeys() []string {
	return []string{wasm.StoreKey}
}

func (app *UmeeApp) setCustomModuleManager() []module.AppModule {
	return []module.AppModule{
		wasm.NewAppModule(app.appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
	}
}

func setCustomOrderInitGenesis() []string {
	return []string{
		// wasm after ibc transfer
		wasm.ModuleName,
	}
}

func setCustomOrderBeginBlocker() []string {
	return []string{
		wasm.ModuleName,
	}
}

func setCustomOrderEndBlocker() []string {
	return []string{
		wasm.ModuleName,
	}
}

func setCustomOrderMigrations() []string {
	return []string{wasm.ModuleName}
}

func setCustomProposalHanndlers() []govclient.ProposalHandler {
	return append([]govclient.ProposalHandler{
		paramsclient.ProposalHandler,
		distrclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		leverageclient.ProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
	}, wasmclient.ProposalHandlers...)
}

func (app *UmeeApp) setCustomAnteHandler(txConfig client.TxConfig,
	wasmConfig *wasmtypes.WasmConfig, wasmStoreKey *storetypes.KVStoreKey) (sdk.AnteHandler, error) {
	return customante.NewAnteHandler(
		customante.HandlerOptions{
			AccountKeeper:     app.AccountKeeper,
			BankKeeper:        app.BankKeeper,
			OracleKeeper:      app.OracleKeeper,
			IBCKeeper:         app.IBCKeeper,
			SignModeHandler:   txConfig.SignModeHandler(),
			FeegrantKeeper:    app.FeeGrantKeeper,
			SigGasConsumer:    ante.DefaultSigVerificationGasConsumer,
			WasmConfig:        wasmConfig,
			TXCounterStoreKey: wasmStoreKey,
		},
		true,
	)
}

func (app *UmeeApp) registerCustomExtensions() {
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}
}

func (app *UmeeApp) registerCustomProposals(
	govRouter govv1beta1.Router,
	wasmEnabledProposals []wasmtypes.ProposalType,
) {
	if len(wasmEnabledProposals) != 0 {
		govRouter.AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(app.WasmKeeper, wasmEnabledProposals))
	}
}

func (app *UmeeApp) setCustomKeepers(
	bApp *baseapp.BaseApp, keys map[string]*storetypes.KVStoreKey, appCodec codec.Codec,
	govRouter govv1beta1.Router, homePath string, appOpts servertypes.AppOptions,
	wasmOpts []wasm.Option,
) {
	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	app.wasmCfg = wasmConfig

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	availableCapabilities := "iterator,staking,stargate,cosmwasm_1_1"
	app.WasmKeeper = wasm.NewKeeper(
		appCodec,
		keys[wasm.StoreKey],
		app.GetSubspace(wasm.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.ScopedWasmKeeper,
		&app.UIBCTransferKeeper.Keeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		availableCapabilities,
		wasmOpts...,
	)
}

func initCustomParamsKeeper(paramsKeeper *paramskeeper.Keeper) {
	paramsKeeper.Subspace(wasm.ModuleName)
}

func (app *UmeeApp) initializeCustomScopedKeepers() {
	app.ScopedWasmKeeper = app.CapabilityKeeper.ScopeToModule(wasm.ModuleName)
}
