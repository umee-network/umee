package keeper

import "github.com/umee-network/umee/v5/x/ugov"

func (k Keeper) ExportGenesis() *ugov.GenesisState {
	return &ugov.GenesisState{
		MinGasPrice:       k.MinGasPrice(),
		LiquidationParams: k.LiquidationParams(),
	}
}

func (k Keeper) InitGenesis(gs *ugov.GenesisState) error {
	if err := k.SetMinGasPrice(gs.MinGasPrice); err != nil {
		return err
	}

	return k.SetLiquidationParams(gs.LiquidationParams)
}
