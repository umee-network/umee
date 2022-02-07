package oracle

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
)

// ComputeVWAP computes the volume weighted average price for all price points
// for each ticker/exchange pair. The provided prices argument reflects a mapping
// of provider => {<base> => <TickerPrice>, ...}.
//
// Ref: https://en.wikipedia.org/wiki/Volume-weighted_average_price
func ComputeVWAP(prices map[string]map[string]provider.TickerPrice) (map[string]sdk.Dec, error) {
	var (
		vwap      = make(map[string]sdk.Dec)
		volumeSum = make(map[string]sdk.Dec)
	)

	for _, providerPrices := range prices {
		for base, tp := range providerPrices {
			if _, ok := vwap[base]; !ok {
				vwap[base] = sdk.ZeroDec()
			}
			if _, ok := volumeSum[base]; !ok {
				volumeSum[base] = sdk.ZeroDec()
			}

			// vwap[base] = Σ {P * V} for all TickerPrice
			vwap[base] = vwap[base].Add(tp.Price.Mul(tp.Volume))

			// track total volume for each base
			volumeSum[base] = volumeSum[base].Add(tp.Volume)
		}
	}

	// compute VWAP for each base by dividing the Σ {P * V} by Σ {V}
	for base, p := range vwap {
		if volumeSum[base] == sdk.ZeroDec() {
			return nil, fmt.Errorf("unable to divide by zero")
		}
		vwap[base] = p.Quo(volumeSum[base])
	}

	return vwap, nil
}

// StandardDeviation returns maps of the standard deviations and means of assets.
// Will skip calculating for an asset if there are less than 3 prices.
func StandardDeviation(
	prices map[string]map[string]provider.TickerPrice) (
	map[string]sdk.Dec, map[string]sdk.Dec, error,
) {
	var (
		deviations = make(map[string]sdk.Dec)
		means      = make(map[string]sdk.Dec)
		priceSlice = make(map[string][]sdk.Dec)
		priceSums  = make(map[string]sdk.Dec)
	)

	// Calculate sums, create price slice
	for _, providerPrices := range prices {
		for base, tp := range providerPrices {
			if _, ok := priceSums[base]; !ok {
				priceSums[base] = sdk.ZeroDec()
			}
			if _, ok := priceSlice[base]; !ok {
				priceSlice[base] = []sdk.Dec{}
			}

			priceSums[base] = priceSums[base].Add(tp.Price)
			priceSlice[base] = append(priceSlice[base], tp.Price)
		}
	}

	// Calculate standard deviations for each asset
	for base, sum := range priceSums {
		// Skip if asset does not have enough prices
		if len(priceSlice[base]) < 3 {
			continue
		}
		if _, ok := deviations[base]; !ok {
			deviations[base] = sdk.ZeroDec()
		}
		if _, ok := means[base]; !ok {
			means[base] = sdk.ZeroDec()
		}

		priceAmount := int64(len(priceSlice))
		means[base] = sum.QuoInt64(priceAmount)
		varianceSum := sdk.ZeroDec()

		for _, price := range priceSlice[base] {
			deviation := price.Sub(means[base])
			varianceSum = varianceSum.Add(deviation.Mul(deviation))
		}
		variance := varianceSum.QuoInt64(priceAmount)
		standardDeviation, err := variance.ApproxSqrt()
		if err != nil {
			return make(map[string]sdk.Dec), make(map[string]sdk.Dec), err
		}
		deviations[base] = standardDeviation
	}

	return deviations, means, nil
}
