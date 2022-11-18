package app

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/nft"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"

	"github.com/umee-network/umee/v3/app/upgradev3"
	leveragetypes "github.com/umee-network/umee/v3/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v3/x/oracle/types"
)

func (app UmeeApp) RegisterUpgradeHandlers(experimental bool) {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	app.registerV3_0Upgrade(upgradeInfo)
	app.registerV3_1Upgrade(upgradeInfo)
	app.registerV3_2Upgrade(upgradeInfo)
}

func onlyRunMigrations(app *UmeeApp, planName string) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Upgrade handler execution", "name", planName)
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	}
}

// performs upgrade from v3.1 -> v3.2
func (app *UmeeApp) registerV3_2Upgrade(_ upgradetypes.Plan) {
	const planName = "v3.2.0"
	app.UpgradeKeeper.SetUpgradeHandler(planName, onlyRunMigrations(app, planName))
}

// performs upgrade from v3.0 -> v3.1
func (app *UmeeApp) registerV3_1Upgrade(_ upgradetypes.Plan) {
	const planName = "v3.1.0"
	app.UpgradeKeeper.SetUpgradeHandler(planName, onlyRunMigrations(app, planName))
}

// performs upgrade from v1->v3
func (app *UmeeApp) registerV3_0Upgrade(upgradeInfo upgradetypes.Plan) {
	const planName = "v1.1-v3.0"
	app.UpgradeKeeper.SetUpgradeHandler(
		planName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			ctx.Logger().Info("Upgrade handler execution", "name", planName)
			ctx.Logger().Info("Running setupBech32ibcKeeper")
			err := upgradev3.SetupBech32ibcKeeper(&app.bech32IbcKeeper, ctx)
			if err != nil {
				return nil, sdkerrors.Wrapf(
					err, "%q Upgrade: Unable to upgrade, bech32ibc module not initialized", planName)
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
					planName)
			}

			ctx.Logger().Info("Upgrade handler execution finished, updating minimum commission rate of all validators",
				"name", planName)
			err = upgradev3.SetMinimumCommissionRateToValidators(ctx, app.StakingKeeper, minCommissionRate)
			if err != nil {
				return vm, sdkerrors.Wrapf(
					err, "%q Upgrade: failed to update minimum commission rate for validators",
					planName)
			}

			return vm, err
		})

	if upgradeInfo.Name == planName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				group.ModuleName,
				nft.ModuleName,
				bech32ibctypes.ModuleName,
				// icacontrollertypes.StoreKey,
				// icahosttypes.StoreKey,

				oracletypes.ModuleName,
				leveragetypes.ModuleName,
			},
		}
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(
			upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
