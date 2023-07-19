package keeper

import (
	"github.com/umee-network/umee/v5/util"
	"github.com/umee-network/umee/v5/x/ugov"
)

func (k Keeper) ExportGenesis() *ugov.GenesisState {
	lcst, err := k.GetInflationCycleStart()
	util.Panic(err)
	return &ugov.GenesisState{
		MinGasPrice:         k.MinGasPrice(),
		InflationParams:     k.InflationParams(),
		InflationCycleStart: *lcst,
	}
}

func (k Keeper) InitGenesis(gs *ugov.GenesisState) error {
	if err := k.SetMinGasPrice(gs.MinGasPrice); err != nil {
		return err
	}
	if err := k.SetInflationParams(gs.InflationParams); err != nil {
		return err
	}
	return k.SetInflationCycleStart(gs.InflationCycleStart)
}
