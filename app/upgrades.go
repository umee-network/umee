package app

import (
	"github.com/cometbft/cometbft/libs/log"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icagenesis "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/genesis/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/x/auction"
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
	app.registerUpgrade("v6.3", upgradeInfo, nil, nil, nil)
	app.registerUpgrade6_4(upgradeInfo)
	app.registerUpgrade("v6.5", upgradeInfo, nil, nil, nil)
	app.registerUpgrade6_6RC1(upgradeInfo)
	app.registerUpgrade("v6.7.0-rc1", upgradeInfo, nil, nil, nil)
	app.registerUpgrade("v6.7.0", upgradeInfo, nil, nil, nil)
	app.registerUpgrade("v6.7.2-rc1", upgradeInfo, nil, nil, nil)
	app.registerUpgrade("v6.7.2", upgradeInfo, nil, nil, nil)
	app.registerUpgrade6_7_3RC1(upgradeInfo)
	app.registerUpgrade6_7_3(upgradeInfo)

	// v6.7.4-rc1
	app.registerUpgrade("v6.7.4-rc1", upgradeInfo, nil, nil, nil)
}

func (app *UmeeApp) registerUpgrade6_7_3(_ upgradetypes.Plan) {
	planName := "v6.7.3"

	app.UpgradeKeeper.SetUpgradeHandler(planName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			printPlanName(planName, ctx.Logger())
			// Hackmd: https://hackmd.io/@DCFJtA8FRayD6p1Q0XgZHQ/HkfelYUuJx
			denom := "ibc/3F972A6BFE64248AF19C9328FA59A1270CBC57D4545A099860E035C2BA4C79FD"
			amount := sdk.NewInt(139771000000000000)
			sfrxEthAmount := sdk.NewCoins(sdk.NewCoin(denom, amount))
			toAddr := sdk.MustAccAddressFromBech32("umee1grppjc06d5p5enypk2vnl6v7n5sdpsp8adfytd")
			if err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, leveragetypes.ModuleName,
				toAddr, sfrxEthAmount); err != nil {
				return nil, err
			}
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		},
	)
}

func (app *UmeeApp) registerUpgrade6_7_3RC1(_ upgradetypes.Plan) {
	planName := "v6.7.3-rc1"

	app.UpgradeKeeper.SetUpgradeHandler(planName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			printPlanName(planName, ctx.Logger())
			// Hackmd: https://hackmd.io/@DCFJtA8FRayD6p1Q0XgZHQ/HkfelYUuJx
			denom := "test37" // sfrxeth on canon-4
			// leverage module account balances 10100000003 test37 and after the transfer it has to be 10099999003
			amount := sdk.NewInt(1000)
			sfrxEthAmount := sdk.NewCoins(sdk.NewCoin(denom, amount))
			// new account for testing the leverage module transfer
			toAddr := sdk.MustAccAddressFromBech32("umee1h5jqhu8n4nr9cnletkklxyajtzpv83r6emsgz9")
			if err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, leveragetypes.ModuleName,
				toAddr, sfrxEthAmount); err != nil {
				return nil, err
			}
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		},
	)
}

func (app *UmeeApp) registerUpgrade6_6RC1(upgradeInfo upgradetypes.Plan) {
	planName := "v6.6"

	app.UpgradeKeeper.SetUpgradeHandler(planName,
		func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			printPlanName(planName, ctx.Logger())

			// update leverage and metoken params to include burn auction fee share.
			lparams := app.LeverageKeeper.GetParams(ctx)
			lparams.RewardsAuctionFee = sdk.MustNewDecFromStr("0.02")
			if err := app.LeverageKeeper.SetParams(ctx, lparams); err != nil {
				return nil, err
			}

			mekeeper := app.MetokenKeeperB.Keeper(&ctx)
			meparams := mekeeper.GetParams()
			meparams.RewardsAuctionFeeFactor = 1000 // 10% of fees goes to rewards auction
			if err := mekeeper.SetParams(meparams); err != nil {
				return nil, err
			}

			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		},
	)

	app.storeUpgrade(planName, upgradeInfo, storetypes.StoreUpgrades{
		Added:   []string{auction.ModuleName},
		Deleted: []string{crisistypes.ModuleName},
	})
}

func (app *UmeeApp) registerUpgrade6_4(upgradeInfo upgradetypes.Plan) {
	planName := "v6.4"

	app.UpgradeKeeper.SetUpgradeHandler(planName,
		func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			printPlanName(planName, ctx.Logger())
			// Add UX denom aliases to metadata
			app.BankKeeper.SetDenomMetaData(ctx, appparams.UmeeTokenMetadata())

			// migrate leverage token settings
			tokens := app.LeverageKeeper.GetAllRegisteredTokens(ctx)
			for _, token := range tokens {
				// this will allow existing interest rate curves to pass new Token validation
				if token.KinkUtilization.GTE(token.MaxSupplyUtilization) {
					token.KinkUtilization = token.MaxSupplyUtilization
					token.KinkBorrowRate = token.MaxBorrowRate
					if err := app.LeverageKeeper.SetTokenSettings(ctx, token); err != nil {
						return fromVM, err
					}
				}
			}
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		},
	)

	app.storeUpgrade(planName, upgradeInfo, storetypes.StoreUpgrades{
		Added: []string{packetforwardtypes.ModuleName},
	})
}

func (app *UmeeApp) registerUpgrade6_0(upgradeInfo upgradetypes.Plan) {
	planName := "v6.0"
	gravityModuleName := "gravity" // hardcoded to avoid dependency on GB module
	emergencyGroup, err := sdk.AccAddressFromBech32("umee1gy3c8n2xysawysq2xf2hxn253srx4ehduevq6c")
	util.Panic(err)

	app.UpgradeKeeper.SetUpgradeHandler(planName,
		func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			printPlanName(planName, ctx.Logger())
			if err := app.LeverageKeeper.SetParams(ctx, leveragetypes.DefaultParams()); err != nil {
				return fromVM, err
			}
			app.UGovKeeperB.Keeper(&ctx).SetEmergencyGroup(emergencyGroup)

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
		func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			ctx.Logger().Info("Upgrade handler execution", "name", planName)

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
				icamodule.InitModule(ctx, g.ControllerGenesisState.Params, g.HostGenesisState.Params)
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
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		printPlanName(planName, ctx.Logger())
		ctx.Logger().Info("-----------------------------\n-----------------------------")
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

// registerUpgrade sets an upgrade handler which only runs module migrations
// and adds new storages storages
//
//nolint:unparam
func (app *UmeeApp) registerUpgrade(planName string, upgradeInfo upgradetypes.Plan, newStores []string,
	deletedStores []string, renamedStores []storetypes.StoreRename) {
	app.UpgradeKeeper.SetUpgradeHandler(planName, onlyModuleMigrations(app, planName))

	app.storeUpgrade(planName, upgradeInfo, storetypes.StoreUpgrades{
		Added:   newStores,
		Deleted: deletedStores,
		Renamed: renamedStores,
	})
}

// oldUpgradePlan is a noop, placeholder handler required for old (completed) upgrade plans.
func (app *UmeeApp) registerOutdatedPlaceholderUpgrade(planName string) {
	app.UpgradeKeeper.SetUpgradeHandler(
		planName,
		func(_ sdk.Context, _ upgradetypes.Plan, _ module.VersionMap) (module.VersionMap, error) {
			panic("Can't migrate state < 'head - 2' while running a logic with the 'head' version")
		})
}

func printPlanName(planName string, logger log.Logger) {
	logger.Info("-----------------------------\n-----------------------------")
	logger.Info("Upgrade handler execution", "name", planName)
}
