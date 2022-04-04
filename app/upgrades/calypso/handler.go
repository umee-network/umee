package calypso

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	bech32ibckeeper "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/keeper"
	"github.com/umee-network/umee/x/leverage"
	leveragekeeper "github.com/umee-network/umee/x/leverage/keeper"
	leveragetypes "github.com/umee-network/umee/x/leverage/types"
	"github.com/umee-network/umee/x/oracle"
	oraclekeeper "github.com/umee-network/umee/x/oracle/keeper"
	oracletypes "github.com/umee-network/umee/x/oracle/types"
)

func GetV2UpgradeHandler(
	mm *module.Manager, configurator *module.Configurator, accountKeeper *authkeeper.AccountKeeper,
	bankKeeper *bankkeeper.BaseKeeper, bech32IbcKeeper *bech32ibckeeper.Keeper, distrKeeper *distrkeeper.Keeper,
	mintKeeper *mintkeeper.Keeper, stakingKeeper *stakingkeeper.Keeper, leverageKeeper *leveragekeeper.Keeper,
	oracleKeeper *oraclekeeper.Keeper,
) func(
	ctx sdk.Context, plan upgradetypes.Plan, vmap module.VersionMap,
) (module.VersionMap, error) {
	if mm == nil || configurator == nil || accountKeeper == nil || bankKeeper == nil || bech32IbcKeeper == nil ||
		distrKeeper == nil || mintKeeper == nil || stakingKeeper == nil {
		panic("Nil argument to GetV2UpgradeHandler")
	}
	return func(ctx sdk.Context, plan upgradetypes.Plan, vmap module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Calypso upgrade: Enter handler")

		ctx.Logger().Info("Calypso Upgrade: Setting up bech32ibc module's native prefix")
		err := setupBech32ibcKeeper(bech32IbcKeeper, ctx)
		if err != nil {
			panic(sdkerrors.Wrap(err, "Calypso Upgrade: Unable to upgrade, bech32ibc module not initialized"))
		}

		vmap[leveragetypes.ModuleName] = leverage.AppModule{}.ConsensusVersion()
		leverage.InitGenesis(ctx, *leverageKeeper, *leveragetypes.DefaultGenesis())

		vmap[oracletypes.ModuleName] = oracle.AppModule{}.ConsensusVersion()
		oracle.InitGenesis(ctx, *oracleKeeper, *oracletypes.DefaultGenesisState())

		ctx.Logger().Info("Calypso Upgrade: Running all configured module migrations (Should only see Gravity run)")
		return mm.RunMigrations(ctx, *configurator, vmap)
	}
}

// Sets up bech32ibc module by setting the native account prefix to "umee".
// Failing to set the native prefix will cause a chain halt on init genesis or 
// in the firstBeginBlocker assertions.
func setupBech32ibcKeeper(bech32IbcKeeper *bech32ibckeeper.Keeper, ctx sdk.Context) error {
	return bech32IbcKeeper.SetNativeHrp(ctx, sdk.GetConfig().GetBech32AccountAddrPrefix())
}
