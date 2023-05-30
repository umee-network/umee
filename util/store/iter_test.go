package store

import (
	"testing"

	"github.com/umee-network/umee/v4/tests/tsdk"
	"gotest.tools/v3/assert"
)

func TestIterate(t *testing.T) {
	// test: insert data out of order
	// expect: iterate will fetch data in order.

	type pair struct {
		K, V []byte
	}
	pairs := []pair{
		{[]byte{0}, []byte{0}},
		{[]byte{1, 1}, []byte{0}},
		{[]byte{1, 7}, []byte{3}},
		{[]byte{1, 8}, []byte{2}},
		{[]byte{2, 4}, []byte{1}},
	}

	db := tsdk.KVStore(t)
	db.Set(pairs[0].K, pairs[0].V)
	db.Set(pairs[4].K, pairs[4].V)
	db.Set(pairs[3].K, pairs[3].V)
	db.Set(pairs[1].K, pairs[1].V)
	db.Set(pairs[2].K, pairs[2].V)

	collected := []pair{}
	collect := func(k, v []byte) error {
		collected = append(collected, pair{k, v})
		return nil
	}
	Iterate(db, []byte{1}, collect)

	assert.DeepEqual(t, pairs[1:4], collected)
}
