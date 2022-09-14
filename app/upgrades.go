package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const UpgradeV110Plan = "v1.1.0"

func (app UmeeApp) RegisterUpgradeHandlers(cfgr module.Configurator) {
	// v1.1.0 upgrade is a no-op on state, but must be registered nonetheless
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeV110Plan,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			ctx.Logger().Info("Upgrade handler execution", "name", UpgradeV110Plan)

			ctx.Logger().Info("Upgrade handler execution finished, running migrations", "name", UpgradeV110Plan)
			return app.mm.RunMigrations(ctx, cfgr, fromVM)
		})
}
