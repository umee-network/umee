package store

import (
	"math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v4/tests/util"
)

func TestGetAndSetDec(t *testing.T) {
	t.Parallel()
	store := util.KVStore(t)
	key := []byte("decKey")
	val := sdk.MustNewDecFromStr("1234")
	err := SetDec(store, key, val, "no error")
	assert.NilError(t, err)

	v := GetDec(store, key, "no error")
	assert.DeepEqual(t, v, val)
}

func TestGetAndSetInt(t *testing.T) {
	t.Parallel()
	store := util.KVStore(t)
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
	assert.DeepEqual(t, val, vOut)
}

func TestStoreNumber(t *testing.T) {
	t.Parallel()
	store := util.KVStore(t)
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
	store := util.KVStore(t)
	key := []byte("uint32")
	val := sdk.AccAddress("1234")
	err := SetAddress(store, key, val, "no error")
	assert.NilError(t, err)

	v := GetAddress(store, key, "no error")
	assert.DeepEqual(t, v, val)
}
