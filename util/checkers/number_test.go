package checkers

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/umee-network/umee/v6/tests/tsdk"
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
	decMaxDiff := func(a, b, maxDiff float64) error {
		return DecMaxDiff(tsdk.DecF(a), tsdk.DecF(b), tsdk.DecF(maxDiff), "")
	}

	assert.NoError(decMaxDiff(1, 1, 0))
	assert.NoError(decMaxDiff(-1, -1, 0))
	assert.NoError(decMaxDiff(0, 0, 0))

	assert.NoError(decMaxDiff(0.0001, 0.0001, 0))
	assert.NoError(decMaxDiff(-0.0001, -0.0001, 0))

	assert.Error(decMaxDiff(1, -1, 0))
	assert.Error(decMaxDiff(1, -1, -1))
	assert.Error(decMaxDiff(-1, 1, 0))
	assert.Error(decMaxDiff(-1, 1, -1))

	assert.NoError(decMaxDiff(1, -1, 2))
	assert.NoError(decMaxDiff(1, -1, 3))
	assert.NoError(decMaxDiff(1, -1, 100))
	assert.NoError(decMaxDiff(1, -1, 60000000))

	assert.NoError(decMaxDiff(-1, 1, 2))
	assert.NoError(decMaxDiff(-1, 1, 3))
	assert.NoError(decMaxDiff(-1, 1, 100))
	assert.NoError(decMaxDiff(-1, 1, 60000000))
}

func TestDecInZeroOne(t *testing.T) {
	assert.NoError(t, DecInZeroOne(tsdk.DecF(0), "", true))
	assert.NoError(t, DecInZeroOne(tsdk.DecF(0.01), "", true))
	assert.NoError(t, DecInZeroOne(tsdk.DecF(0.999), "", true))

	assert.NoError(t, DecInZeroOne(tsdk.DecF(0), "", false))
	assert.NoError(t, DecInZeroOne(tsdk.DecF(0.01), "", false))
	assert.NoError(t, DecInZeroOne(tsdk.DecF(0.999), "", false))

	assert.NoError(t, DecInZeroOne(tsdk.DecF(1), "", true))
	assert.Error(t, DecInZeroOne(tsdk.DecF(1), "", false))

	assert.Error(t, DecInZeroOne(tsdk.DecF(1.01), "", false))
	assert.Error(t, DecInZeroOne(tsdk.DecF(2), "", false))
	assert.Error(t, DecInZeroOne(tsdk.DecF(213812), "", false))
	assert.Error(t, DecInZeroOne(tsdk.DecF(-1), "", false))

	assert.Error(t, DecInZeroOne(tsdk.DecF(1.01), "", true))
	assert.Error(t, DecInZeroOne(tsdk.DecF(2), "", true))
	assert.Error(t, DecInZeroOne(tsdk.DecF(213812), "", true))
	assert.Error(t, DecInZeroOne(tsdk.DecF(-1), "", true))
}

func TestDecNotNegative(t *testing.T) {
	assert.NotNil(t, DecNotNegative(tsdk.DecF(-1), ""))
	assert.NotNil(t, DecNotNegative(sdk.Dec{}, ""))

	assert.Nil(t, DecNotNegative(sdk.ZeroDec(), ""))
	assert.Nil(t, DecNotNegative(tsdk.DecF(5), ""))
}
