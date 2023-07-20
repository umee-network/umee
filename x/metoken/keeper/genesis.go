package keeper

import (
	"github.com/umee-network/umee/v5/util"
	"github.com/umee-network/umee/v5/x/metoken"
)

// InitGenesis initializes the x/metoken module's state from a provided genesis state.
func (k Keeper) InitGenesis(genState metoken.GenesisState) {
	util.Panic(k.SetParams(genState.Params))

	for _, index := range genState.Registry {
		util.Panic(k.setRegisteredIndex(index))
	}

	for _, balance := range genState.Balances {
		util.Panic(k.setIndexBalances(balance))
	}

	k.setNextRebalancingTime(genState.NextRebalancingTime)
	k.setNextInterestClaimTime(genState.NextInterestClaimTime)
}

// ExportGenesis returns the x/metoken module's exported genesis state.
func (k Keeper) ExportGenesis() *metoken.GenesisState {
	nextRebalancingTime, err := k.getNextRebalancingTime()
	util.Panic(err)

	nextInterestClaimTime, err := k.getNextInterestClaimTime()
	util.Panic(err)

	return &metoken.GenesisState{
		Params:                k.GetParams(),
		Registry:              k.GetAllRegisteredIndexes(),
		Balances:              k.GetAllIndexesBalances(),
		NextRebalancingTime:   nextRebalancingTime,
		NextInterestClaimTime: nextInterestClaimTime,
	}
}
