package module

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
	"github.com/umee-network/umee/v4/x/incentive/keeper"
)

// InitGenesis initializes the x/incentive module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState incentive.GenesisState) {
	k.InitGenesis(ctx, genState)
}

// ExportGenesis returns the x/incentive module's exported genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *incentive.GenesisState {
	return k.ExportGenesis(ctx)
}
