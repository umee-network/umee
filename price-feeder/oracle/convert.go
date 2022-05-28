package oracle

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

// getUSDBasedProviders retrieves which providers for an asset have a USD-based pair,
// given the asset and the map of providers to currency pairs.
func getUSDBasedProviders(asset string, providerPairs map[string][]types.CurrencyPair) (map[string]struct{}, error) {
	conversionProviders := make(map[string]struct{})

	for provider, pairs := range providerPairs {
		for _, pair := range pairs {
			if strings.ToUpper(pair.Quote) == config.DenomUSD && strings.ToUpper(pair.Base) == asset {
				conversionProviders[provider] = struct{}{}
			}
		}
	}
	if len(conversionProviders) == 0 {
		return nil, fmt.Errorf("no providers have a usd conversion for this asset")
	}

	return conversionProviders, nil
}

// ConvertCandlesToUSD converts any candles which are not quoted in USD
// to USD by other price feeds.
func convertCandlesToUSD(
	candles provider.AggregatedProviderCandles,
	providerPairs map[string][]types.CurrencyPair,
) (provider.AggregatedProviderCandles, error) {
	if len(candles) == 0 {
		return candles, nil
	}

	conversionRates := make(map[string]sdk.Dec)
	requiredConversions := make(map[string]types.CurrencyPair)

	for pairProviderName, pairs := range providerPairs {
		for _, pair := range pairs {
			if strings.ToUpper(pair.Quote) != config.DenomUSD {
				// Get valid providers and use them to generate a USD-based price for this asset.
				validProviders, err := getUSDBasedProviders(pair.Quote, providerPairs)
				if err != nil {
					return nil, err
				}

				// Find candles which we can use for conversion, and calculate the tvwap
				// to find the conversion rate.
				validCandleList := provider.AggregatedProviderCandles{}
				for providerName, candleSet := range candles {
					if _, ok := validProviders[providerName]; ok {
						for base, candle := range candleSet {
							if base == pair.Quote {
								if _, ok := validCandleList[providerName]; !ok {
									validCandleList[providerName] = make(map[string][]provider.CandlePrice)
								}

								validCandleList[providerName][base] = candle
							}
						}
					}
				}

				if len(validCandleList) == 0 {
					return nil, fmt.Errorf("there are no valid conversion rates for %s", pair.Quote)
				}
				tvwap, err := ComputeTVWAP(validCandleList)
				if err != nil {
					return nil, err
				}

				conversionRates[pair.Quote] = tvwap[pair.Quote]
				requiredConversions[pairProviderName] = pair
			}
		}
	}

	// Convert assets to USD.
	for provider, assetMap := range candles {
		for asset, assetCandles := range assetMap {
			if requiredConversions[provider].Base == asset {
				for i := range assetCandles {
					assetCandles[i].Price = assetCandles[i].Price.Mul(
						conversionRates[requiredConversions[provider].Quote],
					)
				}
			}
		}
	}

	return candles, nil
}

// convertTickersToUSD converts any tickers which are not quoted in USD to USD,
// using the conversion rates of other tickers.
func convertTickersToUSD(
	tickers provider.AggregatedProviderPrices,
	providerPairs map[string][]types.CurrencyPair,
) (provider.AggregatedProviderPrices, error) {
	if len(tickers) == 0 {
		return tickers, nil
	}

	conversionRates := make(map[string]sdk.Dec)
	requiredConversions := make(map[string]types.CurrencyPair)

	for pairProviderName, pairs := range providerPairs {
		for _, pair := range pairs {
			if strings.ToUpper(pair.Quote) != config.DenomUSD {
				// Get valid providers and use them to generate a USD-based price for this asset.
				validProviders, err := getUSDBasedProviders(pair.Quote, providerPairs)
				if err != nil {
					return nil, err
				}

				// Find valid candles, and then let's re-compute the tvwap.
				validTickerList := provider.AggregatedProviderPrices{}
				for providerName, candleSet := range tickers {
					// Find tickers which we can use for conversion, and calculate the vwap
					// to find the conversion rate.
					if _, ok := validProviders[providerName]; ok {
						for base, ticker := range candleSet {
							if base == pair.Quote {
								if _, ok := validTickerList[providerName]; !ok {
									validTickerList[providerName] = make(map[string]provider.TickerPrice)
								}

								validTickerList[providerName][base] = ticker
							}
						}
					}
				}

				if len(validTickerList) == 0 {
					return nil, fmt.Errorf("there are no valid conversion rates for %s", pair.Quote)
				}
				vwap, err := ComputeVWAP(validTickerList)
				if err != nil {
					return nil, err
				}

				conversionRates[pair.Quote] = vwap[pair.Quote]
				requiredConversions[pairProviderName] = pair
			}
		}
	}

	// Convert assets to USD.
	for providerName, assetMap := range tickers {
		for asset := range assetMap {
			if requiredConversions[providerName].Base == asset {
				assetMap[asset] = provider.TickerPrice{
					Price: assetMap[asset].Price.Mul(
						conversionRates[requiredConversions[providerName].Quote],
					),
					Volume: assetMap[asset].Volume,
				}
			}
		}
	}

	return tickers, nil
}
