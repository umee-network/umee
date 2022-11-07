package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	binanceWSHost   = "stream.binance.com:9443"
	binanceWSPath   = "/ws/umeestream"
	binanceRestHost = "https://api1.binance.com"
	binanceRestPath = "/api/v3/ticker/price"
)

var _ Provider = (*BinanceProvider)(nil)

type (
	// BinanceProvider defines an Oracle provider implemented by the Binance public
	// API.
	//
	// REF: https://binance-docs.github.io/apidocs/spot/en/#individual-symbol-mini-ticker-stream
	// REF: https://binance-docs.github.io/apidocs/spot/en/#kline-candlestick-streams
	BinanceProvider struct {
		logger          zerolog.Logger
		mtx             sync.RWMutex
		endpoints       Endpoint
		tickers         map[string]BinanceTicker      // Symbol => BinanceTicker
		candles         map[string][]BinanceCandle    // Symbol => BinanceCandle
		subscribedPairs map[string]types.CurrencyPair // Symbol => types.CurrencyPair
	}

	// BinanceTicker ticker price response. https://pkg.go.dev/encoding/json#Unmarshal
	// Unmarshal matches incoming object keys to the keys used by Marshal (either the
	// struct field name or its tag), preferring an exact match but also accepting a
	// case-insensitive match. C field which is Statistics close time is not used, but
	// it avoids to implement specific UnmarshalJSON.
	BinanceTicker struct {
		Symbol    string `json:"s"` // Symbol ex.: BTCUSDT
		LastPrice string `json:"c"` // Last price ex.: 0.0025
		Volume    string `json:"v"` // Total traded base asset volume ex.: 1000
		C         uint64 `json:"C"` // Statistics close time
	}

	// BinanceCandleMetadata candle metadata used to compute tvwap price.
	BinanceCandleMetadata struct {
		Close     string `json:"c"` // Price at close
		TimeStamp int64  `json:"T"` // Close time in unix epoch ex.: 1645756200000
		Volume    string `json:"v"` // Volume during period
	}

	// BinanceCandle candle binance websocket channel "kline_1m" response.
	BinanceCandle struct {
		Symbol   string                `json:"s"` // Symbol ex.: BTCUSDT
		Metadata BinanceCandleMetadata `json:"k"` // Metadata for candle
	}

	// BinanceSubscribeMsg Msg to subscribe all the tickers channels.
	BinanceSubscriptionMsg struct {
		Method string   `json:"method"` // SUBSCRIBE/UNSUBSCRIBE
		Params []string `json:"params"` // streams to subscribe ex.: usdtatom@ticker
		ID     uint16   `json:"id"`     // identify messages going back and forth
	}

	// BinanceSubscriptionResp the response structure for a binance subscription response
	BinanceSubscriptionResp struct {
		Result string `json:"result"`
		ID     uint16 `json:"id"`
	}

	// BinancePairSummary defines the response structure for a Binance pair
	// summary.
	BinancePairSummary struct {
		Symbol string `json:"symbol"`
	}
)

func NewBinanceProvider(
	ctx context.Context,
	logger zerolog.Logger,
	endpoints Endpoint,
	pairs ...types.CurrencyPair,
) (*BinanceProvider, error) {
	if (endpoints.Name) != ProviderBinance {
		endpoints = Endpoint{
			Name:      ProviderBinance,
			Rest:      binanceRestHost,
			Websocket: binanceWSHost,
		}
	}

	wsURL := url.URL{
		Scheme: "wss",
		Host:   endpoints.Websocket,
		Path:   binanceWSPath,
	}

	binanceLogger := logger.With().Str("provider", string(ProviderBinance)).Logger()

	provider := &BinanceProvider{
		logger:          binanceLogger,
		endpoints:       endpoints,
		tickers:         map[string]BinanceTicker{},
		candles:         map[string][]BinanceCandle{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	provider.setSubscribedPairs(pairs...)

	controller := NewWebsocketController(
		ctx,
		ProviderBinance,
		wsURL,
		provider.getSubscriptionMsgs(),
		provider.messageReceived,
		time.Duration(0),
		websocket.PingMessage,
		binanceLogger,
	)
	go controller.Start()

	return provider, nil
}

func (p *BinanceProvider) getSubscriptionMsgs() []interface{} {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	subscriptionMsgs := make([]interface{}, 0, len(p.subscribedPairs)*2)
	for _, cp := range p.subscribedPairs {
		binanceTickerPair := currencyPairToBinanceTickerPair(cp)
		subscriptionMsgs = append(subscriptionMsgs, newBinanceSubscriptionMsg(binanceTickerPair))

		binanceCandlePair := currencyPairToBinanceCandlePair(cp)
		subscriptionMsgs = append(subscriptionMsgs, newBinanceSubscriptionMsg(binanceCandlePair))
	}
	return subscriptionMsgs
}

// GetTickerPrices returns the tickerPrices based on the provided pairs.
func (p *BinanceProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	tickerPrices := make(map[string]types.TickerPrice, len(pairs))

	for _, cp := range pairs {
		key := cp.String()
		price, err := p.getTickerPrice(key)
		if err != nil {
			return nil, err
		}
		tickerPrices[key] = price
	}

	return tickerPrices, nil
}

// GetCandlePrices returns the candlePrices based on the provided pairs.
func (p *BinanceProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candlePrices := make(map[string][]types.CandlePrice, len(pairs))

	for _, cp := range pairs {
		key := cp.String()
		prices, err := p.getCandlePrices(key)
		if err != nil {
			return nil, err
		}
		candlePrices[key] = prices
	}

	return candlePrices, nil
}

func (p *BinanceProvider) getTickerPrice(key string) (types.TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	ticker, ok := p.tickers[key]
	if !ok {
		return types.TickerPrice{}, fmt.Errorf("binance failed to get ticker price for %s", key)
	}

	return ticker.toTickerPrice()
}

func (p *BinanceProvider) getCandlePrices(key string) ([]types.CandlePrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	candles, ok := p.candles[key]
	if !ok {
		return []types.CandlePrice{}, fmt.Errorf("binance failed to get candle prices for %s", key)
	}

	candleList := []types.CandlePrice{}
	for _, candle := range candles {
		cp, err := candle.toCandlePrice()
		if err != nil {
			return []types.CandlePrice{}, err
		}
		candleList = append(candleList, cp)
	}
	return candleList, nil
}

func (p *BinanceProvider) messageReceived(messageType int, bz []byte) {
	var (
		tickerResp       BinanceTicker
		tickerErr        error
		candleResp       BinanceCandle
		candleErr        error
		subscribeResp    BinanceSubscriptionResp
		subscribeRespErr error
	)

	tickerErr = json.Unmarshal(bz, &tickerResp)
	if len(tickerResp.LastPrice) != 0 {
		p.setTickerPair(tickerResp)
		telemetryWebsocketMessage(ProviderBinance, MessageTypeTicker)
		return
	}

	candleErr = json.Unmarshal(bz, &candleResp)
	if len(candleResp.Metadata.Close) != 0 {
		p.setCandlePair(candleResp)
		telemetryWebsocketMessage(ProviderBinance, MessageTypeCandle)
		return
	}

	subscribeRespErr = json.Unmarshal(bz, &subscribeResp)
	if subscribeResp.ID == 1 {
		return
	}

	p.logger.Error().
		Int("length", len(bz)).
		AnErr("ticker", tickerErr).
		AnErr("candle", candleErr).
		AnErr("subscribeResp", subscribeRespErr).
		Msg("Error on receive message")
}

func (p *BinanceProvider) setTickerPair(ticker BinanceTicker) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.tickers[ticker.Symbol] = ticker
}

func (p *BinanceProvider) setCandlePair(candle BinanceCandle) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	staleTime := PastUnixTime(providerCandlePeriod)
	candleList := []BinanceCandle{}
	candleList = append(candleList, candle)

	for _, c := range p.candles[candle.Symbol] {
		if staleTime < c.Metadata.TimeStamp {
			candleList = append(candleList, c)
		}
	}
	p.candles[candle.Symbol] = candleList
}

func (ticker BinanceTicker) toTickerPrice() (types.TickerPrice, error) {
	return types.NewTickerPrice(string(ProviderBinance), ticker.Symbol, ticker.LastPrice, ticker.Volume)
}

func (candle BinanceCandle) toCandlePrice() (types.CandlePrice, error) {
	return types.NewCandlePrice(string(ProviderBinance), candle.Symbol, candle.Metadata.Close, candle.Metadata.Volume,
		candle.Metadata.TimeStamp)
}

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *BinanceProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

// GetAvailablePairs returns all pairs to which the provider can subscribe.
// ex.: map["ATOMUSDT" => {}, "UMEEUSDC" => {}].
func (p *BinanceProvider) GetAvailablePairs() (map[string]struct{}, error) {
	resp, err := http.Get(p.endpoints.Rest + binanceRestPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary []BinancePairSummary
	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return nil, err
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary))
	for _, pairName := range pairsSummary {
		availablePairs[strings.ToUpper(pairName.Symbol)] = struct{}{}
	}

	return availablePairs, nil
}

// currencyPairToBinanceTickerPair receives a currency pair and return binance
// ticker symbol atomusdt@ticker.
func currencyPairToBinanceTickerPair(cp types.CurrencyPair) string {
	return strings.ToLower(cp.String() + "@ticker")
}

// currencyPairToBinanceCandlePair receives a currency pair and return binance
// candle symbol atomusdt@kline_1m.
func currencyPairToBinanceCandlePair(cp types.CurrencyPair) string {
	return strings.ToLower(cp.String() + "@kline_1m")
}

// newBinanceSubscriptionMsg returns a new subscription Msg.
func newBinanceSubscriptionMsg(params ...string) BinanceSubscriptionMsg {
	return BinanceSubscriptionMsg{
		Method: "SUBSCRIBE",
		Params: params,
		ID:     1,
	}
}
