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
	coinGeckoRestHost        = "https://api.coingecko.com/"
	coinGeckoRestPath        = "api/v3/coins/"
	coinGeckoListEndpoint    = "list"
	coinGeckoTickersEndpoint = "tickers"
	trackingPeriod           = time.Hour * 24
)

var CurrencyProviders = make(map[string][]string)

type (
	// CurrencyProviderTracker queries the CoinGecko API for all the exchanges that
	// support the currency pairs set in the price feeder config. It will poll the API
	// every 24 hours to log any new exchanges that were added for a given currency.
	//
	// REF: https://www.coingecko.com/en/api/documentation
	CurrencyProviderTracker struct {
		logger              zerolog.Logger
		pairs               []CurrencyPair
		coinIDSymbolMap     map[string]string    // ex: map["ATOM"] = "cosmos"
		CurrencyProviders   map[string][]string  // map of price feeder currencies and what exchanges support them
		CurrencyProviderMin map[string]int       // map of price feeder currencies and min required providers for them
	}

	// List of assets on CoinGecko and their corresponding id and symbol.
	CoinGeckoCoinList struct {
		ID     string `json:"id"`     // ex: "cosmos"
		Symbol string `json:"symbol"` // ex: "ATOM"
	}

	// CoinGecko ticker shows market data for a given currency pair including what
	// exchanges they're on.
	CoinGeckoCoinTickerResponse struct {
		Tickers []CoinGeckoCoinTicker `json:"tickers"`
	}
	CoinGeckoCoinTicker struct {
		Base   string              `json:"base"`   // CurrencyPair.Base
		Target string              `json:"target"` // CurrencyPair.Quote
		Market CoinGeckoCoinMarket `json:"market"`
	}
	CoinGeckoCoinMarket struct {
		Name string `json:"name"` // ex: Binance
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

	if err := currencyProviderTracker.setCurrencyProviders(); err != nil {
		return nil, err
	}

	currencyProviderTracker.setCurrencyProviderMin()

	go currencyProviderTracker.trackCurrencyProviders(ctx)

	return currencyProviderTracker, nil
}

func (t *CurrencyProviderTracker) getCurrencyProviders() map[string][]string {
	return t.CurrencyProviders
}

func (t *CurrencyProviderTracker) getCurrencyProviderMin(base string) int {
	return t.CurrencyProviderMin[base]
}

func (t *CurrencyProviderTracker) logCurrencyProviders() {
	for currency, providers := range t.CurrencyProviders {
		t.logger.Info().Msg(fmt.Sprintf("Providers supporting %s: %v", currency, providers))
	}
}

// setCoinIDSymbolMap gets list of assets on coingecko to cross reference coin symbol to id.
func (t *CurrencyProviderTracker) setCoinIDSymbolMap() error {
	resp, err := http.Get(coinGeckoRestHost + coinGeckoRestPath + coinGeckoListEndpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var listResponse []CoinGeckoCoinList
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return err
	}

	for _, coin := range listResponse {
		t.coinIDSymbolMap[coin.Symbol] = coin.ID
	}

	return nil
}

// setCurrencyProviders queries CoinGecko's tickers endpoint to get all the exchanges that
// support each price feeder currency pair and store it in the CurrencyProviders map
func (t *CurrencyProviderTracker) setCurrencyProviders() error {
	for _, pair := range t.pairs {
		pairBaseID := t.coinIDSymbolMap[strings.ToLower(pair.Base)]
		resp, err := http.Get(coinGeckoRestHost + coinGeckoRestPath + pairBaseID + "/" + coinGeckoTickersEndpoint)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var tickerResponse CoinGeckoCoinTickerResponse
		if err = json.NewDecoder(resp.Body).Decode(&tickerResponse); err != nil {
			return err
		}

		for _, ticker := range tickerResponse.Tickers {
			if ticker.Target == pair.Quote {
				t.CurrencyProviders[pair.Base] = append(t.CurrencyProviders[pair.Base], ticker.Market.Name)
			}
		}
	}

	return nil
}

// setCurrencyProviderMin will set the minimum amount of providers for each currency
// to the amount of exchanges that support them if its less than 3. Otherwise it is
// set to 3 providers.
func (t *CurrencyProviderTracker) setCurrencyProviderMin() {
	for base, exchanges := range t.getCurrencyProviders() {
		if len(exchanges) < 3 {
			t.CurrencyProviderMin[base] = len(exchanges)
		} else {
			t.CurrencyProviderMin[base] = 3
		}
	}
}

// trackCurrencyProviders resets CurrencyProviders map and logs out supported
// exchanges for each currency every 24 hours
func (t *CurrencyProviderTracker) trackCurrencyProviders(ctx context.Context) {
	t.logCurrencyProviders()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(trackingPeriod):
			if err := t.setCurrencyProviders(); err != nil {
				t.logger.Error().Err(err).Msg("failed to set available providers for currencies")
			}

			t.logCurrencyProviders()
		}
	}
}
