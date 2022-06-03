package upgrades

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	bech32ibckeeper "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/keeper"
	// oraclekeeper "github.com/umee-network/umee/v2/x/oracle/keeper"
)

const UpgradeV3_0Plan = "v1.0-v3.0"

// RegisterUpgradeHandlers registers handlers for all upgrades
func RegisterUpgradeHandlers(
	mm *module.Manager, configurator *module.Configurator,
	bech32IbcKeeper *bech32ibckeeper.Keeper, upgradeKeeper *upgradekeeper.Keeper,
) {
	// v3 upgrade handler performs upgrade from v1->v3
	upgradeKeeper.SetUpgradeHandler(
		UpgradeV3_0Plan,
		func(
			ctx sdk.Context, plan upgradetypes.Plan, vmap module.VersionMap,
		) (module.VersionMap, error) {

			ctx.Logger().Info("Upgrade handler execution", "name", UpgradeV3_0Plan)

			err := setupBech32ibcKeeper(bech32IbcKeeper, ctx)
			if err != nil {
				return sdkerrors.Wrap(err, "Calypso Upgrade: Unable to upgrade, bech32ibc module not initialized")
			}

			ctx.Logger().Info("Upgrade handler execution finished, running migrations", "name", UpgradeV3_0Plan)
			return mm.RunMigrations(ctx, *configurator, vmap)

		})
}

// Sets up bech32ibc module by setting the native account prefix to "umee".
// Failing to set the native prefix will cause a chain halt on init genesis or
// in the firstBeginBlocker assertions.
func setupBech32ibcKeeper(bech32IbcKeeper *bech32ibckeeper.Keeper, ctx sdk.Context) error {
	return bech32IbcKeeper.SetNativeHrp(ctx, sdk.GetConfig().GetBech32AccountAddrPrefix())
}
