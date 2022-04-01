package upgrades

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	bech32ibckeeper "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/keeper"
	"github.com/umee-network/umee/app/upgrades/calypso"
	leveragekeeper "github.com/umee-network/umee/x/leverage/keeper"
	oraclekeeper "github.com/umee-network/umee/x/oracle/keeper"
)

// RegisterUpgradeHandlers registers handlers for all upgrades
// Note: This method has crazy parameters because of circular import issues, I didn't want to make a Gravity struct
// along with a Gravity interface
func RegisterUpgradeHandlers(
	mm *module.Manager, configurator *module.Configurator, accountKeeper *authkeeper.AccountKeeper,
	bankKeeper *bankkeeper.BaseKeeper, bech32IbcKeeper *bech32ibckeeper.Keeper, distrKeeper *distrkeeper.Keeper,
	mintKeeper *mintkeeper.Keeper, stakingKeeper *stakingkeeper.Keeper, upgradeKeeper *upgradekeeper.Keeper,
	leverageKeeper *leveragekeeper.Keeper, oracleKeeper *oraclekeeper.Keeper,
) {
	fmt.Println("Enter RegisterUpgradeHandlers")
	if mm == nil || configurator == nil || accountKeeper == nil || bankKeeper == nil || bech32IbcKeeper == nil ||
		distrKeeper == nil || mintKeeper == nil || stakingKeeper == nil || upgradeKeeper == nil {
		panic("Nil argument to RegisterUpgradeHandlers()!")
	}
	// Calypso aka v1->v2 UPGRADE HANDLER SETUP
	upgradeKeeper.SetUpgradeHandler(
		calypso.PlanName, // Codename Calypso
		calypso.GetV2UpgradeHandler(
			mm, configurator, accountKeeper, bankKeeper, bech32IbcKeeper,
			distrKeeper, mintKeeper, stakingKeeper, leverageKeeper, oracleKeeper,
		),
	)
}
