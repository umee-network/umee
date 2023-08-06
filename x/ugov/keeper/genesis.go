package keeper

import (
	"github.com/umee-network/umee/v5/x/ugov"
)

func (k Keeper) ExportGenesis() *ugov.GenesisState {
	cycleEndTime := k.InflationCycleEnd()
	return &ugov.GenesisState{
		MinGasPrice:       k.MinGasPrice(),
		InflationParams:   k.InflationParams(),
		InflationCycleEnd: cycleEndTime,
	}
}

func (k Keeper) InitGenesis(gs *ugov.GenesisState) error {
	if err := k.SetMinGasPrice(gs.MinGasPrice); err != nil {
		return err
	}
	if err := k.SetInflationParams(gs.InflationParams); err != nil {
		return err
	}
	return k.SetInflationCycleEnd(gs.InflationCycleEnd)
}
