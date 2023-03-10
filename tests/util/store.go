package util

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"gotest.tools/v3/assert"
)

// DefaultContext creates a sdk.Context with a fresh MemDB that can be used in tests.
func DefaultContext(t *testing.T, key storetypes.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	// cms.MountStoreWithDB(tkey, storetypes.StoreTypeTransient, db)
	err := cms.LoadLatestVersion()
	assert.NilError(t, err)
	return sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())
}

func KVStore(t *testing.T) sdk.KVStore {
	key := sdk.NewKVStoreKey("test")
	ctx := DefaultContext(t, key)
	return ctx.KVStore(key)
}
