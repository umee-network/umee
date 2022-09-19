package app

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/nft"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ica "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	bech32ibckeeper "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/keeper"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"

	"github.com/umee-network/umee/v3/app/upgradev3"
	leveragetypes "github.com/umee-network/umee/v3/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v3/x/oracle/types"
)

const UpgradeV3_0Plan = "v1.1-v3.0"

func (app UmeeApp) RegisterUpgradeHandlers() {
	// v3 upgrade handler performs upgrade from v1->v3
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeV3_0Plan,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			ctx.Logger().Info("Upgrade handler execution", "name", UpgradeV3_0Plan)
			ctx.Logger().Info("Running setupBech32ibcKeeper")
			err := setupBech32ibcKeeper(&app.bech32IbcKeeper, ctx)
			if err != nil {
				return nil, sdkerrors.Wrapf(
					err, "%q Upgrade: Unable to upgrade, bech32ibc module not initialized", UpgradeV3_0Plan)
			}

			ctx.Logger().Info("Running module migrations")
			vm, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
			if err != nil {
				return vm, err
			}

			ctx.Logger().Info("Updating validator minimum commission rate param of staking module")
			minCommissionRate, err := upgradev3.UpdateMinimumCommissionRateParam(ctx, app.StakingKeeper)
			if err != nil {
				return vm, sdkerrors.Wrapf(
					err, "%q Upgrade: failed to update minimum commission rate param of staking module",
					UpgradeV3_0Plan)
			}
			err = upgradev3.SetMinimumCommissionRateToValidatros(ctx, app.StakingKeeper, minCommissionRate)
			if err != nil {
				return vm, sdkerrors.Wrapf(
					err, "%q Upgrade: failed to update minimum commission rate for validators",
					UpgradeV3_0Plan)
			}

			return vm, err
		})

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == UpgradeV3_0Plan && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				group.ModuleName,
				nft.ModuleName,
				bech32ibctypes.ModuleName,
				icacontrollertypes.StoreKey,
				icahosttypes.StoreKey,

				oracletypes.ModuleName,
				leveragetypes.ModuleName,
			},
		}
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(
			upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

// Sets up bech32ibc module by setting the native account prefix to "umee".
// Failing to set the native prefix will cause a chain halt on init genesis or
// in the firstBeginBlocker assertions.
func setupBech32ibcKeeper(bech32IbcKeeper *bech32ibckeeper.Keeper, ctx sdk.Context) error {
	return bech32IbcKeeper.SetNativeHrp(ctx, sdk.GetConfig().GetBech32AccountAddrPrefix())
}

// setupIBCUpdate updates IBC from v2 to v5
func setupIBCUpdate(ctx sdk.Context, app *UmeeApp, fromVM module.VersionMap) {
	// manually set the ICA params
	// the ICA module's default genesis has host and controller enabled.
	// we want these to be enabled via gov param change.

	// Add Interchain Accounts host module
	// set the ICS27 consensus version so InitGenesis is not run
	fromVM[icatypes.ModuleName] = app.mm.Modules[icatypes.ModuleName].ConsensusVersion()

	// create ICS27 Controller submodule params, controller module not enabled.
	controllerParams := icacontrollertypes.Params{ControllerEnabled: false}

	// create ICS27 Host submodule params, host module not enabled.
	hostParams := icahosttypes.Params{
		HostEnabled:   false,
		AllowMessages: []string{},
	}

	mod, found := app.mm.Modules[icatypes.ModuleName]
	if !found {
		panic(fmt.Sprintf("module %s is not in the module manager", icatypes.ModuleName))
	}

	icaMod, ok := mod.(ica.AppModule)
	if !ok {
		panic(fmt.Sprintf("expected module %s to be type %T, got %T", icatypes.ModuleName, ica.AppModule{}, mod))
	}
	icaMod.InitModule(ctx, controllerParams, hostParams)
}
