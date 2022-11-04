//go:build !experimental
// +build !experimental

// DONTCOVER

package app

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
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
	WasmProposalsEnabled        = "false"
	EnableSpecificWasmProposals = ""
	EmptyWasmOpts               []wasm.Option
)

// GetWasmEnabledProposals parses the WasmProposalsEnabled and
// EnableSpecificWasmProposals values to produce a list of enabled proposals to
// pass into the application.
func GetWasmEnabledProposals() []wasm.ProposalType {
	return []wasm.ProposalType{}
}

func setCustomKVStoreKeys() []string {
	return []string{}
}

func setCustomOrderInitGenesis() []string {
	return []string{}
}

func setCustomOrderBeginBlocker() []string {
	return []string{}
}

func setCustomOrderEndBlocker() []string {
	return []string{}
}

func setCustomOrderMigrations() []string {
	return []string{}
}

func initCustomParamsKeeper(_ *paramskeeper.Keeper) {}

func setCustomProposalHanndlers() []govclient.ProposalHandler {
	return []govclient.ProposalHandler{
		paramsclient.ProposalHandler,
		distrclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		leverageclient.ProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
	}
}

func (app *UmeeApp) setCustomAnteHandler(txConfig client.TxConfig,
	wasmConfig *wasmtypes.WasmConfig, wasmStoreKey *storetypes.KVStoreKey) (sdk.AnteHandler, error) {
	return customante.NewAnteHandler(
		customante.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			OracleKeeper:    app.OracleKeeper,
			IBCKeeper:       app.IBCKeeper,
			SignModeHandler: txConfig.SignModeHandler(),
			FeegrantKeeper:  app.FeeGrantKeeper,
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		},
		false,
	)
}

func (app *UmeeApp) setCustomModuleManager() []module.AppModule {
	return []module.AppModule{}
}

func (app *UmeeApp) registerCustomExtensions() {}

func (app *UmeeApp) registerCustomProposals(
	_ govv1beta1.Router,
	_ []wasmtypes.ProposalType,
) {
}

func (app *UmeeApp) setCustomKeepers(
	_ *baseapp.BaseApp,
	_ map[string]*storetypes.KVStoreKey,
	_ codec.Codec, _ govv1beta1.Router, _ string,
	_ servertypes.AppOptions,
	_ []wasm.Option) {
}

func (app *UmeeApp) initializeCustomScopedKeepers() {}
