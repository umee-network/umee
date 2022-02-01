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
	krakenBaseURL        = "https://api.kraken.com"
	krakenTickerEndpoint = "/0/public/Ticker"
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
	// Note, we only care about 'c', which is the last trade closed [<price>, <lot volume>]
	// and 'v', the volume.
	KrakenTickerPair struct {
		C []string `json:"c"`
		V []string `json:"v"`
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
		client:  newDefaultHttpClient(),
	}
}

func NewKrakenProviderWithTimeout(timeout time.Duration) *KrakenProvider {
	return &KrakenProvider{
		baseURL: krakenBaseURL,
		client:  newHttpClientWithTimeout(timeout),
	}
}

func (p KrakenProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
	tickers := make([]string, len(pairs))
	for i, cp := range pairs {
		tickers[i] = cp.String()
	}

	path := fmt.Sprintf("%s%s?pair=%s", p.baseURL, krakenTickerEndpoint, strings.Join(tickers, ","))

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

	tickerPrices := make(map[string]TickerPrice, len(tickers))
	for _, t := range tickers {
		// TODO: We may need to transform 't' prior to looking it up in the response
		// as Kraken may represent currencies differently.
		pair, ok := tickerResp.Result[t]
		if !ok {
			return nil, fmt.Errorf("failed to find ticker in Kraken response: %s", t)
		}

		price, err := sdk.NewDecFromStr(pair.C[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse Kraken price (%s) for %s", pair.C[0], t)
		}

		volume, err := sdk.NewDecFromStr(pair.V[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse Kraken volume (%s) for %s", pair.V[1], t)
		}

		tickerPrices[t] = TickerPrice{Price: price, Volume: volume}
	}

	return tickerPrices, nil
}
