package store

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

var _ codec.ProtoMarshaler = Slice[*sdk.Coin, sdk.Coin]{}

func TestCoinMarshal(t *testing.T) {
	c1 := sdk.NewInt64Coin("aa12", 5566)
	bz, err := c1.Marshal()
	assert.NilError(t, err)

	var c2 sdk.Coin
	err = c2.Unmarshal(bz)
	assert.NilError(t, err)

	assert.DeepEqual(t, c1, c2)

	c2 = sdk.NewInt64Coin("xrr", 8157)
	s := NewSlice(&c1, &c2, &c2)
	bz, err = s.Marshal()
	assert.NilError(t, err)

	var s2 = Slice[*sdk.Coin, sdk.Coin]{Content: []*sdk.Coin{}}
	err = s2.Unmarshal(bz)
	assert.NilError(t, err)

	assert.DeepEqual(t, s, s2)
}
