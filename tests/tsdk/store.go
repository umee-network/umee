package tsdk

import (
	"io"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	types "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

// NewCtx creates new context with store and mounted store keys and transient store keys.
func NewCtx(t *testing.T, keys []types.StoreKey, tkeys []types.StoreKey) (sdk.Context, types.CommitMultiStore) {
	cms := NewCommitMultiStore(t, keys, tkeys)
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())

	return ctx, cms
}

// NewCtxOneStore creates new context with only one store key
func NewCtxOneStore(t *testing.T, keys types.StoreKey) (sdk.Context, types.CommitMultiStore) {
	return NewCtx(t, []types.StoreKey{keys}, nil)
}

// NewCommitMultiStore creats SDK Multistore
func NewCommitMultiStore(t *testing.T, keys []types.StoreKey, tkeys []types.StoreKey) types.CommitMultiStore {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	for _, k := range keys {
		cms.MountStoreWithDB(k, types.StoreTypeIAVL, db)
	}
	for _, k := range tkeys {
		cms.MountStoreWithDB(k, types.StoreTypeTransient, db)
	}
	err := cms.LoadLatestVersion()
	assert.NoError(t, err)
	return cms
}

// NewKVStore creates a memory based kv store without commit / wrapping functionality
func NewKVStore(t *testing.T) types.KVStore {
	db := dbm.NewMemDB()
	return kvStoreDB{db, t}
}

type kvStoreDB struct {
	db *dbm.MemDB
	t  *testing.T
}

func (kvStoreDB) CacheWrap() types.CacheWrap {
	panic("not implemented")
}

func (kvStoreDB) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	panic("not implemented")
}

func (s kvStoreDB) Get(key []byte) []byte {
	o, err := s.db.Get(key)
	assert.NoError(s.t, err)
	return o
}

func (s kvStoreDB) Has(key []byte) bool {
	o, err := s.db.Has(key)
	assert.NoError(s.t, err)
	return o
}

func (s kvStoreDB) Set(key, val []byte) {
	err := s.db.Set(key, val)
	assert.NoError(s.t, err)
}

func (s kvStoreDB) Delete(key []byte) {
	err := s.db.Delete(key)
	assert.NoError(s.t, err)
}

func (s kvStoreDB) GetStoreType() types.StoreType {
	return types.StoreTypeMemory
}

func (s kvStoreDB) Iterator(start, end []byte) types.Iterator {
	o, err := s.db.Iterator(start, end)
	assert.NoError(s.t, err)
	return o
}

func (s kvStoreDB) ReverseIterator(start, end []byte) types.Iterator {
	o, err := s.db.ReverseIterator(start, end)
	assert.NoError(s.t, err)
	return o
}
