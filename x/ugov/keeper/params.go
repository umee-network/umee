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

func (k Keeper) SetInflationParams(lp ugov.InflationParams) error {
	return store.SetValue(k.store, KeyInflationParams, &lp, "liquidation_params")
}

func (k Keeper) InflationParams() ugov.InflationParams {
	lp := store.GetValue[*ugov.InflationParams](k.store, KeyInflationParams, "liquidation_params")
	if lp == nil {
		return ugov.InflationParams{}
	}
	return *lp
}

func (k Keeper) SetInflationCycleStart(startTime time.Time) error {
	return store.SetBinValue(k.store, KeyInflationCycleStart, &startTime, "inflation_cycle_start")
}

func (k Keeper) GetInflationCycleStart() (*time.Time, error) {
	return store.GetBinValue[*time.Time](k.store, KeyInflationCycleStart, "inflation_cycle_start")
}
