package gravity

import (
	"github.com/umee-network/umee/x/gravity/simulation"

	"github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity"
	"github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// type check to ensure the interface is properly implemented
var (
	_ module.AppModuleSimulation = AppModule{}
)

// AppModule object for module implementation
type AppModule struct {
	gravity.AppModule
	keeper        keeper.Keeper
	bankKeeper    bankkeeper.Keeper
	accountKeeper distrtypes.AccountKeeper
	stakingKeeper stakingkeeper.Keeper
	cdc           codec.Codec
}

// NewAppModule creates a new AppModule Object
func NewAppModule(
	k keeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
	accountKeeper distrtypes.AccountKeeper,
) AppModule {
	return AppModule{
		keeper:        k,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		stakingKeeper: stakingKeeper,
	}
}

// WeightedOperations returns the all the gravity module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc, am.stakingKeeper, am.accountKeeper, am.bankKeeper, am.keeper, am.cdc,
	)
}
