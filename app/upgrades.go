package app

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/nft"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"

	"github.com/umee-network/umee/v3/app/upgradev3"
	"github.com/umee-network/umee/v3/app/upgradev3_3"
	leveragetypes "github.com/umee-network/umee/v3/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v3/x/oracle/types"
)

func (app UmeeApp) RegisterUpgradeHandlers(experimental bool) {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	app.registerV3_0_Upgrade(upgradeInfo)
	app.registerV3_1_Upgrade(upgradeInfo)

	app.registerV3_1to3_3_Upgrade(upgradeInfo)
	app.registerV3_2to3_3_Upgrade(upgradeInfo)
}

// performs upgrade from v3.1 -> v3.3 (including the v3.2 chanages)
func (app *UmeeApp) registerV3_1to3_3_Upgrade(_ upgradetypes.Plan) {
	const planName = "v3.1-v3.3"
	app.UpgradeKeeper.SetUpgradeHandler(
		planName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			ctx.Logger().Info("Upgrade handler execution", "name", planName)
			ctx.Logger().Info("Run v3.3 migrator")
			err := upgradev3_3.Migrator(app.GovKeeper, app.interfaceRegistry)(ctx)
			if err != nil {
				return fromVM, err
			}
			ctx.Logger().Info("Run x/bank v0.46.5 migration")
			err = bankkeeper.NewMigrator(app.BankKeeper).Migrate3_V046_4_To_V046_5(ctx)
			if err != nil {
				return fromVM, err
			}
			ctx.Logger().Info("Run module migrations")
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})
}

// performs upgrade from v3.2 -> v3.3
func (app *UmeeApp) registerV3_2to3_3_Upgrade(_ upgradetypes.Plan) {
	const planName = "v3.3"
	app.UpgradeKeeper.SetUpgradeHandler(
		planName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			ctx.Logger().Info("Upgrade handler execution", "name", planName)
			ctx.Logger().Info("Run v3.3 migrator")
			err := upgradev3_3.Migrator(app.GovKeeper, app.interfaceRegistry)(ctx)
			if err != nil {
				return fromVM, err
			}
			ctx.Logger().Info("Run module migrations")
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})
}

// performs upgrade from v3.0 -> v3.1
func (app *UmeeApp) registerV3_1_Upgrade(_ upgradetypes.Plan) {
	const planName = "v3.1.0"
	app.UpgradeKeeper.SetUpgradeHandler(planName, onlyModuleMigrations(app, planName))
}

// performs upgrade from v1->v3
func (app *UmeeApp) registerV3_0_Upgrade(upgradeInfo upgradetypes.Plan) {
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

	app.storeUpgrade(planName, upgradeInfo, storetypes.StoreUpgrades{
		Added: []string{
			group.ModuleName,
			nft.ModuleName,
			bech32ibctypes.ModuleName,
			// icacontrollertypes.StoreKey,
			// icahosttypes.StoreKey,

			oracletypes.ModuleName,
			leveragetypes.ModuleName,
		}})
}

func onlyModuleMigrations(app *UmeeApp, planName string) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Upgrade handler execution", "name", planName)
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	}
}

// helper function to check if the store loader should be upgraded
func (app *UmeeApp) storeUpgrade(planName string, ui upgradetypes.Plan, stores storetypes.StoreUpgrades) {
	if ui.Name == planName && !app.UpgradeKeeper.IsSkipHeight(ui.Height) {
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(
			upgradetypes.UpgradeStoreLoader(ui.Height, &stores))
	}
}
