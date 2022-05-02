package cosmwasm

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
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

		vmap[wasm.ModuleName] = wasm.AppModule{}.ConsensusVersion()
		gn := wasm.GenesisState{
			Params: wasmtypes.Params{
				CodeUploadAccess:             wasmtypes.AllowNobody,
				InstantiateDefaultPermission: wasmtypes.AccessTypeEverybody,
				// DefaultMaxWasmCodeSize limit max bytes read to prevent gzip bombs
				// It is 1200 KB in x/wasm, update it later via governance if really needed
				MaxWasmCodeSize: wasmtypes.DefaultMaxWasmCodeSize,
			},
		}

		_, err := wasm.InitGenesis(
			ctx,
			wasmKeeper,
			gn,
			stakingKeeper,
			wasm.NewHandler(wasmkeeper.NewGovPermissionKeeper(wasmKeeper)),
		)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("GetCosmwasm Upgrade: Running all configured module migrations (Should only see Gravity run)")
		return mm.RunMigrations(ctx, *configurator, vmap)
	}
}
