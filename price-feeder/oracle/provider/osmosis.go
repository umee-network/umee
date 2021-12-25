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
	osmosisBaseURL       = "https://api-osmosis.imperator.co"
	osmosisTokenEndpoint = "/tokens/v1"
)

var _ Provider = (*OsmosisProvider)(nil)

type (
	// OsmosisProvider defines an Oracle provider implemented by the Osmosis public
	// API.
	//
	// REF: https://api-osmosis.imperator.co/swagger/
	OsmosisProvider struct {
		baseURL string
		client  *http.Client
	}

	// OsmosisTokenResponse defines the response structure for an Osmosis token
	// request.
	OsmosisTokenResponse struct {
		Price  float64 `json:"price"`
		Symbol string  `json:"symbol"`
		Volume float64 `json:"volume_24h"`
	}
)

func NewOsmosisProvider() *OsmosisProvider {
	return &OsmosisProvider{
		baseURL: osmosisBaseURL,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func NewOsmosisProviderWithTimeout(timeout time.Duration) *OsmosisProvider {
	return &OsmosisProvider{
		baseURL: osmosisBaseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p OsmosisProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
	path := fmt.Sprintf("%s%s/all", p.baseURL, osmosisBaseURL)

	resp, err := p.client.Get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to make Osmosis request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Osmosis response body: %w", err)
	}

	var tokensResp []OsmosisTokenResponse
	if err := json.Unmarshal(bz, &tokensResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Osmosis response body: %w", err)
	}

	baseDenoms := make(map[string]struct{})
	for _, cp := range pairs {
		baseDenoms[strings.ToUpper(cp.Base)] = struct{}{}
	}

	tickerPrices := make(map[string]TickerPrice, len(pairs))
	for _, tr := range tokensResp {
		symbol := strings.ToUpper(tr.Symbol) // symbol == base in a currency pair
		if _, ok := baseDenoms[symbol]; !ok {
			// skip tokens that are not requested
			continue
		}

		if _, ok := tickerPrices[symbol]; ok {
			return nil, fmt.Errorf("duplicate token found in Osmosis response: %s", symbol)
		}

		priceRaw := fmt.Sprintf("%f", tr.Price)
		price, err := sdk.NewDecFromStr(priceRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to read Osmosis price (%s) for %s", priceRaw, symbol)
		}

		volumeRaw := fmt.Sprintf("%f", tr.Volume)
		volume, err := sdk.NewDecFromStr(volumeRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to read Osmosis volume (%s) for %s", volumeRaw, symbol)
		}

		tickerPrices[symbol] = TickerPrice{Price: price, Volume: volume}
	}

	return tickerPrices, nil
}
