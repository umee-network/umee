package store

import (
	"github.com/umee-network/umee/v4/util"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"
)

// Iterate through all keys in a kvStore that start with a given prefix
// using a provided function. If the provided function returns an error,
// iteration stops and the error is returned.
func Iterate(store sdk.KVStore, prefix []byte, cb func(key, val []byte) error) error {
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	return iterate(iter, cb)
}

// IteratePaginated through keys in a kvStore that start with a given prefix
// using a provided function. If the provided function returns an error,
// iteration stops and the error is returned.
// Accepts pagination parameters: limit defines a number of keys per page, and page
// indicates what page to skip to when iterating.
// For example, page = 3 and limit = 10 will iterate over the 21st - 30th keys that
// would be found by a non-paginated iterator.
func IteratePaginated(store sdk.KVStore, prefix []byte, page, limit uint, cb func(key, val []byte) error) error {
	iter := sdk.KVStorePrefixIteratorPaginated(store, prefix, page, limit)
	defer iter.Close()
	return iterate(iter, cb)
}

func iterate(iter db.Iterator, cb func(key, val []byte) error) error {
	for ; iter.Valid(); iter.Next() {
		key, val := iter.Key(), iter.Value()
		if err := cb(key, val); err != nil {
			return err
		}
	}
	return nil
}

// LoadAll iterates over all records in the prefix store and unmarshals value into the list.
func LoadAll[TPtr PtrProtoMarshalable[T], T any](s storetypes.KVStore, prefix []byte) ([]T, error) {
	iter := sdk.KVStorePrefixIterator(s, prefix)
	defer iter.Close()
	var out = make([]T, 0)
	for ; iter.Valid(); iter.Next() {
		var o TPtr = new(T)
		if err := o.Unmarshal(iter.Value()); err != nil {
			return nil, err
		}
		out = append(out, *o)
	}

	return out, nil
}

// MustLoadAll executes LoadAll and panics on error
func MustLoadAll[TPtr PtrProtoMarshalable[T], T any](s storetypes.KVStore, prefix []byte) []T {
	ls, err := LoadAll[TPtr, T](s, prefix)
	util.Panic(err)
	return ls
}

type Marshalable interface {
	Marshal() ([]byte, error)
	MarshalTo(data []byte) (n int, err error)
	Unmarshal(data []byte) error
}

type PtrProtoMarshalable[T any] interface {
	Marshalable
	*T
}
