package util

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMedian(t *testing.T) {
	require := require.New(t)
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.1"),
		sdk.MustNewDecFromStr("1.05"),
		sdk.MustNewDecFromStr("1.15"),
		sdk.MustNewDecFromStr("1.2"),
	}

	median, err := Median(prices)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.125"), median)

	// test empty prices list
	median, err = Median([]sdk.Dec{})
	require.ErrorIs(err, ErrEmptyList)
}

func TestMedianDeviation(t *testing.T) {
	require := require.New(t)
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.1"),
		sdk.MustNewDecFromStr("1.05"),
		sdk.MustNewDecFromStr("1.15"),
		sdk.MustNewDecFromStr("1.2"),
	}
	median := sdk.MustNewDecFromStr("1.125")

	medianDeviation, err := MedianDeviation(median, prices)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("0.003125"), medianDeviation)

	// test empty prices list
	medianDeviation, err = MedianDeviation(median, []sdk.Dec{})
	require.ErrorIs(err, ErrEmptyList)
}

func TestAverage(t *testing.T) {
	require := require.New(t)
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.1"),
		sdk.MustNewDecFromStr("1.05"),
		sdk.MustNewDecFromStr("1.15"),
		sdk.MustNewDecFromStr("1.2"),
	}

	average, err := Average(prices)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.125"), average)

	// test empty prices list
	average, err = Average([]sdk.Dec{})
	require.ErrorIs(err, ErrEmptyList)
}

func TestMin(t *testing.T) {
	require := require.New(t)
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.1"),
		sdk.MustNewDecFromStr("1.05"),
		sdk.MustNewDecFromStr("1.15"),
		sdk.MustNewDecFromStr("1.2"),
	}

	min, err := Min(prices)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.05"), min)

	// test empty prices list
	min, err = Min([]sdk.Dec{})
	require.ErrorIs(err, ErrEmptyList)
}

func TestMax(t *testing.T) {
	require := require.New(t)
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.1"),
		sdk.MustNewDecFromStr("1.05"),
		sdk.MustNewDecFromStr("1.15"),
		sdk.MustNewDecFromStr("1.2"),
	}

	max, err := Max(prices)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.2"), max)

	// test empty prices list
	max, err = Max([]sdk.Dec{})
	require.ErrorIs(err, ErrEmptyList)
}
