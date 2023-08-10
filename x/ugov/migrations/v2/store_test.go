package v2_test

import (
	"testing"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/tests/tsdk"
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/ugov"
	v2 "github.com/umee-network/umee/v6/x/ugov/migrations/v2"
)

func TestMigrateStore(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey(ugov.ModuleName)
	sdkContext, _ := tsdk.NewCtx(t, []storetypes.StoreKey{storeKey}, []storetypes.StoreKey{})
	kvStore := sdkContext.KVStore(storeKey)

	getInflationParams := func() *ugov.InflationParams {
		return store.GetValue[*ugov.InflationParams](kvStore, ugov.KeyInflationParams, "ip")
	}

	getInflationCycleEnd := func() (time.Time, bool) {
		return store.GetTimeMs(kvStore, ugov.KeyInflationCycleEnd)
	}

	// before migration
	_, ok := getInflationCycleEnd()
	assert.Equal(t, ok, false)
	ip := getInflationParams()
	assert.DeepEqual(t, 0, ip.Size())

	// after migration
	err := v2.MigrateStore(sdkContext, storeKey)
	assert.NilError(t, err)
	ip = getInflationParams()
	assert.DeepEqual(t, ugov.DefaultInflationParams(), *ip)

	cycleEnd, ok := getInflationCycleEnd()
	assert.Equal(t, ok, true)
	assert.DeepEqual(t, sdkContext.BlockTime().Add(ip.InflationCycle), cycleEnd)
}
