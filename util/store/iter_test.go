package store

import (
	"testing"

	prefixstore "github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v5/tests/tsdk"
	"github.com/umee-network/umee/v5/util/keys"
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

func TestSumCoins(t *testing.T) {
	// test SumCoins using the Prefix Store, which will automatically strip the prefix from
	// keys

	prefix := "p1"
	pairs := []struct {
		K string
		V uint64
	}{
		{"atom", 1},
		{"umee", 8},
		{"atom", 8}, // we overwrite
		{"ato", 2},
		{"atoma", 3},
	}
	expected := sdk.NewCoins(
		sdk.NewInt64Coin("ato", 2),
		sdk.NewInt64Coin("atom", 8),
		sdk.NewInt64Coin("atoma", 3),
		sdk.NewInt64Coin("umee", 8))

	withPrefixAnNull := func(s string) []byte {
		return append([]byte(prefix+s), 0)
	}

	db := tsdk.KVStore(t)
	for i, p := range pairs {
		err := SetInt(db, withPrefixAnNull(p.K), sdk.NewIntFromUint64(p.V), "amount")
		assert.NilError(t, err, "pairs[%d]", i)
	}

	pdb := prefixstore.NewStore(db, []byte(prefix))
	sum := SumCoins(pdb, keys.NoLastByte)
	sum.Sort()
	assert.DeepEqual(t, expected, sum)
}
