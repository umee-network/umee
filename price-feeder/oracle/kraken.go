package oracle

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
	defaultTimeout = 10 * time.Second
	krakenBaseURL  = "https://api.kraken.com"
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

	// TickerPair defines the structure returned from Kraken for a ticker query.
	//
	// Note, we only care about 'c', which is the last trade
	// closed [<price>, <lot volume>].
	TickerPair struct {
		C []string `json:"c"`
	}

	// TickerResponse defines the response structure of a Kraken ticker request.
	// The response may contain one or more tickers.
	TickerResponse struct {
		Error  []interface{}
		Result map[string]TickerPair
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

	var tickerResp TickerResponse
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
		// TODO: We may need to transform 't' prior to lookin it up in the response
		// as Kraken may represent currencies differently.
		pair, ok := tickerResp.Result[t]
		if !ok {
			return nil, fmt.Errorf("failed to find ticker in Kraken response: %s", t)
		}

		closePrice, err := sdk.NewDecFromStr(pair.C[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse close price for %s: %s", t, pair.C[0])
		}

		tickerPrices[t] = closePrice
	}

	return tickerPrices, nil
}
