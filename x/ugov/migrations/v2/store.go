package v2

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v5/util/store"
	"github.com/umee-network/umee/v5/x/ugov"
)

// MigrateStore performs in-place store migrations from 1 to 2
func MigrateStore(ctx sdk.Context, key storetypes.StoreKey) error {
	kvStore := ctx.KVStore(key)

	ip := ugov.DefaultInflationParams()
	if err := store.SetValue(kvStore, ugov.KeyInflationParams, &ip, "inflation_params"); err != nil {
		return err
	}

	cycleEnd := ctx.BlockTime().Add(ip.InflationCycle)
	store.SetTimeMs(kvStore, ugov.KeyInflationCycleEnd, cycleEnd)

	return nil
}
