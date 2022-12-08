package decmath

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ErrEmptyList = fmt.Errorf("empty price list passed in")
)

// Median returns the median of a list of prices. Returns error
// if prices is empty list.
func Median(prices []sdk.Dec) (sdk.Dec, error) {
	lenPrices := len(prices)
	if lenPrices == 0 {
		return sdk.ZeroDec(), ErrEmptyList
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].BigInt().
			Cmp(prices[j].BigInt()) < 0
	})

	if lenPrices%2 == 0 {
		return prices[lenPrices/2-1].
			Add(prices[lenPrices/2]).
			QuoInt64(2), nil
	}
	return prices[lenPrices/2], nil
}

// MedianDeviation returns the standard deviation around the
// median of a list of prices. Returns error if prices is empty list.
// MedianDeviation = ∑((price - median)^2 / len(prices))
func MedianDeviation(median sdk.Dec, prices []sdk.Dec) (sdk.Dec, error) {
	medianDeviation := sdk.ZeroDec()
	lenPrices := len(prices)
	if lenPrices == 0 {
		return medianDeviation, ErrEmptyList
	}

	for _, p := range prices {
		medianDeviation = medianDeviation.Add(
			p.Sub(median).Abs().Power(2).QuoInt64(int64(lenPrices)))
	}

	return medianDeviation, nil
}

// Average returns the average value of a list of prices. Returns error
// if prices is empty list.
func Average(prices []sdk.Dec) (sdk.Dec, error) {
	lenPrices := len(prices)
	if lenPrices == 0 {
		return sdk.ZeroDec(), ErrEmptyList
	}

	sumPrices := sdk.ZeroDec()
	for _, price := range prices {
		sumPrices = sumPrices.Add(price)
	}

	return sumPrices.QuoInt64(int64(lenPrices)), nil
}

// Max returns the max value of a list of prices. Returns error
// if prices is empty list.
func Max(prices []sdk.Dec) (sdk.Dec, error) {
	lenPrices := len(prices)
	if lenPrices == 0 {
		return sdk.ZeroDec(), ErrEmptyList
	}

	max := prices[0]
	for _, price := range prices {
		if price.GT(max) {
			max = price
		}
	}

	return max, nil
}

// Min returns the min value of a list of prices. Returns error
// if prices is empty list.
func Min(prices []sdk.Dec) (sdk.Dec, error) {
	lenPrices := len(prices)
	if lenPrices == 0 {
		return sdk.ZeroDec(), ErrEmptyList
	}

	min := prices[0]
	for _, price := range prices {
		if price.LT(min) {
			min = price
		}
	}

	return min, nil
}
