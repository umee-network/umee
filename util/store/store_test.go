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
