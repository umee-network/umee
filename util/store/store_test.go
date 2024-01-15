package store

import (
	"math"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/tests/tsdk"
)

func TestGetAndSetDec(t *testing.T) {
	t.Parallel()
	store := tsdk.KVStore(t)
	key := []byte("decKey")
	v1 := sdkmath.LegacyMustNewDecFromStr("1234.5679")
	v2, ok := GetDec(store, key, "no error")
	assert.Equal(t, false, ok)
	assert.DeepEqual(t, sdkmath.LegacyZeroDec(), v2)

	err := SetDec(store, key, v1, "no error")
	assert.NilError(t, err)

	v2, ok = GetDec(store, key, "no error")
	assert.Equal(t, true, ok)
	assert.DeepEqual(t, v2, v1)
}

func TestGetAndSetInt(t *testing.T) {
	t.Parallel()
	store := tsdk.KVStore(t)
	key := []byte("intKey")
	v2, ok := GetInt(store, key, "no error")
	assert.Equal(t, false, ok)
	assert.DeepEqual(t, sdkmath.ZeroInt(), v2)

	v1, ok := sdkmath.NewIntFromString("1234")
	assert.Equal(t, true, ok)
	err := SetInt(store, key, v1, "no error")
	assert.NilError(t, err)

	v2, ok = GetInt(store, key, "no error")
	assert.Equal(t, true, ok)
	assert.DeepEqual(t, v2, v1)
}

func checkStoreNumber[T Integer](name string, val T, store store.KVStore, key []byte, t *testing.T) {
	SetInteger(store, key, val)
	vOut, ok := GetInteger[T](store, key)
	require.True(t, ok)
	require.Equal(t, val, vOut, name)
}

func TestStoreNumber(t *testing.T) {
	t.Parallel()
	store := tsdk.KVStore(t)
	key := []byte("integer")

	checkStoreNumber("int32-0", int32(0), store, key, t)
	checkStoreNumber("int32-min", int32(math.MinInt32), store, key, t)
	checkStoreNumber("int32-max", int32(math.MaxInt32), store, key, t)
	checkStoreNumber("uint32-0", uint32(0), store, key, t)
	checkStoreNumber("uint32-max", uint32(math.MaxUint32), store, key, t)
	checkStoreNumber("int64-0", int64(0), store, key, t)
	checkStoreNumber("int64-min", int64(math.MinInt64), store, key, t)
	checkStoreNumber("int64-max", int64(math.MaxInt64), store, key, t)
	checkStoreNumber("uint64-0", uint64(0), store, key, t)
	checkStoreNumber("uint64-max", uint64(math.MaxUint64), store, key, t)
}

func TestSetAndGetAddress(t *testing.T) {
	store := tsdk.KVStore(t)
	key := []byte("uint32")
	val := sdk.AccAddress("1234")

	SetAddress(store, key, val)
	assert.DeepEqual(t, val, GetAddress(store, key))
}

func TestGetAndSetTime(t *testing.T) {
	t.Parallel()
	store := tsdk.KVStore(t)
	key := []byte("tKey")

	_, ok := GetTimeMs(store, key)
	assert.Equal(t, false, ok)

	val := time.Now()
	SetTimeMs(store, key, val)

	val2, ok := GetTimeMs(store, key)
	assert.Equal(t, true, ok)
	val = val.Truncate(time.Millisecond)
	assert.Equal(t, val, val2)
}
