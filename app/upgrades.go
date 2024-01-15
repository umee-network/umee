package app

import (
	"context"

	"cosmossdk.io/log"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icagenesis "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/genesis/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/umee-network/umee/v6/util"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
)

// RegisterUpgradeHandlersregisters upgrade handlers.
func (app UmeeApp) RegisterUpgradeHandlers() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	// below we only kee interesting upgrades
	app.registerOutdatedPlaceholderUpgrade("v1.1-v3.0") // upgrade from v1->v3.0
	app.registerOutdatedPlaceholderUpgrade("v3.1.0")    // upgrade from v3.0->v3.1
	app.registerOutdatedPlaceholderUpgrade("v3.1-v3.3") // upgrade from v3.1->v3.3
	app.registerOutdatedPlaceholderUpgrade("v3.2-v3.3") // upgrade from v3.2 -> v3.3
	app.registerOutdatedPlaceholderUpgrade("v4.0")      // upgrade from v3.3 -> v4.0
	app.registerOutdatedPlaceholderUpgrade("v4.0.1")    // upgrade from v4.0 -> v4.0.1
	app.registerOutdatedPlaceholderUpgrade("v4.1.0")    // upgrade from v4.0 -> v4.1
	app.registerOutdatedPlaceholderUpgrade("v4.2")
	app.registerUpgrade4_3(upgradeInfo)
	app.registerOutdatedPlaceholderUpgrade("v4.4")
	app.registerOutdatedPlaceholderUpgrade("v5.0")
	app.registerOutdatedPlaceholderUpgrade("v5.1")
	app.registerOutdatedPlaceholderUpgrade("v5.2")
	app.registerUpgrade6_0(upgradeInfo)
	app.registerOutdatedPlaceholderUpgrade("v6.1")
	app.registerOutdatedPlaceholderUpgrade("v6.2")
	app.registerUpgrade("v6.3", upgradeInfo)
}

func (app *UmeeApp) registerUpgrade6_0(upgradeInfo upgradetypes.Plan) {
	planName := "v6.0"
	gravityModuleName := "gravity" // hardcoded to avoid dependency on GB module
	emergencyGroup, err := sdk.AccAddressFromBech32("umee1gy3c8n2xysawysq2xf2hxn253srx4ehduevq6c")
	util.Panic(err)

	app.UpgradeKeeper.SetUpgradeHandler(planName,
		func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			printPlanName(planName, sdkCtx.Logger())
			if err := app.LeverageKeeper.SetParams(sdkCtx, leveragetypes.DefaultParams()); err != nil {
				return fromVM, err
			}
			app.UGovKeeperB.Keeper(&sdkCtx).SetEmergencyGroup(emergencyGroup)

			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		},
	)

	app.storeUpgrade(planName, upgradeInfo, storetypes.StoreUpgrades{
		Deleted: []string{gravityModuleName},
	})
}

// performs upgrade from v4.2 to v4.3
func (app *UmeeApp) registerUpgrade4_3(upgradeInfo upgradetypes.Plan) {
	const planName = "v4.3"
	app.UpgradeKeeper.SetUpgradeHandler(planName, onlyModuleMigrations(app, planName))
	app.UpgradeKeeper.SetUpgradeHandler(planName,
		func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			sdkCtx.Logger().Info("Upgrade handler execution", "name", planName)

			// set the ICS27 consensus version so InitGenesis is not run
			oldIcaVersion := fromVM[icatypes.ModuleName]
			g := icagenesis.GenesisState{HostGenesisState: icagenesis.DefaultHostGenesis()}
			g.HostGenesisState.Params.AllowMessages = []string{
				sdk.MsgTypeURL(&banktypes.MsgSend{}),
				sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgCancelUnbondingDelegation{}),
				sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
				sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
				sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
				sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
				sdk.MsgTypeURL(&govv1.MsgVote{}),
				sdk.MsgTypeURL(&govv1beta1.MsgVote{}),

				sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{}),
			}
			// initialize ICS27 module
			icamodule, ok := app.mm.Modules[icatypes.ModuleName].(ica.AppModule)
			if !ok {
				panic("Modules[icatypes.ModuleName] is not of type ica.AppModule")
			}
			// skip InitModule in upgrade tests after the upgrade has gone through.
			if oldIcaVersion != fromVM[icatypes.ModuleName] {
				icamodule.InitModule(sdkCtx, g.ControllerGenesisState.Params, g.HostGenesisState.Params)
			}

			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		},
	)

	app.storeUpgrade(planName, upgradeInfo, storetypes.StoreUpgrades{
		Added: []string{
			icahosttypes.StoreKey,
		},
	})
}

func onlyModuleMigrations(app *UmeeApp, planName string) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		printPlanName(planName, sdkCtx.Logger())
		sdkCtx.Logger().Info("-----------------------------\n-----------------------------")
		sdkCtx.Logger().Info("Upgrade handler execution", "name", planName)
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

// registerUpgrade sets an upgrade handler which only runs module migrations
// and adds new storages storages
func (app *UmeeApp) registerUpgrade(planName string, upgradeInfo upgradetypes.Plan, newStores ...string) {
	app.UpgradeKeeper.SetUpgradeHandler(planName, onlyModuleMigrations(app, planName))

	if len(newStores) > 0 {
		app.storeUpgrade(planName, upgradeInfo, storetypes.StoreUpgrades{
			Added: newStores,
		})
	}
}

// oldUpgradePlan is a noop, placeholder handler required for old (completed) upgrade plans.
func (app *UmeeApp) registerOutdatedPlaceholderUpgrade(planName string) {
	app.UpgradeKeeper.SetUpgradeHandler(
		planName,
		func(_ context.Context, _ upgradetypes.Plan, _ module.VersionMap) (module.VersionMap, error) {
			panic("Can't migrate state < 'head - 2' while running a logic with the 'head' version")
		})
}

func printPlanName(planName string, logger log.Logger) {
	logger.Info("-----------------------------\n-----------------------------")
	logger.Info("Upgrade handler execution", "name", planName)
}
