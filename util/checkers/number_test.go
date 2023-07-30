package checkers

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/umee-network/umee/v5/tests/tsdk"
)

func TestNumberDiff(t *testing.T) {
	assert := assert.New(t)

	assert.NoError(IntegerMaxDiff(1, 1, 0, ""))
	assert.NoError(IntegerMaxDiff(-1, -1, 0, ""))
	assert.NoError(IntegerMaxDiff(0, 0, 0, ""))

	assert.Error(IntegerMaxDiff(1, -1, 0, ""))
	assert.Error(IntegerMaxDiff(1, -1, -1, ""))
	assert.Error(IntegerMaxDiff(-1, 1, 0, ""))
	assert.Error(IntegerMaxDiff(-1, 1, -1, ""))

	assert.NoError(IntegerMaxDiff(1, -1, 2, ""))
	assert.NoError(IntegerMaxDiff(1, -1, 3, ""))
	assert.NoError(IntegerMaxDiff(1, -1, 100, ""))
	assert.NoError(IntegerMaxDiff(1, -1, 60000000, ""))

	assert.NoError(IntegerMaxDiff(-1, 1, 2, ""))
	assert.NoError(IntegerMaxDiff(-1, 1, 3, ""))
	assert.NoError(IntegerMaxDiff(-1, 1, 100, ""))
	assert.NoError(IntegerMaxDiff(-1, 1, 60000000, ""))
}

func TestDecDiff(t *testing.T) {
	assert := assert.New(t)

	zero := sdk.ZeroDec()
	assert.NoError(DecMaxDiff(tsdk.DecF(1), tsdk.DecF(1), zero, ""))
	assert.NoError(DecMaxDiff(tsdk.DecF(-1), tsdk.DecF(-1), zero, ""))
	assert.NoError(DecMaxDiff(tsdk.DecF(0), tsdk.DecF(0), zero, ""))

	assert.NoError(DecMaxDiff(tsdk.DecF(0.0001), tsdk.DecF(0.0001), zero, ""))
	assert.NoError(DecMaxDiff(tsdk.DecF(-0.0001), tsdk.DecF(-0.0001), zero, ""))

	// assert.Error(DecMaxDiff(1, -1, 0, ""))
	// assert.Error(DecMaxDiff(1, -1, -1, ""))
	// assert.Error(DecMaxDiff(-1, 1, 0, ""))
	// assert.Error(DecMaxDiff(-1, 1, -1, ""))

	// assert.NoError(DecMaxDiff(1, -1, 2, ""))
	// assert.NoError(DecMaxDiff(1, -1, 3, ""))
	// assert.NoError(DecMaxDiff(1, -1, 100, ""))
	// assert.NoError(DecMaxDiff(1, -1, 60000000, ""))

	// assert.NoError(DecMaxDiff(-1, 1, 2, ""))
	// assert.NoError(DecMaxDiff(-1, 1, 3, ""))
	// assert.NoError(DecMaxDiff(-1, 1, 100, ""))
	// assert.NoError(DecMaxDiff(-1, 1, 60000000, ""))
}
