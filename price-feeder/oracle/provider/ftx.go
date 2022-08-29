package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	ftxRestURL         = "https://ftx.com/api"
	ftxMarketsEndpoint = "/markets"
	ftxCandleEndpoint  = "/candles"
	ftxTimeLayout      = "2006-01-02T15:04:05+00:00"
	// candleWindowLength is the amount of seconds between
	// each candle
	candleWindowLength = 15
)

var _ Provider = (*FTXProvider)(nil)

type (
	// FTXProvider defines an Oracle provider implemented by the FTX public
	// API.
	//
	// REF: https://docs.ftx.com/
	FTXProvider struct {
		baseURL string
		client  *http.Client
	}

	// FTXMarketsResponse is the response object used for
	// available exchange rates and tickers.
	FTXMarketsResponse struct {
		Success bool         `json:"success"`
		Markets []FTXMarkets `json:"result"`
	}
	FTXMarkets struct {
		Base   string  `json:"baseCurrency"`   // e.x. "BTC"
		Quote  string  `json:"quoteCurrency"`  // e.x. "USDT"
		Price  float64 `json:"price"`          // e.x. 10579.52
		Volume float64 `json:"quoteVolume24h"` // e.x. 28914.76
	}

	// FTXCandleResponse is the response object used for
	// candle information.
	FTXCandleResponse struct {
		Success bool        `json:"success"`
		Candle  []FTXCandle `json:"result"`
	}
	FTXCandle struct {
		Price     float64 `json:"close"`     // e.x. 11055.25
		Volume    float64 `json:"volume"`    // e.x. 464193.95725
		StartTime string  `json:"startTime"` // e.x. "2019-06-24T17:15:00+00:00"
	}
)

// parseTime parses a string such as "2022-08-29T20:23:00+00:00" into time.Time
func (c FTXCandle) parseTime() (time.Time, error) {
	t, err := time.Parse(ftxTimeLayout, c.StartTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse ftx timestamp")
	}
	return t, nil
}

func NewFTXProvider(endpoint Endpoint) *FTXProvider {
	if endpoint.Name == ProviderFTX {
		return &FTXProvider{
			baseURL: endpoint.Rest,
			client:  newDefaultHTTPClient(),
		}
	}
	return &FTXProvider{
		baseURL: ftxRestURL,
		client:  newDefaultHTTPClient(),
	}
}

// GetCandlePrices returns the tickerPrices based on the provided pairs.
func (p FTXProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	path := fmt.Sprintf("%s%s", p.baseURL, ftxMarketsEndpoint)

	resp, err := p.client.Get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to make FTX request: %w", err)
	}
	err = checkHTTPStatus(resp)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read FTX response body: %w", err)
	}

	var tokensResp FTXMarketsResponse
	if err := json.Unmarshal(bz, &tokensResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal FTX markets response body: %w", err)
	}

	baseDenomIdx := make(map[string]types.CurrencyPair)
	for _, cp := range pairs {
		baseDenomIdx[strings.ToUpper(cp.Base)] = cp
	}

	tickerPrices := make(map[string]types.TickerPrice, len(pairs))
	for _, tr := range tokensResp.Markets {
		symbol := strings.ToUpper(tr.Base)

		cp, ok := baseDenomIdx[symbol]
		if !ok {
			// skip tokens that are not requested
			continue
		}

		if _, ok := tickerPrices[symbol]; ok {
			return nil, fmt.Errorf("duplicate token found in FTX response: %s", symbol)
		}

		priceRaw := strconv.FormatFloat(float64(tr.Price), 'f', -1, 64)
		price, err := sdk.NewDecFromStr(priceRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to read FTX price (%f) for %s", tr.Price, symbol)
		}

		volumeRaw := strconv.FormatFloat(tr.Volume, 'f', -1, 64)
		volume, err := sdk.NewDecFromStr(volumeRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to read FTX volume (%f) for %s", tr.Volume, symbol)
		}

		tickerPrices[cp.String()] = types.TickerPrice{Price: price, Volume: volume}
	}

	for _, cp := range pairs {
		if _, ok := tickerPrices[cp.String()]; !ok {
			return nil, fmt.Errorf("missing exchange rate for %s", cp.String())
		}
	}

	return tickerPrices, nil
}

// GetCandlePrices returns the candlePrices based on the provided pairs.
func (p FTXProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candles := make(map[string][]types.CandlePrice)
	for _, pair := range pairs {
		if _, ok := candles[pair.Base]; !ok {
			candles[pair.String()] = []types.CandlePrice{}
		}

		startTime := time.Now().Add(providerCandlePeriod * -1).Unix()
		endTime := time.Now().Unix()
		path := fmt.Sprintf("%s%s/%s/%s%s?resolution=%d&start_time=%d&end_time=%d",
			p.baseURL,
			ftxMarketsEndpoint,
			pair.Base,
			pair.Quote,
			ftxCandleEndpoint,
			candleWindowLength,
			startTime,
			endTime,
		)

		resp, err := p.client.Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to make FTX candle request: %w", err)
		}
		err = checkHTTPStatus(resp)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		bz, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read FTX candle response body: %w", err)
		}

		var candlesResp FTXCandleResponse
		if err := json.Unmarshal(bz, &candlesResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal FTX response body: %w", err)
		}

		candlePrices := []types.CandlePrice{}
		for _, responseCandle := range candlesResp.Candle {
			// the ftx api does not provide the endtime for these candles,
			// so we have to calculate it
			candleStart, err := responseCandle.parseTime()
			if err != nil {
				return nil, err
			}
			candleEnd := candleStart.Add(candleWindowLength).Unix() * int64(time.Second/time.Millisecond)

			closeStr := fmt.Sprintf("%f", responseCandle.Price)
			volumeStr := fmt.Sprintf("%f", responseCandle.Volume)
			candlePrices = append(candlePrices, types.CandlePrice{
				Price:     sdk.MustNewDecFromStr(closeStr),
				Volume:    sdk.MustNewDecFromStr(volumeStr),
				TimeStamp: candleEnd,
			})
		}
		candles[pair.String()] = candlePrices
	}

	return candles, nil
}

// GetAvailablePairs return all available pairs symbol to susbscribe.
func (p FTXProvider) GetAvailablePairs() (map[string]struct{}, error) {
	path := fmt.Sprintf("%s%s", p.baseURL, ftxMarketsEndpoint)

	resp, err := p.client.Get(path)
	if err != nil {
		return nil, err
	}
	err = checkHTTPStatus(resp)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary FTXMarketsResponse
	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return nil, err
	}

	if !pairsSummary.Success {
		return nil, fmt.Errorf("ftx markets api returned with failure")
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary.Markets))
	for _, pair := range pairsSummary.Markets {
		cp := types.CurrencyPair{
			Base:  strings.ToUpper(pair.Base),
			Quote: strings.ToUpper(pair.Quote),
		}
		availablePairs[cp.String()] = struct{}{}
	}

	return availablePairs, nil
}

// SubscribeCurrencyPairs performs a no-op since ftx does not use websockets
func (p FTXProvider) SubscribeCurrencyPairs(pairs ...types.CurrencyPair) error {
	return nil
}
