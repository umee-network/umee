package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
	"github.com/umee-network/umee/price-feeder/telemetry"
)

// deviationThreshold defines how many 𝜎 a provider can be away from the mean
// without being considered faulty.
var deviationThreshold = sdk.MustNewDecFromStr("1.0")

// FilterTickerDeviations finds the standard deviations of the prices of
// all assets, and filters out any providers that are not within 2𝜎 of the mean.
func FilterTickerDeviations(
	logger zerolog.Logger,
	prices provider.AggregatedProviderPrices,
) (provider.AggregatedProviderPrices, error) {
	var (
		filteredPrices = make(provider.AggregatedProviderPrices)
		priceMap       = make(map[string]map[string]sdk.Dec)
	)

	for providerName, priceTickers := range prices {
		if _, ok := priceMap[providerName]; !ok {
			priceMap[providerName] = make(map[string]sdk.Dec)
		}
		for base, tp := range priceTickers {
			priceMap[providerName][base] = tp.Price
		}
	}

	deviations, means, err := StandardDeviation(priceMap)
	if err != nil {
		return nil, err
	}

	// accept any prices that are within 2𝜎, or for which we couldn't get 𝜎
	for providerName, priceTickers := range prices {
		for base, tp := range priceTickers {
			if _, ok := deviations[base]; !ok ||
				(tp.Price.GTE(means[base].Sub(deviations[base].Mul(deviationThreshold))) &&
					tp.Price.LTE(means[base].Add(deviations[base].Mul(deviationThreshold)))) {
				if _, ok := filteredPrices[providerName]; !ok {
					filteredPrices[providerName] = make(map[string]provider.TickerPrice)
				}

				filteredPrices[providerName][base] = tp
			} else {
				telemetry.IncrCounter(1, "failure", "provider", "type", "ticker")
				logger.Warn().
					Str("base", base).
					Str("provider", providerName).
					Str("price", tp.Price.String()).
					Msg("provider deviating from other prices")
			}
		}
	}

	return filteredPrices, nil
}

// FilterCandleDeviations finds the standard deviations of the tvwaps of
// all assets, and filters out any providers that are not within 2𝜎 of the mean.
func FilterCandleDeviations(
	logger zerolog.Logger,
	candles provider.AggregatedProviderCandles,
) (provider.AggregatedProviderCandles, error) {
	var (
		filteredCandles = make(provider.AggregatedProviderCandles)
		tvwaps          = make(map[string]map[string]sdk.Dec)
	)

	for providerName, priceCandles := range candles {
		candlePrices := make(provider.AggregatedProviderCandles)

		for base, cp := range priceCandles {
			if _, ok := candlePrices[providerName]; !ok {
				candlePrices[providerName] = make(map[string][]provider.CandlePrice)
			}

			candlePrices[providerName][base] = cp
		}

		tvwap, err := ComputeTVWAP(candlePrices)
		if err != nil {
			return nil, err
		}

		for base, asset := range tvwap {
			if _, ok := tvwaps[providerName]; !ok {
				tvwaps[providerName] = make(map[string]sdk.Dec)
			}

			tvwaps[providerName][base] = asset
		}
	}

	deviations, means, err := StandardDeviation(tvwaps)
	if err != nil {
		return nil, err
	}

	// accept any tvwaps that are within 2𝜎, or for which we couldn't get 𝜎
	for providerName, priceMap := range tvwaps {
		for base, price := range priceMap {
			if _, ok := deviations[base]; !ok ||
				(price.GTE(means[base].Sub(deviations[base].Mul(deviationThreshold))) &&
					price.LTE(means[base].Add(deviations[base].Mul(deviationThreshold)))) {
				if _, ok := filteredCandles[providerName]; !ok {
					filteredCandles[providerName] = make(map[string][]provider.CandlePrice)
				}

				filteredCandles[providerName][base] = candles[providerName][base]
			} else {
				telemetry.IncrCounter(1, "failure", "provider", "type", "candle")
				logger.Warn().
					Str("base", base).
					Str("provider", providerName).
					Str("price", price.String()).
					Msg("provider deviating from other candles")
			}
		}
	}

	return filteredCandles, nil
}
