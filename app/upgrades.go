package app

import (
	"github.com/cometbft/cometbft/libs/log"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icagenesis "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/genesis/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/cosmos/cosmos-sdk/baseapp"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/umee-network/umee/v6/util"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/uibc"
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
	app.registerUpgrade6(upgradeInfo)
	app.registerOutdatedPlaceholderUpgrade("v6.1")

	app.registerUpgrade6_2(upgradeInfo)
}

func (app *UmeeApp) registerUpgrade6_2(upgradeInfo upgradetypes.Plan) {
	planName := "v6.2"

	// Set param key table for params module migration
	for _, subspace := range app.ParamsKeeper.GetSubspaces() {
		subspace := subspace
		found := true
		var keyTable paramstypes.KeyTable
		switch subspace.Name() {
		case authtypes.ModuleName:
			keyTable = authtypes.ParamKeyTable() //nolint: staticcheck // deprecated but required for upgrade
		case banktypes.ModuleName:
			keyTable = banktypes.ParamKeyTable() //nolint: staticcheck // deprecated but required for upgrade
		case stakingtypes.ModuleName:
			keyTable = stakingtypes.ParamKeyTable()
		case minttypes.ModuleName:
			keyTable = minttypes.ParamKeyTable() //nolint: staticcheck // deprecated but required for upgrade
		case distrtypes.ModuleName:
			keyTable = distrtypes.ParamKeyTable() //nolint: staticcheck // deprecated but required for upgrade
		case slashingtypes.ModuleName:
			keyTable = slashingtypes.ParamKeyTable() //nolint: staticcheck // deprecated but required for upgrade
		case govtypes.ModuleName:
			keyTable = govv1.ParamKeyTable() //nolint: staticcheck // deprecated but required for upgrade
		case crisistypes.ModuleName:
			keyTable = crisistypes.ParamKeyTable() //nolint: staticcheck // deprecated but required for upgrade
		case wasmtypes.ModuleName:
			keyTable = wasmtypes.ParamKeyTable() //nolint: staticcheck // deprecated but required for upgrade
		default:
			// subspace not handled
			found = false
		}
		if found && !subspace.HasKeyTable() {
			subspace.WithKeyTable(keyTable)
		}
	}
	baseAppLegacySS := app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

	app.UpgradeKeeper.SetUpgradeHandler(planName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			printPlanName(planName, ctx.Logger())

			// Migrate CometBFT consensus parameters from x/params module to a dedicated x/consensus module.
			baseapp.MigrateParams(ctx, baseAppLegacySS, &app.ConsensusParamsKeeper)

			// explicitly update the IBC 02-client params, adding the localhost client type
			params := app.IBCKeeper.ClientKeeper.GetParams(ctx)
			params.AllowedClients = append(params.AllowedClients, ibcexported.Localhost)
			app.IBCKeeper.ClientKeeper.SetParams(ctx, params)

			fromVM, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
			if err != nil {
				return fromVM, err
			}

			// migration UMEE -> UX token metadata
			app.BankKeeper.SetDenomMetaData(ctx, umeeTokenMetadata())

			// Initialize a new Cosmos SDK 0.47 parameter: MinInitialDepositRatio
			govParams := app.GovKeeper.GetParams(ctx)
			govParams.MinInitialDepositRatio = sdk.NewDecWithPrec(1, 1).String()
			err = app.GovKeeper.SetParams(ctx, govParams)
			if err != nil {
				return fromVM, err
			}

			// uibc migrations
			uIBCKeeper := app.UIbcQuotaKeeperB.Keeper(&ctx)
			uIBCKeeper.MigrateTotalOutflowSum()
			err = uIBCKeeper.SetParams(uibc.DefaultParams())

			return fromVM, err
		},
	)

	app.storeUpgrade(planName, upgradeInfo, storetypes.StoreUpgrades{
		Added: []string{
			consensustypes.ModuleName,
			crisistypes.ModuleName,
		},
	})
	// app.registerNewTokenEmissionUpgrade(upgradeInfo)
}

func (app *UmeeApp) registerUpgrade6(upgradeInfo upgradetypes.Plan) {
	planName := "v6.0"
	gravityModuleName := "gravity" // hardcoded to avoid dependency on GB module
	emergencyGroup, err := sdk.AccAddressFromBech32("umee1gy3c8n2xysawysq2xf2hxn253srx4ehduevq6c")
	util.Panic(err)

	app.UpgradeKeeper.SetUpgradeHandler(planName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
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
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
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
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
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
//nolint:unused
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
		func(_ sdk.Context, _ upgradetypes.Plan, _ module.VersionMap) (module.VersionMap, error) {
			panic("Can't migrate state < 'head - 2' while running a logic with the 'head' version")
		})
}

func printPlanName(planName string, logger log.Logger) {
	logger.Info("-----------------------------\n-----------------------------")
	logger.Info("Upgrade handler execution", "name", planName)
}
