package decmath

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

var ErrEmptyList = fmt.Errorf("empty price list passed in")

// Median returns the median of a list of sdkmath.LegacyDec. Returns error
// if ds is empty list.
func Median(ds []sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	if len(ds) == 0 {
		return sdkmath.LegacyZeroDec(), ErrEmptyList
	}

	sort.Slice(ds, func(i, j int) bool {
		return ds[i].BigInt().
			Cmp(ds[j].BigInt()) < 0
	})

	if len(ds)%2 == 0 {
		return ds[len(ds)/2-1].
			Add(ds[len(ds)/2]).
			QuoInt64(2), nil
	}
	return ds[len(ds)/2], nil
}

// MedianDeviation returns the standard deviation around the
// median of a list of sdkmath.LegacyDec. Returns error if ds is empty list.
// MedianDeviation = sqrt(∑((d - median)^2 / len(ds)))
func MedianDeviation(median sdkmath.LegacyDec, ds []sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	if len(ds) == 0 {
		return sdkmath.LegacyZeroDec(), ErrEmptyList
	}

	variance := sdkmath.LegacyZeroDec()
	for _, d := range ds {
		variance = variance.Add(
			d.Sub(median).Abs().Power(2).QuoInt64(int64(len(ds))))
	}

	medianDeviation, err := variance.ApproxSqrt()
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}

	return medianDeviation, nil
}

// Average returns the average value of a list of sdkmath.LegacyDec. Returns error
// if ds is empty list.
func Average(ds []sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	if len(ds) == 0 {
		return sdkmath.LegacyZeroDec(), ErrEmptyList
	}

	sumPrices := sdkmath.LegacyZeroDec()
	for _, d := range ds {
		sumPrices = sumPrices.Add(d)
	}

	return sumPrices.QuoInt64(int64(len(ds))), nil
}

// Max returns the max value of a list of sdkmath.LegacyDec. Returns error
// if ds is empty list.
func Max(ds []sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	if len(ds) == 0 {
		return sdkmath.LegacyZeroDec(), ErrEmptyList
	}

	max := ds[0]
	for _, d := range ds[1:] {
		if d.GT(max) {
			max = d
		}
	}

	return max, nil
}

// Min returns the min value of a list of sdkmath.LegacyDec. Returns error
// if ds is empty list.
func Min(ds []sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	if len(ds) == 0 {
		return sdkmath.LegacyZeroDec(), ErrEmptyList
	}

	min := ds[0]
	for _, d := range ds[1:] {
		if d.LT(min) {
			min = d
		}
	}

	return min, nil
}
