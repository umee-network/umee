package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	binanceBaseURL        = "https://api.binance.com"
	binanceTickerEndpoint = "/api/v3/ticker/24hr"
)

var _ Provider = (*BinanceProvider)(nil)

type (
	// BinanceProvider defines an Oracle provider implemented by the Binance public
	// API.
	//
	// REF: https://github.com/binance/binance-spot-api-docs/blob/master/rest-api.md
	BinanceProvider struct {
		baseURL string
		client  *http.Client
	}

	// BinanceTickerResponse defines the response structure of a Binance ticker
	// request.
	BinanceTickerResponse struct {
		Symbol    string `json:"symbol"`
		LastPrice string `json:"lastPrice"`
		Volume    string `json:"volume"`

		// Code and Msg are populated on failed requests
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
)

func NewBinanceProvider() *BinanceProvider {
	return &BinanceProvider{
		baseURL: binanceBaseURL,
		client:  newDefaultHttpClient(),
	}
}

func NewBinanceProviderWithTimeout(timeout time.Duration) *BinanceProvider {
	return &BinanceProvider{
		baseURL: binanceBaseURL,
		client:  newHttpClientWithTimeout(timeout),
	}
}

func (p BinanceProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
	tickerPrices := make(map[string]TickerPrice, len(pairs))
	for _, cp := range pairs {
		price, err := p.getTickerPrice(cp.String())
		if err != nil {
			return nil, err
		}

		tickerPrices[cp.String()] = price
	}

	return tickerPrices, nil
}

func (p BinanceProvider) getTickerPrice(ticker string) (TickerPrice, error) {
	path := fmt.Sprintf("%s%s?symbol=%s", p.baseURL, binanceTickerEndpoint, ticker)

	resp, err := p.client.Get(path)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to make Binance request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to read Binance response body: %w", err)
	}

	var tickerResp BinanceTickerResponse
	if err := json.Unmarshal(bz, &tickerResp); err != nil {
		return TickerPrice{}, fmt.Errorf("failed to unmarshal Binance response body: %w", err)
	}

	if tickerResp.Code != 0 {
		return TickerPrice{}, fmt.Errorf(
			"received unexpected error from Binance response: %v (%d)",
			tickerResp.Msg, tickerResp.Code,
		)
	}

	if strings.ToUpper(tickerResp.Symbol) != strings.ToUpper(ticker) {
		return TickerPrice{}, fmt.Errorf(
			"received unexpected symbol from Binance response; expected: %s, got: %s",
			ticker, tickerResp.Symbol,
		)
	}

	price, err := sdk.NewDecFromStr(tickerResp.LastPrice)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse Binance price (%s) for %s", tickerResp.LastPrice, ticker)
	}

	volume, err := sdk.NewDecFromStr(tickerResp.Volume)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse Binance volume (%s) for %s", tickerResp.Volume, ticker)
	}

	return TickerPrice{Price: price, Volume: volume}, nil
}
