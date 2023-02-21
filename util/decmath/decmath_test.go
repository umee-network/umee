package decmath

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestMedian(t *testing.T) {
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.12"),
		sdk.MustNewDecFromStr("1.07"),
		sdk.MustNewDecFromStr("1.11"),
		sdk.MustNewDecFromStr("1.2"),
	}

	median, err := Median(prices)
	assert.NilError(t, err)
	assert.DeepEqual(t, sdk.MustNewDecFromStr("1.115"), median)

	// test empty prices list
	median, err = Median([]sdk.Dec{})
	assert.ErrorIs(t, err, ErrEmptyList)
}

func TestMedianDeviation(t *testing.T) {
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.12"),
		sdk.MustNewDecFromStr("1.07"),
		sdk.MustNewDecFromStr("1.11"),
		sdk.MustNewDecFromStr("1.2"),
	}
	median := sdk.MustNewDecFromStr("1.115")

	medianDeviation, err := MedianDeviation(median, prices)
	assert.NilError(t, err)
	assert.DeepEqual(t, sdk.MustNewDecFromStr("0.048218253804964775"), medianDeviation)

	// test empty prices list
	_, err = MedianDeviation(median, []sdk.Dec{})
	assert.ErrorIs(t, err, ErrEmptyList)
}

func TestAverage(t *testing.T) {
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.12"),
		sdk.MustNewDecFromStr("1.07"),
		sdk.MustNewDecFromStr("1.11"),
		sdk.MustNewDecFromStr("1.2"),
	}

	average, err := Average(prices)
	assert.NilError(t, err)
	assert.DeepEqual(t, sdk.MustNewDecFromStr("1.125"), average)

	// test empty prices list
	_, err = Average([]sdk.Dec{})
	assert.ErrorIs(t, err, ErrEmptyList)
}

func TestMin(t *testing.T) {
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.12"),
		sdk.MustNewDecFromStr("1.07"),
		sdk.MustNewDecFromStr("1.11"),
		sdk.MustNewDecFromStr("1.2"),
	}

	min, err := Min(prices)
	assert.NilError(t, err)
	assert.DeepEqual(t, sdk.MustNewDecFromStr("1.07"), min)

	// test empty prices list
	_, err = Min([]sdk.Dec{})
	assert.ErrorIs(t, err, ErrEmptyList)
}

func TestMax(t *testing.T) {
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.12"),
		sdk.MustNewDecFromStr("1.07"),
		sdk.MustNewDecFromStr("1.11"),
		sdk.MustNewDecFromStr("1.2"),
	}

	max, err := Max(prices)
	assert.NilError(t, err)
	assert.DeepEqual(t, sdk.MustNewDecFromStr("1.2"), max)

	// test empty prices list
	_, err = Max([]sdk.Dec{})
	assert.ErrorIs(t, err, ErrEmptyList)
}
