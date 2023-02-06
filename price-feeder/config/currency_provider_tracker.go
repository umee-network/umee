package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	coinGeckoRestURL            = "https://api.coingecko.com/api/v3/coins"
	coinGeckoListEndpoint       = "list"
	coinGeckoTickersEndpoint    = "tickers"
	osmosisV2RestURL            = "https://api.osmo-api.network.umee.cc"
	osmosisV2AssetPairsEndpoint = "assetpairs"
	requestTimeout              = time.Second * 2
	trackingPeriod              = time.Hour * 24
)

type (
	// CurrencyProviderTracker queries the CoinGecko API and UMEE's osmosis-api for all
	// the exchanges that support the currency pairs set in the price feeder config. It
	// will poll the APIs every 24 hours to log any new exchanges that were added for a
	// given currency.
	//
	// REF: https://www.coingecko.com/en/api/documentation
	// REF: https://github.com/umee-network/osmosis-api
	CurrencyProviderTracker struct {
		logger              zerolog.Logger
		pairs               []CurrencyPair
		coinIDSymbolMap     map[string]string   // ex: map["ATOM"] = "cosmos"
		CurrencyProviders   map[string][]string // map of price feeder currencies and what exchanges support them
		CurrencyProviderMin map[string]int      // map of price feeder currencies and min required providers for them
	}

	// List of assets on CoinGecko and their corresponding id and symbol.
	coinList struct {
		ID     string `json:"id"`     // ex: "cosmos"
		Symbol string `json:"symbol"` // ex: "ATOM"
	}

	// CoinGecko ticker shows market data for a given currency pair including what
	// exchanges they're on.
	coinTickerResponse struct {
		Tickers []coinTicker `json:"tickers"`
	}
	coinTicker struct {
		Base   string     `json:"base"`   // CurrencyPair.Base
		Target string     `json:"target"` // CurrencyPair.Quote
		Market coinMarket `json:"market"`
	}
	coinMarket struct {
		Name string `json:"name"` // ex: Binance
	}

	// Response from the osmosis-api REST server.
	assetPair struct {
		Base  string `json:"base"`
		Quote string `json:"quote"`
	}
)

func NewCurrencyProviderTracker(
	ctx context.Context,
	logger zerolog.Logger,
	pairs ...CurrencyPair,
) (*CurrencyProviderTracker, error) {
	currencyProviderTracker := &CurrencyProviderTracker{
		logger:              logger,
		pairs:               pairs,
		coinIDSymbolMap:     map[string]string{},
		CurrencyProviders:   map[string][]string{},
		CurrencyProviderMin: map[string]int{},
	}

	if err := currencyProviderTracker.setCoinIDSymbolMap(); err != nil {
		return nil, err
	}

	osmosisAPIPairs, err := currencyProviderTracker.getOsmosisAPIPairs()
	if err != nil {
		return nil, err
	}

	if err := currencyProviderTracker.setCurrencyProviders(osmosisAPIPairs); err != nil {
		return nil, err
	}

	currencyProviderTracker.setCurrencyProviderMin()

	go currencyProviderTracker.trackCurrencyProviders(ctx)

	return currencyProviderTracker, nil
}

func (t *CurrencyProviderTracker) logCurrencyProviders() {
	for currency, providers := range t.CurrencyProviders {
		t.logger.Info().Msg(fmt.Sprintf("providers supporting %s: %v", currency, providers))
	}
}

// setCoinIDSymbolMap gets list of assets on CoinGecko to cross reference coin symbol to id.
func (t *CurrencyProviderTracker) setCoinIDSymbolMap() error {
	resp, err := http.Get(fmt.Sprintf("%s/%s", coinGeckoRestURL, coinGeckoListEndpoint))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var listResponse []coinList
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return err
	}

	for _, coin := range listResponse {
		if _, ok := t.coinIDSymbolMap[coin.Symbol]; !ok {
			t.coinIDSymbolMap[coin.Symbol] = coin.ID
		}
	}

	return nil
}

// getOsmosisAPIPairs queries the osmosis-api assetpairs endpoint to get the asset pairs
// supported by it.
func (t *CurrencyProviderTracker) getOsmosisAPIPairs() (map[string]string, error) {
	client := &http.Client{
		Timeout: requestTimeout,
	}
	osmosisAPIPairs := make(map[string]string)

	osmosisResp, err := client.Get(fmt.Sprintf("%s/%s", osmosisV2RestURL, osmosisV2AssetPairsEndpoint))
	if err != nil {
		return nil, err
	}
	defer osmosisResp.Body.Close()
	var assetPairsResponse []assetPair
	if err = json.NewDecoder(osmosisResp.Body).Decode(&assetPairsResponse); err != nil {
		return nil, err
	}

	for _, assetPair := range assetPairsResponse {
		osmosisAPIPairs[assetPair.Base] = assetPair.Quote
	}

	return osmosisAPIPairs, nil
}

// setCurrencyProviders queries CoinGecko's tickers endpoint to get all the exchanges
// that support each price feeder currency pair and store it in the CurrencyProviders map.
func (t *CurrencyProviderTracker) setCurrencyProviders(osmosisAPIPairs map[string]string) error {
	client := &http.Client{
		Timeout: requestTimeout,
	}
	for _, pair := range t.pairs {
		// check if its a pair supported by the osmosis api
		if osmosisAPIPairs[strings.ToUpper(pair.Base)] == strings.ToUpper(pair.Quote) {
			t.CurrencyProviders[pair.Base] = append(t.CurrencyProviders[pair.Base], "osmosisv2")
		}

		// check if CoinGecko API supports pair
		pairBaseID := t.coinIDSymbolMap[strings.ToLower(pair.Base)]
		coinGeckoResp, err := client.Get(fmt.Sprintf("%s/%s/%s", coinGeckoRestURL, pairBaseID, coinGeckoTickersEndpoint))
		if err != nil {
			t.logger.Error().Err(err).Msg(fmt.Sprintf("Failed to query coin gecko api tickers endpoint for %s", pair.Base))
		} else {
			defer coinGeckoResp.Body.Close()
			var tickerResponse coinTickerResponse
			if err = json.NewDecoder(coinGeckoResp.Body).Decode(&tickerResponse); err != nil {
				return err
			}

			for _, ticker := range tickerResponse.Tickers {
				if ticker.Target == pair.Quote {
					t.CurrencyProviders[pair.Base] = append(t.CurrencyProviders[pair.Base], ticker.Market.Name)
				}
			}

			if len(t.CurrencyProviders[pair.Base]) == 0 {
				t.logger.Warn().Msg(fmt.Sprintf("Coin gecko found 0 providers for %s", pair.Base))
			}
		}
	}

	return nil
}

// setCurrencyProviderMin will set the minimum amount of providers for each currency
// to the amount of exchanges that support them if it's less than 3. Otherwise it is
// set to 3 providers.
func (t *CurrencyProviderTracker) setCurrencyProviderMin() {
	for base, exchanges := range t.CurrencyProviders {
		if len(exchanges) < 3 && len(exchanges) > 0 { // can be 0 if request errored
			t.CurrencyProviderMin[base] = len(exchanges)
		} else {
			t.CurrencyProviderMin[base] = 3
		}
	}
}

// trackCurrencyProviders resets CurrencyProviders map and logs out supported
// exchanges for each currency every 24 hours.
func (t *CurrencyProviderTracker) trackCurrencyProviders(ctx context.Context) {
	t.logCurrencyProviders()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(trackingPeriod):
			osmosisAPIPairs, err := t.getOsmosisAPIPairs()
			if err != nil {
				t.logger.Error().Err(err).Msg("failed to query osmosis-api for available asset pairs")
			}
			if err := t.setCurrencyProviders(osmosisAPIPairs); err != nil {
				t.logger.Error().Err(err).Msg("failed to set available providers for currencies")
			}

			t.logCurrencyProviders()
		}
	}
}
