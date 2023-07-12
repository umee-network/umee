package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/util/store"
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
