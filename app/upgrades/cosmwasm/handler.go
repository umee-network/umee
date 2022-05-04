package cosmwasm

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// GetCosmwasmUpgradeHandler contains the handler for the Cosmwasm upgrade.
// It setups the Cosmwasm module.
func GetCosmwasmUpgradeHandler(
	mm *module.Manager, configurator *module.Configurator, accountKeeper *authkeeper.AccountKeeper,
	stakingKeeper *stakingkeeper.Keeper, wasmKeeper *wasm.Keeper,
) upgradetypes.UpgradeHandler {
	if mm == nil || configurator == nil || wasmKeeper == nil || accountKeeper == nil || stakingKeeper == nil {
		panic("Nil argument to GetCosmwasmUpgradeHandler")
	}
	return func(ctx sdk.Context, plan upgradetypes.Plan, vmap module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("GetCosmwasm upgrade: Enter handler")

		// Set wasm old version to 1 if we want to call wasm's InitGenesis ourselves
		// in this upgrade logic ourselves.
		//
		// vm[wasm.ModuleName] = wasm.ConsensusVersion
		//
		// Otherwise we run this, which will run wasm.InitGenesis(wasm.DefaultGenesis())
		// and then override it after.
		newVM, err := mm.RunMigrations(ctx, *configurator, vmap)
		if err != nil {
			return newVM, err
		}
		ctx.Logger().Info("GetCosmwasm upgrade: Run the migration")

		// Since we provide custom DefaultGenesis (privileges StoreCode) in
		// app/genesis.go rather than the wasm module, we need to set the params
		// here when migrating (is it is not customized).
		params := wasmKeeper.GetParams(ctx)
		params.CodeUploadAccess = wasmtypes.AllowNobody
		wasmKeeper.SetParams(ctx, params)
		ctx.Logger().Info("GetCosmwasm upgrade: Set params")

		return newVM, err
	}
}
