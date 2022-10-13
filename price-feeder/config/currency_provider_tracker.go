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
	coinGeckoTickersEndpoint = "/tickers"
	trackingPeriod           = time.Hour * 24
)

var CurrencyProviders = make(map[string][]string)

type (
	CurrencyProviderTracker struct {
		logger              zerolog.Logger
		pairs               []CurrencyPair
		coinIDSymbolMap     map[string]string
		CurrencyProviders   map[string][]string
		CurrencyProviderMin map[string]int
	}

	// REF: https://www.coingecko.com/en/api/documentation
	CoinGeckoCoinList struct {
		ID     string `json:"id"`
		Symbol string `json:"symbol"`
	}

	CoinGeckoCoinTickerResponse struct {
		Tickers []CoinGeckoCoinTicker `json:"tickers"`
	}
	CoinGeckoCoinTicker struct {
		Base   string              `json:"base"`
		Target string              `json:"target"`
		Market CoinGeckoCoinMarket `json:"market"`
	}
	CoinGeckoCoinMarket struct {
		Name string `json:"name"`
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

func (t *CurrencyProviderTracker) getCurrencyProviderMin() map[string]int {
	return t.CurrencyProviderMin
}

func (t *CurrencyProviderTracker) setCoinIDSymbolMap() error {
	// Get list of assets on coingecko to cross reference coin symbol to id.
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

func (t *CurrencyProviderTracker) setCurrencyProviders() error {
	for _, pair := range t.pairs {
		pairBaseID := t.coinIDSymbolMap[strings.ToLower(pair.Base)]
		resp, err := http.Get(coinGeckoRestHost + coinGeckoRestPath + pairBaseID + coinGeckoTickersEndpoint)
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

func (t *CurrencyProviderTracker) setCurrencyProviderMin() {
	for base, exchanges := range t.getCurrencyProviders() {
		if len(exchanges) < 3 {
			t.CurrencyProviderMin[base] = len(exchanges)
		} else {
			t.CurrencyProviderMin[base] = 3
		}
	}
}

func (t *CurrencyProviderTracker) trackCurrencyProviders(ctx context.Context) {
	for currency, providers := range t.CurrencyProviders {
		t.logger.Info().Msg(fmt.Sprintf("Providers supporting %s: %v", currency, providers))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(trackingPeriod):
			if err := t.setCurrencyProviders(); err != nil {
				t.logger.Error().Err(err).Msg("failed to set available providers for currencies")
			}

			t.setCurrencyProviderMin()

			for currency, providers := range t.CurrencyProviders {
				t.logger.Info().Msg(fmt.Sprintf("Providers supporting %s: %v", currency, providers))
			}
		}
	}
}
