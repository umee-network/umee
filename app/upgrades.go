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

const UpgradeV3_0Plan = "v1.1-v3.0"

func (app UmeeApp) RegisterUpgradeHandlers() {
	// v3 upgrade handler performs upgrade from v1->v3
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeV3_0Plan,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			ctx.Logger().Info("Upgrade handler execution", "name", UpgradeV3_0Plan)
			ctx.Logger().Info("Running setupBech32ibcKeeper")
			err := upgradev3.SetupBech32ibcKeeper(&app.bech32IbcKeeper, ctx)
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

			ctx.Logger().Info("Upgrade handler execution finished, updating minimum commission rate of all validators",
				"name", UpgradeV3_0Plan)
			err = upgradev3.SetMinimumCommissionRateToValidators(ctx, app.StakingKeeper, minCommissionRate)
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
