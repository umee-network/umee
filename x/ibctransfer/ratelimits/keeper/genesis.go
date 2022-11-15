package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/ibctransfer"
)

// InitGenesis initializes the x/leverage module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState ibctransfer.GenesisState) {
	k.SetParams(ctx, genState.Params)
	if err := k.SetRateLimitsOfIBCDenoms(ctx, genState.RateLimits); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the x/leverage module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *ibctransfer.GenesisState {
	rateLimits, err := k.GetRateLimitsOfIBCDenoms(ctx)
	if err != nil {
		panic(err)
	}

	return &ibctransfer.GenesisState{
		Params:     k.GetParams(ctx),
		RateLimits: rateLimits,
	}
}
