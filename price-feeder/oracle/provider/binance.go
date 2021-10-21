package provider

import (
	"net/http"
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
	panic("IMPLEMENT ME!")
}
