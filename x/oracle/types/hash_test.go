package types

import (
	"encoding/hex"
	"testing"

	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestAggregateVoteHash(t *testing.T) {
	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
	}

	aggregateVoteHash := GetAggregateVoteHash("salt", "UMEE:100,ATOM:100", sdk.ValAddress(addrs[0]))
	hexStr := hex.EncodeToString(aggregateVoteHash)
	aggregateVoteHashRes, err := AggregateVoteHashFromHex(hexStr)
	assert.NilError(t, err)
	assert.DeepEqual(t, aggregateVoteHash, aggregateVoteHashRes)
	assert.Equal(t, true, aggregateVoteHash.Equal(aggregateVoteHash))
	assert.Equal(t, true, AggregateVoteHash([]byte{}).Empty())

	got, _ := yaml.Marshal(&aggregateVoteHash)
	assert.Equal(t, aggregateVoteHash.String()+"\n", string(got))

	res := AggregateVoteHash{}
	testMarshal(t, &aggregateVoteHash, &res, aggregateVoteHash.MarshalJSON, (&res).UnmarshalJSON)
	testMarshal(t, &aggregateVoteHash, &res, aggregateVoteHash.Marshal, (&res).Unmarshal)
}

func testMarshal(t *testing.T, original, res interface{}, marshal func() ([]byte, error), unmarshal func([]byte) error) {
	bz, err := marshal()
	assert.NilError(t, err)
	err = unmarshal(bz)
	assert.NilError(t, err)
	assert.DeepEqual(t, original, res)
}
