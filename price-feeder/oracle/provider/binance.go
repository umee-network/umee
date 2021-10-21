package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	binanceBaseURL = "https://api.binance.com"
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
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
		Code   int    `json:"code"`
		Msg    string `json:"msg"`
	}
)

func NewBinanceProvider() *BinanceProvider {
	return &BinanceProvider{
		baseURL: binanceBaseURL,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func NewBinanceProviderWithTimeout(timeout time.Duration) *BinanceProvider {
	return &BinanceProvider{
		baseURL: binanceBaseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p BinanceProvider) GetTickerPrices(tickers ...string) (map[string]sdk.Dec, error) {
	tickerPrices := make(map[string]sdk.Dec, len(tickers))
	for _, t := range tickers {
		price, err := p.getTickerPrice(t)
		if err != nil {
			return nil, err
		}

		tickerPrices[t] = price
	}

	return tickerPrices, nil
}

func (p BinanceProvider) getTickerPrice(ticker string) (sdk.Dec, error) {
	path := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", p.baseURL, ticker)

	resp, err := p.client.Get(path)
	if err != nil {
		return sdk.Dec{}, fmt.Errorf("failed to make Binance request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return sdk.Dec{}, fmt.Errorf("failed to read Binance response body: %w", err)
	}

	var tickerResp BinanceTickerResponse
	if err := json.Unmarshal(bz, &tickerResp); err != nil {
		return sdk.Dec{}, fmt.Errorf("failed to unmarshal Binance response body: %w", err)
	}

	if tickerResp.Code != 0 {
		return sdk.Dec{}, fmt.Errorf(
			"received unexpected error from Binance response: %v (%d)",
			tickerResp.Msg, tickerResp.Code,
		)
	}

	if strings.ToUpper(tickerResp.Symbol) != strings.ToUpper(ticker) {
		return sdk.Dec{}, fmt.Errorf(
			"received unexpected symbol from Binance response; expected: %s, got: %s",
			ticker, tickerResp.Symbol,
		)
	}

	price, err := sdk.NewDecFromStr(tickerResp.Price)
	if err != nil {
		return sdk.Dec{}, fmt.Errorf("failed to parse Binance price (%ss) for %s", tickerResp.Price, ticker)
	}

	return price, nil
}
