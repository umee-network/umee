package util

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Median returns the median of a list of prices.
func Median(prices []sdk.Dec) sdk.Dec {
	lenPrices := len(prices)
	if lenPrices == 0 {
		return sdk.ZeroDec()
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].BigInt().
			Cmp(prices[j].BigInt()) < 0
	})

	if lenPrices%2 == 0 {
		return prices[lenPrices/2-1].
			Add(prices[lenPrices/2]).
			QuoInt64(2)
	}
	return prices[lenPrices/2]
}

// MedianDeviation returns the standard deviation around the
// median of a list of prices.
// MedianDeviation = ∑((price - median)^2 / len(prices))
func MedianDeviation(median sdk.Dec, prices []sdk.Dec) sdk.Dec {
	lenPrices := len(prices)
	medianDeviation := sdk.ZeroDec()

	for _, price := range prices {
		medianDeviation = medianDeviation.Add(price.
			Sub(median).Abs().Power(2).
			QuoInt64(int64(lenPrices)))
	}

	return medianDeviation
}

// Average returns the average value of a list of prices
func Average(prices []sdk.Dec) sdk.Dec {
	lenPrices := len(prices)
	sumPrices := sdk.ZeroDec()
	if lenPrices == 0 {
		return sdk.ZeroDec()
	}

	for _, price := range prices {
		sumPrices = sumPrices.Add(price)
	}

	return sumPrices.QuoInt64(int64(lenPrices))
}

// Max returns the max value of a list of prices
func Max(prices []sdk.Dec) sdk.Dec {
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].BigInt().
			Cmp(prices[j].BigInt()) < 0
	})

	return prices[len(prices)-1]
}

// Min returns the min value of a list of prices
func Min(prices []sdk.Dec) sdk.Dec {
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].BigInt().
			Cmp(prices[j].BigInt()) < 0
	})

	return prices[0]
}
