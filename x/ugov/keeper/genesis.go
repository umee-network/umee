package keeper

import "github.com/umee-network/umee/v5/x/ugov"

func (k Keeper) ExportGenesis() *ugov.GenesisState {
	return &ugov.GenesisState{
		MinGasPrice: k.MinGasPrice(),
	}
}

func (k Keeper) InitGenesis(gs *ugov.GenesisState) error {
	return k.SetMinGasPrice(gs.MinGasPrice)
}
