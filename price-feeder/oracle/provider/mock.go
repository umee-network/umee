package provider

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	// Google Sheets document containing mock exchange rates.
	//
	// Ref: https://docs.google.com/spreadsheets/d/1wdaXPwlTqWnwhco4KQGEK5M0tXjCRQ1X4D-dIEiI1Xc/edit#gid=0
	mockBaseURL = "https://docs.google.com/spreadsheets/d/e/2PACX-1vSuFLjDFs5ajCoVZ8wFXaJ4DV8MkAKBcX2BJzWkLjx9i-jN-IDclePrBByXm1It8jgaZJGvsglUQuZ6/pub?output=csv&gid=0"
)

var _ Provider = (*MockProvider)(nil)

type (
	// MockProvider defines a mocked exchange rate provider using a published
	// Google sheets document to fetch mocked/fake exchange rates.
	MockProvider struct {
		baseURL string
		client  *http.Client
	}
)

func NewMockProvider() *MockProvider {
	return &MockProvider{
		baseURL: mockBaseURL,
		client: &http.Client{
			Timeout: defaultTimeout,
			// the mock provider is the only one who allows redirect
			// because it gets the mocked prices from the google spreadsheet
		},
	}
}

func (p MockProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
	tickerPrices := make(map[string]TickerPrice, len(pairs))

	resp, err := p.client.Get(p.baseURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	csvReader := csv.NewReader(resp.Body)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	tickerMap := make(map[string]struct{})
	for _, cp := range pairs {
		tickerMap[strings.ToUpper(cp.String())] = struct{}{}
	}

	// Records are of the form [base, quote, price, volume] and we skip the first
	// record as that contains the header.
	for _, r := range records[1:] {
		ticker := strings.ToUpper(r[0] + r[1])
		if _, ok := tickerMap[ticker]; !ok {
			// skip records that are not requested
			continue
		}

		price, err := sdk.NewDecFromStr(r[2])
		if err != nil {
			return nil, fmt.Errorf("failed to read mock price (%s) for %s", r[2], ticker)
		}

		volume, err := sdk.NewDecFromStr(r[3])
		if err != nil {
			return nil, fmt.Errorf("failed to read mock volume (%s) for %s", r[3], ticker)
		}

		if _, ok := tickerPrices[ticker]; ok {
			return nil, fmt.Errorf("found duplicate ticker: %s", ticker)
		}

		tickerPrices[ticker] = TickerPrice{Price: price, Volume: volume}
	}

	for t := range tickerMap {
		if _, ok := tickerPrices[t]; !ok {
			return nil, fmt.Errorf("missing exchange rate for %s", t)
		}
	}

	return tickerPrices, nil
}
