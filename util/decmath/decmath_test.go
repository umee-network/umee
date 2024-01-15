package decmath

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"gotest.tools/v3/assert"
)

func TestMedian(t *testing.T) {
	prices := []sdkmath.LegacyDec{
		sdkmath.LegacyMustNewDecFromStr("1.12"),
		sdkmath.LegacyMustNewDecFromStr("1.07"),
		sdkmath.LegacyMustNewDecFromStr("1.11"),
		sdkmath.LegacyMustNewDecFromStr("1.2"),
	}

	median, err := Median(prices)
	assert.NilError(t, err)
	assert.DeepEqual(t, sdkmath.LegacyMustNewDecFromStr("1.115"), median)

	// test empty prices list
	median, err = Median([]sdkmath.LegacyDec{})
	assert.ErrorIs(t, err, ErrEmptyList)
}

func TestMedianDeviation(t *testing.T) {
	prices := []sdkmath.LegacyDec{
		sdkmath.LegacyMustNewDecFromStr("1.12"),
		sdkmath.LegacyMustNewDecFromStr("1.07"),
		sdkmath.LegacyMustNewDecFromStr("1.11"),
		sdkmath.LegacyMustNewDecFromStr("1.2"),
	}
	median := sdkmath.LegacyMustNewDecFromStr("1.115")

	medianDeviation, err := MedianDeviation(median, prices)
	assert.NilError(t, err)
	assert.DeepEqual(t, sdkmath.LegacyMustNewDecFromStr("0.048218253804964775"), medianDeviation)

	// test empty prices list
	_, err = MedianDeviation(median, []sdkmath.LegacyDec{})
	assert.ErrorIs(t, err, ErrEmptyList)
}

func TestAverage(t *testing.T) {
	prices := []sdkmath.LegacyDec{
		sdkmath.LegacyMustNewDecFromStr("1.12"),
		sdkmath.LegacyMustNewDecFromStr("1.07"),
		sdkmath.LegacyMustNewDecFromStr("1.11"),
		sdkmath.LegacyMustNewDecFromStr("1.2"),
	}

	average, err := Average(prices)
	assert.NilError(t, err)
	assert.DeepEqual(t, sdkmath.LegacyMustNewDecFromStr("1.125"), average)

	// test empty prices list
	_, err = Average([]sdkmath.LegacyDec{})
	assert.ErrorIs(t, err, ErrEmptyList)
}

func TestMin(t *testing.T) {
	prices := []sdkmath.LegacyDec{
		sdkmath.LegacyMustNewDecFromStr("1.12"),
		sdkmath.LegacyMustNewDecFromStr("1.07"),
		sdkmath.LegacyMustNewDecFromStr("1.11"),
		sdkmath.LegacyMustNewDecFromStr("1.2"),
	}

	min, err := Min(prices)
	assert.NilError(t, err)
	assert.DeepEqual(t, sdkmath.LegacyMustNewDecFromStr("1.07"), min)

	// test empty prices list
	_, err = Min([]sdkmath.LegacyDec{})
	assert.ErrorIs(t, err, ErrEmptyList)
}

func TestMax(t *testing.T) {
	prices := []sdkmath.LegacyDec{
		sdkmath.LegacyMustNewDecFromStr("1.12"),
		sdkmath.LegacyMustNewDecFromStr("1.07"),
		sdkmath.LegacyMustNewDecFromStr("1.11"),
		sdkmath.LegacyMustNewDecFromStr("1.2"),
	}

	max, err := Max(prices)
	assert.NilError(t, err)
	assert.DeepEqual(t, sdkmath.LegacyMustNewDecFromStr("1.2"), max)

	// test empty prices list
	_, err = Max([]sdkmath.LegacyDec{})
	assert.ErrorIs(t, err, ErrEmptyList)
}
