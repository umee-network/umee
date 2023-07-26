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

func (k Keeper) SetInflationParams(ip ugov.InflationParams) error {
	return store.SetValue(k.store, keyInflationParams, &ip, "inflation_params")
}

func (k Keeper) InflationParams() ugov.InflationParams {
	ip := store.GetValue[*ugov.InflationParams](k.store, keyInflationParams, "inflation_params")
	if ip == nil {
		return ugov.InflationParams{}
	}
	return *ip
}

func (k Keeper) SetInflationCycleEnd(startTime time.Time) error {
	store.SetTimeMs(k.store, keyInflationCycleEnd, startTime)
	return nil
}

func (k Keeper) GetInflationCycleEnd() (time.Time, error) {
	return store.GetTimeMs(k.store, keyInflationCycleEnd)
}
