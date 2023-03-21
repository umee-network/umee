package store

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Iterate through all keys in a kvStore that start with a given prefix
// using a provided function. If the provided function returns an error,
// iteration stops and the error is returned.
func Iterate(store sdk.KVStore, prefix []byte, cb func(key, val []byte) error) error {
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, val := iter.Key(), iter.Value()

		if err := cb(key, val); err != nil {
			return err
		}
	}

	return nil
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

	for ; iter.Valid(); iter.Next() {
		key, val := iter.Key(), iter.Value()

		if err := cb(key, val); err != nil {
			return err
		}
	}

	return nil
}
