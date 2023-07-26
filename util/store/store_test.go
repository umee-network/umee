package store

import (
	"math"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v5/tests/tsdk"
)

func TestGetAndSetDec(t *testing.T) {
	t.Parallel()
	store := tsdk.KVStore(t)
	key := []byte("decKey")
	val := sdk.MustNewDecFromStr("1234")
	err := SetDec(store, key, val, "no error")
	assert.NilError(t, err)

	v := GetDec(store, key, "no error")
	assert.DeepEqual(t, v, val)
}

func TestGetAndSetInt(t *testing.T) {
	t.Parallel()
	store := tsdk.KVStore(t)
	key := []byte("intKey")
	val, ok := sdk.NewIntFromString("1234")
	assert.Equal(t, true, ok)
	err := SetInt(store, key, val, "no error")
	assert.NilError(t, err)

	v := GetInt(store, key, "no error")
	assert.DeepEqual(t, v, val)
}

func checkStoreNumber[T Integer](name string, val T, store sdk.KVStore, key []byte, t *testing.T) {
	SetInteger(store, key, val)
	vOut := GetInteger[T](store, key)
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
