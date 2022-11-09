package ibc_rate_limit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/keeper"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/types"
)

// InitGenesis initializes the x/leverage module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	if err := k.SetRateLimitsOfIBCDenoms(ctx, genState.RateLimits); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the x/leverage module's exported genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	rateLimits, err := k.GetRateLimitsOfIBCDenoms(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params:     k.GetParams(ctx),
		RateLimits: rateLimits,
	}
}
