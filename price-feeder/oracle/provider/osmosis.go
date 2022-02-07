package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
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
		client:  newDefaultHttpClient(),
	}
}

func NewOsmosisProviderWithTimeout(timeout time.Duration) *OsmosisProvider {
	return &OsmosisProvider{
		baseURL: osmosisBaseURL,
		client:  newHttpClientWithTimeout(timeout),
	}
}

func (p OsmosisProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
	path := fmt.Sprintf("%s%s/all", p.baseURL, osmosisTokenEndpoint)

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

	baseDenomIdx := make(map[string]types.CurrencyPair)
	for _, cp := range pairs {
		baseDenomIdx[strings.ToUpper(cp.Base)] = cp
	}

	tickerPrices := make(map[string]TickerPrice, len(pairs))
	for _, tr := range tokensResp {
		symbol := strings.ToUpper(tr.Symbol) // symbol == base in a currency pair

		cp, ok := baseDenomIdx[symbol]
		if !ok {
			// skip tokens that are not requested
			continue
		}

		if _, ok := tickerPrices[symbol]; ok {
			return nil, fmt.Errorf("duplicate token found in Osmosis response: %s", symbol)
		}

		priceRaw := strconv.FormatFloat(tr.Price, 'f', -1, 64)
		price, err := sdk.NewDecFromStr(priceRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to read Osmosis price (%s) for %s", priceRaw, symbol)
		}

		volumeRaw := strconv.FormatFloat(tr.Volume, 'f', -1, 64)
		volume, err := sdk.NewDecFromStr(volumeRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to read Osmosis volume (%s) for %s", volumeRaw, symbol)
		}

		tickerPrices[cp.String()] = TickerPrice{Price: price, Volume: volume}
	}

	for _, cp := range pairs {
		if _, ok := tickerPrices[cp.String()]; !ok {
			return nil, fmt.Errorf("missing exchange rate for %s", cp.String())
		}
	}

	return tickerPrices, nil
}
