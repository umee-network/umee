package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
)

// InitGenesis initializes the x/incentive module state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState incentive.GenesisState) {
	if err := k.setParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	// TODO: Set everything else
}

// ExportGenesis returns the x/incentive module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *incentive.GenesisState {
	return incentive.NewGenesisState(
		// TODO: Get everything using iterators
		k.GetParams(ctx),
		nil,
		nil,
		nil,
		k.getNextProgramID(ctx),
		k.getLastRewardsTime(ctx),
		nil,
		nil,
		nil,
		nil,
	)
}
