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
	okxBaseURL        = "https://www.okx.com"
	okxTickerEndpoint = "/api/v5/market/ticker"
)

var _ Provider = (*OkxProvider)(nil)

type (
	// OkxProvider defines an Oracle provider implemented by the Okx public
	// API.
	//
	// REF: https://www.okx.com/docs-v5/en/#rest-api
	OkxProvider struct {
		baseURL string
		client  *http.Client
	}

	// OkxTickerPair defines a ticker pair of Okx
	OkxTickerPair struct {
		InstId string `json:"instId"` // Instrument ID ex.: BTC-USDT
		Last   string `json:"last"`   // Last traded price ex.: 43508.9
		Vol24h string `json:"vol24h"` // 24h trading volume ex.: 11159.87127845
	}

	// OkxTickerResponse defines the response structure of a Okx ticker
	// request.
	OkxTickerResponse struct {
		Data []OkxTickerPair `json:"data"`

		// Code and Msg are populated on failed requests
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
)

func NewOkxProvider() *OkxProvider {
	return &OkxProvider{
		baseURL: okxBaseURL,
		client:  newDefaultHTTPClient(),
	}
}

func NewOkxProviderWithTimeout(timeout time.Duration) *OkxProvider {
	return &OkxProvider{
		baseURL: okxBaseURL,
		client:  newHTTPClientWithTimeout(timeout),
	}
}

func (p OkxProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
	tickerPrices := make(map[string]TickerPrice, len(pairs))
	for _, cp := range pairs {
		price, err := p.getTickerPrice(getInstrumentId(cp))
		if err != nil {
			return nil, err
		}

		tickerPrices[cp.String()] = price
	}

	return tickerPrices, nil
}

func (p OkxProvider) getTickerPrice(ticker string) (TickerPrice, error) {
	// https://www.okx.com/api/v5/market/ticker?instId=BTC-USDT
	path := fmt.Sprintf("%s%s?instId=%s", p.baseURL, okxTickerEndpoint, ticker)

	resp, err := p.client.Get(path)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to make Okx request: %w", err)
	}
	defer resp.Body.Close()

	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to read Okx response body: %w", err)
	}

	var tickerResp OkxTickerResponse
	if err := json.Unmarshal(bz, &tickerResp); err != nil {
		fmt.Printf("\n Err %+v - %s", err, err.Error())
		return TickerPrice{}, fmt.Errorf("failed to unmarshal Okx response body: %w", err)
	}

	if tickerResp.Code != "0" {
		return TickerPrice{}, fmt.Errorf(
			"received unexpected error from Okx response: %v (%s)",
			tickerResp.Msg, tickerResp.Code,
		)
	}

	if len(tickerResp.Data) == 0 {
		return TickerPrice{}, fmt.Errorf(
			"ticker price not found from Okx response: %v (%s)",
			tickerResp.Msg, tickerResp.Code,
		)
	}

	tickerPair := tickerResp.Data[0]

	if !strings.EqualFold(tickerPair.InstId, ticker) {
		return TickerPrice{}, fmt.Errorf(
			"received unexpected symbol from Okx response; expected: %s, got: %s",
			ticker, tickerPair.InstId,
		)
	}

	price, err := sdk.NewDecFromStr(tickerPair.Last)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse Okx price (%s) for %s", tickerPair.Last, ticker)
	}

	volume, err := sdk.NewDecFromStr(tickerPair.Vol24h)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse Okx volume (%s) for %s", tickerPair.Vol24h, ticker)
	}

	return TickerPrice{Price: price, Volume: volume}, nil
}

// getInstrumentId returns the expected pair instrument ID for Okx ex.: BTC-USDT
func getInstrumentId(pair types.CurrencyPair) string {
	return pair.Base + "-" + pair.Quote
}
