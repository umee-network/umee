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

	median := Median(prices)
	require.Equal(sdk.MustNewDecFromStr("1.125"), median)
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

	medianDeviation := MedianDeviation(median, prices)
	require.Equal(sdk.MustNewDecFromStr("0.003125"), medianDeviation)
}

func TestAverage(t *testing.T) {
	require := require.New(t)
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.1"),
		sdk.MustNewDecFromStr("1.05"),
		sdk.MustNewDecFromStr("1.15"),
		sdk.MustNewDecFromStr("1.2"),
	}

	average := Average(prices)
	require.Equal(sdk.MustNewDecFromStr("1.125"), average)
}

func TestMin(t *testing.T) {
	require := require.New(t)
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.1"),
		sdk.MustNewDecFromStr("1.05"),
		sdk.MustNewDecFromStr("1.15"),
		sdk.MustNewDecFromStr("1.2"),
	}

	min := Min(prices)
	require.Equal(sdk.MustNewDecFromStr("1.05"), min)
}

func TestMax(t *testing.T) {
	require := require.New(t)
	prices := []sdk.Dec{
		sdk.MustNewDecFromStr("1.1"),
		sdk.MustNewDecFromStr("1.05"),
		sdk.MustNewDecFromStr("1.15"),
		sdk.MustNewDecFromStr("1.2"),
	}

	max := Max(prices)
	require.Equal(sdk.MustNewDecFromStr("1.2"), max)
}
