package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	finRestURL               = "https://api.kujira.app"
	finPairsEndpoint         = "/api/coingecko/pairs"
	finTickersEndpoint       = "/api/coingecko/tickers"
	finCandlesEndpoint       = "/api/trades/candles"
	finCandleBinSizeMinutes  = 5
	finCandleWindowSizeHours = 240
)

var _ Provider = (*FinProvider)(nil)

type (
	FinProvider struct {
		baseURL string
		client  *http.Client
	}

	// FinTicker is a single price ticker in the
	// FinTickers response object.
	FinTicker struct {
		Base   string `json:"base_currency"`
		Target string `json:"target_currency"`
		Symbol string `json:"ticker_id"`
		Price  string `json:"last_price"`
		Volume string `json:"base_volume"`
	}
	// FinTickers is the response object for the FIN
	// API.
	FinTickers struct {
		Tickers []FinTicker `json:"tickers"`
	}

	// FinCandle is a single candle in the
	// FinCandles response object.
	FinCandle struct {
		Bin    string `json:"bin"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}
	// FinCandles is the response object for the FIN
	// API.
	FinCandles struct {
		Candles []FinCandle `json:"candles"`
	}

	// FinPair is a single set of assets within the
	// FinPairs response object.
	FinPair struct {
		Base    string `json:"base"`
		Target  string `json:"target"`
		Symbol  string `json:"ticker_id"`
		Address string `json:"pool_id"`
	}
	// FinPairs is the response object for the FIN
	// API.
	FinPairs struct {
		Pairs []FinPair `json:"pairs"`
	}
)

// NewFinProvider returns a new instance of the FinProvider object.
func NewFinProvider(endpoint Endpoint) *FinProvider {
	if endpoint.Name == ProviderFin {
		return &FinProvider{
			baseURL: endpoint.Rest,
			client:  newDefaultHTTPClient(),
		}
	}
	return &FinProvider{
		baseURL: finRestURL,
		client:  newDefaultHTTPClient(),
	}
}

// GetTickerPrices calls the Fin API to get the set of prices in `pairs`.
func (p FinProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	path := fmt.Sprintf("%s%s", p.baseURL, finTickersEndpoint)

	tickerResponse, err := p.client.Get(path)
	if err != nil {
		return nil, fmt.Errorf("FIN tickers request failed: %w", err)
	}
	defer tickerResponse.Body.Close()
	bz, err := io.ReadAll(tickerResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("FIN tickers response read failed: %w", err)
	}
	var tickers FinTickers
	err = json.Unmarshal(bz, &tickers)
	if err != nil {
		return nil, fmt.Errorf("FIN tickers response unmarshal failed: %w", err)
	}

	tickerSymbolPairs := make(map[string]types.CurrencyPair, len(pairs))
	for _, pair := range pairs {
		tickerSymbolPairs[pair.Base+"_"+pair.Quote] = pair
	}
	tickerPrices := make(map[string]types.TickerPrice, len(pairs))
	for _, ticker := range tickers.Tickers {
		pair, ok := tickerSymbolPairs[strings.ToUpper(ticker.Symbol)]
		// skip tokens that are not requested
		if !ok {
			continue
		}
		_, ok = tickerPrices[pair.String()]
		if ok {
			return nil, fmt.Errorf("FIN tickers response contained duplicate: %s", ticker.Symbol)
		}
		price, err := strToDec(ticker.Price)
		if err != nil {
			return nil, fmt.Errorf("FIN ticker price failed to parse for %s: %s", ticker.Symbol, ticker.Price)
		}
		volume, err := strToDec(ticker.Volume)
		if err != nil {
			return nil, fmt.Errorf("FIN ticker volume failed to parse for %s: %s", ticker.Symbol, ticker.Volume)
		}
		tickerPrices[pair.String()] = types.TickerPrice{Price: price, Volume: volume}
	}
	for _, pair := range pairs {
		_, ok := tickerPrices[pair.String()]
		if !ok {
			return nil, fmt.Errorf("FIN ticker price missing for pair: %s", pair.String())
		}
	}
	return tickerPrices, nil
}

// GetCandlePrices calls the fin API and parses the responses to get
// candles from the API.
func (p FinProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	pairAddresses, err := p.getFinPairAddresses()
	if err != nil {
		return nil, fmt.Errorf("FIN pair addresses lookup failed: %w", err)
	}

	candlePricesPairs := make(map[string][]types.CandlePrice)
	for _, pair := range pairs {
		address, ok := pairAddresses[pair.String()]
		if !ok {
			return nil, fmt.Errorf("FIN contract address lookup failed for pair: %s", pair.String())
		}
		_, ok = candlePricesPairs[pair.Base]
		if !ok {
			candlePricesPairs[pair.String()] = []types.CandlePrice{}
		}

		windowEndTime := time.Now()
		windowStartTime := windowEndTime.Add(-finCandleWindowSizeHours * time.Hour)
		path := fmt.Sprintf("%s%s?contract=%s&precision=%d&from=%s&to=%s",
			p.baseURL,
			finCandlesEndpoint,
			address,
			finCandleBinSizeMinutes,
			windowStartTime.Format(time.RFC3339),
			windowEndTime.Format(time.RFC3339),
		)
		candlesResponse, err := p.client.Get(path)
		if err != nil {
			return nil, fmt.Errorf("FIN candles request failed: %w", err)
		}
		defer candlesResponse.Body.Close()
		bz, err := io.ReadAll(candlesResponse.Body)
		if err != nil {
			return nil, fmt.Errorf("FIN candles response read failed: %w", err)
		}
		var candles FinCandles
		err = json.Unmarshal(bz, &candles)
		if err != nil {
			return nil, fmt.Errorf("FIN candles response unmarshal failed: %w", err)
		}

		candlePrices := []types.CandlePrice{}
		for _, candle := range candles.Candles {
			timeStamp, err := binToTimeStamp(candle.Bin)
			if err != nil {
				return nil, fmt.Errorf("FIN candle timestamp failed to parse: %s", candle.Bin)
			}
			price, err := strToDec(candle.Close)
			if err != nil {
				return nil, fmt.Errorf("FIN provider unable to parse candle price")
			}
			volume, err := strToDec(candle.Volume)
			if err != nil {
				return nil, fmt.Errorf("FIN provider unable to parse candle volume")
			}

			candlePrices = append(candlePrices, types.CandlePrice{
				Price:     price,
				Volume:    volume,
				TimeStamp: timeStamp,
			})
		}
		candlePricesPairs[pair.String()] = candlePrices
	}

	return candlePricesPairs, nil
}

// GetAvailablePairs hits the FIN API to see how many listed fin pairs
// are supported, and parses them into a map of
// types.CurrencyPair.String() => struct{}
func (p FinProvider) GetAvailablePairs() (map[string]struct{}, error) {
	finPairs, err := p.getFinPairs()
	if err != nil {
		return nil, err
	}
	availablePairs := make(map[string]struct{}, len(finPairs.Pairs))
	for _, pair := range finPairs.Pairs {
		pair := types.CurrencyPair{
			Base:  strings.ToUpper(pair.Base),
			Quote: strings.ToUpper(pair.Target),
		}
		availablePairs[pair.String()] = struct{}{}
	}
	return availablePairs, nil
}

// getFinPairs calls the fin endpoint for all the listed fin pairs.
func (p FinProvider) getFinPairs() (FinPairs, error) {
	path := fmt.Sprintf("%s%s", p.baseURL, finPairsEndpoint)
	pairsResponse, err := p.client.Get(path)
	if err != nil {
		return FinPairs{}, err
	}
	defer pairsResponse.Body.Close()
	var pairs FinPairs
	err = json.NewDecoder(pairsResponse.Body).Decode(&pairs)
	if err != nil {
		return FinPairs{}, err
	}
	return pairs, nil
}

// getFinPairsAddresses gets the kujira contract addresses for
// currently supported assets.
// Ex. "ATOMAXLUSDC" => "kujira1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9..."
func (p FinProvider) getFinPairAddresses() (map[string]string, error) {
	finPairs, err := p.getFinPairs()
	if err != nil {
		return nil, err
	}

	pairAddresses := make(map[string]string, len(finPairs.Pairs))
	for _, pair := range finPairs.Pairs {
		pairAddresses[strings.ToUpper(pair.Base+pair.Target)] = pair.Address
	}
	return pairAddresses, nil
}

// SubscribeCurrencyPairs performs a no-op since fin does not use websockets
func (p FinProvider) SubscribeCurrencyPairs(pairs ...types.CurrencyPair) error {
	return nil
}

// strToDec converts a price or volume value from the fin API's response
// body, into an sdk.Dec type.
func strToDec(str string) (sdk.Dec, error) {
	if strings.Contains(str, ".") {
		split := strings.Split(str, ".")
		// sdk.NewDecFromStr will error if decimal precision is greater than 18
		if len(split[1]) > 18 {
			str = split[0] + "." + split[1][0:18]
		}
	}
	return sdk.NewDecFromStr(str)
}

// binToTimeStamp takes a candle's "BIN" and returns a unix timestamp.
func binToTimeStamp(bin string) (int64, error) {
	timeParsed, err := time.Parse(time.RFC3339, bin)
	if err != nil {
		return -1, err
	}
	return timeParsed.Unix(), nil
}
