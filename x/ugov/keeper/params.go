package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/util/store"
	"github.com/umee-network/umee/v5/x/ugov"
)

func (k Keeper) SetMinGasPrice(p sdk.DecCoin) error {
	return store.SetValue(k.store, keyMinGasPrice, &p, "gas_price")
}

func (k Keeper) MinGasPrice() sdk.DecCoin {
	gp := store.GetValue[*sdk.DecCoin](k.store, keyMinGasPrice, "gas_price")
	if gp == nil {
		return coin.Umee0dec
	}
	return *gp
}

func (k Keeper) SetEmergencyGroup(p sdk.AccAddress) {
	store.SetAddress(k.store, keyEmergencyGroup, p)
}

func (k Keeper) EmergencyGroup() sdk.AccAddress {
	return store.GetAddress(k.store, keyEmergencyGroup)
}

func (k Keeper) SetLiquidationParams(lp ugov.LiquidationParams) error {
	return store.SetValue(k.store, KeyLiquidationParams, &lp, "liquidation_params")
}

func (k Keeper) LiquidationParams() ugov.LiquidationParams {
	lp := store.GetValue[*ugov.LiquidationParams](k.store, KeyLiquidationParams, "liquidation_params")
	if lp == nil {
		return ugov.LiquidationParams{}
	}
	return *lp
}

func (k Keeper) SetInflationCycleStartTime(startTime time.Time) error {
	return store.SetBinValue(k.store, KeyInflationCycleStartTime, &startTime, "inflation_cycle_start_time")
}

func (k Keeper) GetInflationCycleStartTime() (*time.Time, error) {
	return store.GetBinValue[*time.Time](k.store, KeyInflationCycleStartTime, "inflation_cycle_start_time")
}
