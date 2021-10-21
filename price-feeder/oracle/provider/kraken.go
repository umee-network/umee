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
	krakenBaseURL = "https://api.kraken.com"
)

var _ Provider = (*KrakenProvider)(nil)

type (
	// KrakenProvider defines an Oracle provider implemented by the Kraken public
	// API.
	//
	// REF: https://docs.kraken.com/rest/
	KrakenProvider struct {
		baseURL string
		client  *http.Client
	}

	// KrakenTickerPair defines the structure returned from Kraken for a ticker query.
	//
	// Note, we only care about 'c', which is the last trade closed [<price>, <lot volume>].
	KrakenTickerPair struct {
		C []string `json:"c"`
	}

	// KrakenTickerResponse defines the response structure of a Kraken ticker request.
	// The response may contain one or more tickers.
	KrakenTickerResponse struct {
		Error  []interface{}
		Result map[string]KrakenTickerPair
	}
)

func NewKrakenProvider() *KrakenProvider {
	return &KrakenProvider{
		baseURL: krakenBaseURL,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func NewKrakenProviderWithTimeout(timeout time.Duration) *KrakenProvider {
	return &KrakenProvider{
		baseURL: krakenBaseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p KrakenProvider) GetTickerPrices(tickers ...string) (map[string]sdk.Dec, error) {
	path := fmt.Sprintf("%s/0/public/Ticker?pair=%s", p.baseURL, strings.Join(tickers, ","))

	resp, err := p.client.Get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to make Kraken request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Kraken response body: %w", err)
	}

	var tickerResp KrakenTickerResponse
	if err := json.Unmarshal(bz, &tickerResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kraken response body: %w", err)
	}

	if len(tickerResp.Error) != 0 {
		return nil, fmt.Errorf("received unexpected error from Kraken response: %v", tickerResp.Error)
	}

	if len(tickers) != len(tickerResp.Result) {
		return nil, fmt.Errorf(
			"received unexpected number of tickers; expected: %d, got: %d",
			len(tickers), len(tickerResp.Result),
		)
	}

	tickerPrices := make(map[string]sdk.Dec, len(tickers))
	for _, t := range tickers {
		// TODO: We may need to transform 't' prior to looking it up in the response
		// as Kraken may represent currencies differently.
		pair, ok := tickerResp.Result[t]
		if !ok {
			return nil, fmt.Errorf("failed to find ticker in Kraken response: %s", t)
		}

		closePrice, err := sdk.NewDecFromStr(pair.C[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse Kraken close price (%s) for %s", pair.C[0], t)
		}

		tickerPrices[t] = closePrice
	}

	return tickerPrices, nil
}
