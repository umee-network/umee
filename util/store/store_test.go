package store

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v4/tests/util"
)

func TestGetAndSetDec(t *testing.T) {
	store := util.KVStore(t)
	key := []byte("decKey")
	val := sdk.MustNewDecFromStr("1234")
	// set dec
	err := SetDec(store, key, val, "no error")
	assert.NilError(t, err)

	// get dec
	v := GetDec(store, key, "no error")
	assert.DeepEqual(t, v, val)
}

func TestGetAndSetInt(t *testing.T) {
	store := util.KVStore(t)
	key := []byte("intKey")
	val, ok := sdk.NewIntFromString("1234")
	assert.Equal(t, true, ok)
	// set int
	err := SetInt(store, key, val, "no error")
	assert.NilError(t, err)

	// get int
	v := GetInt(store, key, "no error")
	assert.DeepEqual(t, v, val)
}

func TestGetAndSetUint32(t *testing.T) {
	store := util.KVStore(t)
	key := []byte("uint32")
	val := uint32(1234)
	// set uint32
	err := SetUint32(store, key, val, "no error")
	assert.NilError(t, err)

	// get uint32
	v := GetUint32(store, key, "no error")
	assert.DeepEqual(t, v, val)
}

func TestGetAndSetUint64(t *testing.T) {
	store := util.KVStore(t)
	key := []byte("uint64")
	val := uint64(1234)
	// set uint64
	err := SetUint64(store, key, val, "no error")
	assert.NilError(t, err)

	// get uint64
	v := GetUint64(store, key, "no error")
	assert.DeepEqual(t, v, val)
}
