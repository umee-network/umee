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
