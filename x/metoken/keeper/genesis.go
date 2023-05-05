package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util"
	"github.com/umee-network/umee/v4/x/metoken"
)

// InitGenesis initializes the x/metoken module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState metoken.GenesisState) {
	util.Panic(k.SetParams(ctx, genState.Params))

	for _, index := range genState.Registry {
		util.Panic(k.setRegisteredIndex(ctx, index))
	}

	for _, balance := range genState.Balances {
		util.Panic(k.setIndexBalance(ctx, balance))
	}

	util.Panic(k.setNextRebalancingTime(ctx, genState.NextRebalancingTime))
	util.Panic(k.setNextInterestClaimTime(ctx, genState.NextInterestClaimTime))
}

// ExportGenesis returns the x/metoken module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *metoken.GenesisState {
	return &metoken.GenesisState{
		Params:                k.GetParams(ctx),
		Registry:              k.GetAllRegisteredIndexes(ctx),
		Balances:              k.GetAllIndexesBalances(ctx),
		NextRebalancingTime:   k.getNextRebalancingTime(ctx),
		NextInterestClaimTime: k.getNextInterestClaimTime(ctx),
	}
}
