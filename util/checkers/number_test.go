package checkers

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"

	"github.com/umee-network/umee/v6/tests/tsdk"
)

func TestNumberDiff(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	assert.NotNil(t, DecNotNegative(tsdk.DecF(-1), ""))
	assert.NotNil(t, DecNotNegative(sdkmath.LegacyDec{}, ""))

	assert.Nil(t, DecNotNegative(sdkmath.LegacyZeroDec(), ""))
	assert.Nil(t, DecNotNegative(tsdk.DecF(5), ""))
}

func TestNumPositive(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	assert.NoError(NumberPositive(1, ""))
	assert.NoError(NumberPositive(2, ""))
	assert.NoError(NumberPositive(9999999999999, ""))

	assert.Error(NumberPositive(0, ""))
	assert.Error(NumberPositive(-1, ""))
	assert.Error(NumberPositive(-2, ""))
	assert.Error(NumberPositive(-999999999999, ""))

	assert.NoError(BigNumPositive(tsdk.DecF(1.01), ""))
	assert.NoError(BigNumPositive(tsdk.DecF(0.000001), ""))
	assert.NoError(BigNumPositive(tsdk.DecF(0.123), ""))
	assert.NoError(BigNumPositive(tsdk.DecF(9999999999999999999), ""))

	assert.Error(BigNumPositive(tsdk.DecF(0), ""))
	assert.Error(BigNumPositive(tsdk.DecF(-0.01), ""))
	assert.Error(BigNumPositive(sdk.NewDec(0), ""))
	assert.Error(BigNumPositive(sdk.NewDec(-99999999999999999), ""))

	assert.NoError(BigNumPositive(sdk.NewInt(1), ""))
	assert.NoError(BigNumPositive(sdk.NewInt(2), ""))
	assert.NoError(BigNumPositive(sdk.NewInt(9), ""))
	n, ok := sdk.NewIntFromString("111111119999999999999999999")
	assert.True(ok)
	assert.NoError(BigNumPositive(n, ""))
}

func TestNumberMin(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	assert.NoError(NumberMin(-1, -10, ""))
	assert.NoError(NumberMin(1, 0, ""))
	assert.NoError(NumberMin(10, 2, ""))
	assert.NoError(NumberMin(999999999999999, 2, ""))
	assert.NoError(NumberMin(999999999999999, 999999999999998, ""))
	assert.NoError(NumberMin(0, 0, ""))
	assert.NoError(NumberMin(-10, -10, ""))
	assert.NoError(NumberMin(999999999999999, 999999999999999, ""))

	assert.Error(NumberMin(-10, -1, ""))
	assert.Error(NumberMin(-1, 10, ""))
	assert.Error(NumberMin(-10, 0, ""))
	assert.Error(NumberMin(-10, 10, ""))
	assert.Error(NumberMin(10, 11, ""))
}
