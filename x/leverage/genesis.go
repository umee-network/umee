package leverage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/x/leverage/keeper"
	"github.com/umee-network/umee/x/leverage/types"
)

// InitGenesis initializes the x/leverage module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.InitGenesis(ctx, genState)
}

// ExportGenesis returns the x/leverage module's exported genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return k.ExportGenesis(ctx)
}
