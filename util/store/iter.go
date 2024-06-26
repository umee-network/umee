package store

import (
	"fmt"

	"github.com/umee-network/umee/v6/util"

	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	db "github.com/cometbft/cometbft-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Iterate through all keys in a kvStore that start with a given prefix
// using a provided function. If the provided function returns an error,
// iteration stops and the error is returned.
func Iterate(store store.KVStore, prefix []byte, cb func(key, val []byte) error) error {
	iter := storetypes.KVStorePrefixIterator(store, prefix)
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
func IteratePaginated(store store.KVStore, prefix []byte, page, limit uint, cb func(key, val []byte) error) error {
	iter := storetypes.KVStorePrefixIteratorPaginated(store, prefix, page, limit)
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
func LoadAll[TPtr PtrUnmarshalable[T], T any](s storetypes.KVStore, prefix []byte) ([]T, error) {
	iter := storetypes.KVStorePrefixIterator(s, prefix)
	defer iter.Close()
	out := make([]T, 0)
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
func MustLoadAll[TPtr PtrUnmarshalable[T], T any](s storetypes.KVStore, prefix []byte) []T {
	ls, err := LoadAll[TPtr, T](s, prefix)
	util.Panic(err)
	return ls
}

// LoadAllKV iterates over all records in the prefix store and unmarshals value into the list.
func LoadAllKV[KPtr PtrUnmarshalable[K], K any, TPtr PtrUnmarshalable[T], T any](
	s storetypes.KVStore, prefix []byte) ([]KV[K, T], error) {

	iter := sdk.KVStorePrefixIterator(s, prefix)
	defer iter.Close()
	out := make([]KV[K, T], 0)
	for ; iter.Valid(); iter.Next() {
		var o TPtr = new(T)
		if err := o.Unmarshal(iter.Value()); err != nil {
			return nil, fmt.Errorf("can not unmarshal value: %w", err)
		}
		var k KPtr = new(K)
		bz := iter.Key()[len(prefix):]
		if err := k.Unmarshal(bz); err != nil {
			return nil, fmt.Errorf("can not unmarshal key: %w", err)
		}
		out = append(out, KV[K, T]{*k, *o})
	}

	return out, nil
}

// SumCoins aggregates all coins saved as (denom: Int) pairs in store. Use store/prefix.NewStore
// to create a prefix store which will automatically look only at the given prefix.
func SumCoins(s storetypes.KVStore, f StrExtractor) sdk.Coins {
	total := sdk.NewCoins()
	iter := storetypes.KVStorePrefixIterator(s, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		denom := f(iter.Key())
		amount := Int(iter.Value(), "amount")
		total = total.Add(sdk.NewCoin(denom, amount))
	}
	return total
}

// StrExtractor is a function type which will take a bytes string value and extracts
// string out of it.
type StrExtractor func([]byte) string

// LoadAllDecCoins iterates over all records in the prefix store and unmarshals value into the dec coin list.
func LoadAllDecCoins(iter db.Iterator, prefixLen int) (sdk.DecCoins, error) {
	var coins sdk.DecCoins
	for ; iter.Valid(); iter.Next() {
		key, val := iter.Key(), iter.Value()
		o := sdk.DecCoin{Denom: string(key[prefixLen:])}
		if err := o.Amount.Unmarshal(val); err != nil {
			return nil, err
		}
		coins = append(coins, o)
	}

	return coins, nil
}
