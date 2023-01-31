package provider

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/umee-network/umee/price-feeder/v2/oracle/types"
)

const (
	huobiWSHost        = "api-aws.huobi.pro"
	huobiWSPath        = "/ws"
	huobiReconnectTime = time.Minute * 2
	huobiRestHost      = "https://api.huobi.pro"
	huobiRestPath      = "/market/tickers"
)

var _ Provider = (*HuobiProvider)(nil)

type (
	// HuobiProvider defines an Oracle provider implemented by the Huobi public
	// API.
	//
	// REF: https://huobiapi.github.io/docs/spot/v1/en/#market-ticker
	// REF: https://huobiapi.github.io/docs/spot/v1/en/#get-klines-candles
	HuobiProvider struct {
		wsc             *WebsocketController
		logger          zerolog.Logger
		mtx             sync.RWMutex
		endpoints       Endpoint
		tickers         map[string]HuobiTicker        // market.$symbol.ticker => HuobiTicker
		candles         map[string][]HuobiCandle      // market.$symbol.kline.$period => HuobiCandle
		subscribedPairs map[string]types.CurrencyPair // Symbol => types.CurrencyPair
	}

	// HuobiTicker defines the response type for the channel and the tick object for a
	// given ticker/symbol.
	HuobiTicker struct {
		CH   string    `json:"ch"` // Channel name. Format：market.$symbol.ticker
		Tick HuobiTick `json:"tick"`
	}

	// HuobiTick defines the response type for the last 24h market summary and the last
	// traded price for a given ticker/symbol.
	HuobiTick struct {
		Vol       float64 `json:"vol"`       // Accumulated trading value of last 24 hours
		LastPrice float64 `json:"lastPrice"` // Last traded price
	}

	// HuobiCandle defines the response type for the channel and the tick object for a
	// given ticker/symbol.
	HuobiCandle struct {
		CH   string          `json:"ch"` // Channel name. Format：market.$symbol.kline.$period
		Tick HuobiCandleTick `json:"tick"`
	}

	// HuobiCandleTick defines the response type for the candle.
	HuobiCandleTick struct {
		Close     float64 `json:"close"` // Closing price during this period
		TimeStamp int64   `json:"id"`    // TimeStamp for this as an ID
		Volume    float64 `json:"vol"`   // Volume during this period
	}

	// HuobiSubscriptionMsg Msg to subscribe to one ticker channel at time.
	HuobiSubscriptionMsg struct {
		Sub string `json:"sub"` // channel to subscribe market.$symbol.ticker
	}

	// HuobiSubscriptionResp the response structure for a Huobi subscription response
	HuobiSubscriptionResp struct {
		Status string `json:"status"`
	}

	// HuobiPairsSummary defines the response structure for an Huobi pairs
	// summary.
	HuobiPairsSummary struct {
		Data []HuobiPairData `json:"data"`
	}

	// HuobiPairData defines the data response structure for an Huobi pair.
	HuobiPairData struct {
		Symbol string `json:"symbol"`
	}
)

// NewHuobiProvider returns a new Huobi provider with the WS connection and msg handler.
func NewHuobiProvider(
	ctx context.Context,
	logger zerolog.Logger,
	endpoints Endpoint,
	pairs ...types.CurrencyPair,
) (*HuobiProvider, error) {
	if endpoints.Name != ProviderHuobi {
		endpoints = Endpoint{
			Name:      ProviderHuobi,
			Rest:      huobiRestHost,
			Websocket: huobiWSHost,
		}
	}

	wsURL := url.URL{
		Scheme: "wss",
		Host:   endpoints.Websocket,
		Path:   huobiWSPath,
	}

	huobiLogger := logger.With().Str("provider", string(ProviderHuobi)).Logger()

	provider := &HuobiProvider{
		logger:          huobiLogger,
		endpoints:       endpoints,
		tickers:         map[string]HuobiTicker{},
		candles:         map[string][]HuobiCandle{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	provider.setSubscribedPairs(pairs...)

	provider.wsc = NewWebsocketController(
		ctx,
		ProviderHuobi,
		wsURL,
		provider.getSubscriptionMsgs(pairs...),
		provider.messageReceived,
		disabledPingDuration,
		websocket.PingMessage,
		huobiLogger,
	)
	go provider.wsc.Start()

	return provider, nil
}

func (p *HuobiProvider) getSubscriptionMsgs(cps ...types.CurrencyPair) []interface{} {
	subscriptionMsgs := make([]interface{}, 0, len(cps)*2)
	for _, cp := range cps {
		subscriptionMsgs = append(subscriptionMsgs, newHuobiTickerSubscriptionMsg(cp))
		subscriptionMsgs = append(subscriptionMsgs, newHuobiCandleSubscriptionMsg(cp))
	}
	return subscriptionMsgs
}

// SubscribeCurrencyPairs sends the new subscription messages to the websocket
// and adds them to the providers subscribedPairs array
func (p *HuobiProvider) SubscribeCurrencyPairs(cps ...types.CurrencyPair) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	newPairs := []types.CurrencyPair{}
	for _, cp := range cps {
		if _, ok := p.subscribedPairs[cp.String()]; !ok {
			newPairs = append(newPairs, cp)
		}
	}

	newSubscriptionMsgs := p.getSubscriptionMsgs(newPairs...)
	if err := p.wsc.AddSubscriptionMsgs(newSubscriptionMsgs); err != nil {
		return err
	}
	p.setSubscribedPairs(newPairs...)
	return nil
}

// GetTickerPrices returns the tickerPrices based on the provided pairs.
func (p *HuobiProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	tickerPrices := make(map[string]types.TickerPrice, len(pairs))

	tickerErrs := 0
	for _, cp := range pairs {
		price, err := p.getTickerPrice(cp)
		if err != nil {
			p.logger.Warn().Err(err)
			tickerErrs++
			continue
		}
		tickerPrices[cp.String()] = price
	}

	if tickerErrs == len(pairs) {
		return nil, fmt.Errorf(
			types.ErrNoTickers.Error(),
			p.endpoints.Name,
			pairs,
		)
	}
	return tickerPrices, nil
}

// GetCandlePrices returns the candlePrices based on the provided pairs.
func (p *HuobiProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candlePrices := make(map[string][]types.CandlePrice, len(pairs))

	candleErrs := 0
	for _, cp := range pairs {
		prices, err := p.getCandlePrices(cp)
		if err != nil {
			p.logger.Warn().Err(err)
			candleErrs++
			continue
		}
		candlePrices[cp.String()] = prices
	}

	if candleErrs == len(pairs) {
		return nil, fmt.Errorf(
			types.ErrNoCandles.Error(),
			p.endpoints.Name,
			pairs,
		)
	}
	return candlePrices, nil
}

// messageReceived handles the received data from the Huobi websocket. All return
// data of websocket Market APIs are compressed with GZIP so they need to be
// decompressed.
func (p *HuobiProvider) messageReceived(messageType int, bz []byte) {
	if messageType != websocket.BinaryMessage {
		return
	}

	bz, err := decompressGzip(bz)
	if err != nil {
		p.logger.Err(err).Msg("failed to decompress gzipped message")
		return
	}

	if bytes.Contains(bz, ping) {
		p.pong(bz)
		return
	}

	var (
		tickerResp    HuobiTicker
		tickerErr     error
		candleResp    HuobiCandle
		candleErr     error
		subscribeResp HuobiSubscriptionResp
	)

	// sometimes the message received is not a ticker or a candle response.
	tickerErr = json.Unmarshal(bz, &tickerResp)
	if tickerResp.Tick.LastPrice != 0 {
		p.setTickerPair(tickerResp)
		telemetryWebsocketMessage(ProviderHuobi, MessageTypeTicker)
		return
	}

	candleErr = json.Unmarshal(bz, &candleResp)
	if candleResp.Tick.Close != 0 {
		p.setCandlePair(candleResp)
		telemetryWebsocketMessage(ProviderHuobi, MessageTypeCandle)
		return
	}

	err = json.Unmarshal(bz, &subscribeResp)
	if subscribeResp.Status == "ok" {
		return
	}

	p.logger.Error().
		Int("length", len(bz)).
		AnErr("ticker", tickerErr).
		AnErr("candle", candleErr).
		AnErr("subscribeResp", err).
		Msg("Error on receive message")
}

// pong return a heartbeat message when a "ping" is received and reset the
// reconnect ticker because the connection is alive. After connected to Huobi's
// Websocket server, the server will send heartbeat periodically (5s interval).
// When client receives an heartbeat message, it should respond with a matching
// "pong" message which has the same integer in it, e.g. {"ping": 1492420473027}
// and then the return pong message should be {"pong": 1492420473027}.
func (p *HuobiProvider) pong(bz []byte) {
	var heartbeat struct {
		Ping uint64 `json:"ping"`
	}

	if err := json.Unmarshal(bz, &heartbeat); err != nil {
		p.logger.Err(err).Msg("could not unmarshal heartbeat")
		return
	}

	if err := p.wsc.SendJSON(struct {
		Pong uint64 `json:"pong"`
	}{Pong: heartbeat.Ping}); err != nil {
		p.logger.Err(err).Msg("could not send pong message back")
	}
}

func (p *HuobiProvider) setTickerPair(ticker HuobiTicker) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.tickers[ticker.CH] = ticker
}

func (p *HuobiProvider) setCandlePair(candle HuobiCandle) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	// convert huobi timestamp seconds -> milliseconds
	candle.Tick.TimeStamp = SecondsToMilli(candle.Tick.TimeStamp)
	staleTime := PastUnixTime(providerCandlePeriod)
	candleList := []HuobiCandle{}
	candleList = append(candleList, candle)

	for _, c := range p.candles[candle.CH] {
		if staleTime < c.Tick.TimeStamp {
			candleList = append(candleList, c)
		}
	}
	p.candles[candle.CH] = candleList
}

func (p *HuobiProvider) getTickerPrice(cp types.CurrencyPair) (types.TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	ticker, ok := p.tickers[currencyPairToHuobiTickerPair(cp)]
	if !ok {
		return types.TickerPrice{}, fmt.Errorf(
			types.ErrTickerNotFound.Error(),
			p.endpoints.Name,
			cp.String(),
		)
	}

	return ticker.toTickerPrice()
}

func (p *HuobiProvider) getCandlePrices(cp types.CurrencyPair) ([]types.CandlePrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	candles, ok := p.candles[currencyPairToHuobiCandlePair(cp)]
	if !ok {
		return []types.CandlePrice{}, fmt.Errorf(
			types.ErrCandleNotFound.Error(),
			p.endpoints.Name,
			cp.String(),
		)
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

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *HuobiProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

// GetAvailablePairs returns all pairs to which the provider can subscribe.
func (p *HuobiProvider) GetAvailablePairs() (map[string]struct{}, error) {
	resp, err := http.Get(p.endpoints.Rest + huobiRestPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary HuobiPairsSummary
	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return nil, err
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary.Data))
	for _, pair := range pairsSummary.Data {
		availablePairs[strings.ToUpper(pair.Symbol)] = struct{}{}
	}

	return availablePairs, nil
}

// decompressGzip uncompress gzip compressed messages. All data returned from the
// websocket Market APIs is compressed with GZIP, so it needs to be unzipped.
func decompressGzip(bz []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(bz))
	if err != nil {
		return nil, err
	}

	return io.ReadAll(r)
}

// toTickerPrice converts current HuobiTicker to TickerPrice.
func (ticker HuobiTicker) toTickerPrice() (types.TickerPrice, error) {
	return types.NewTickerPrice(
		string(ProviderHuobi),
		ticker.CH,
		strconv.FormatFloat(ticker.Tick.LastPrice, 'f', -1, 64),
		strconv.FormatFloat(ticker.Tick.Vol, 'f', -1, 64),
	)
}

func (candle HuobiCandle) toCandlePrice() (types.CandlePrice, error) {
	return types.NewCandlePrice(
		string(ProviderHuobi),
		candle.CH,
		strconv.FormatFloat(candle.Tick.Close, 'f', -1, 64),
		strconv.FormatFloat(candle.Tick.Volume, 'f', -1, 64),
		candle.Tick.TimeStamp,
	)
}

// newHuobiTickerSubscriptionMsg returns a new ticker subscription Msg.
func newHuobiTickerSubscriptionMsg(cp types.CurrencyPair) HuobiSubscriptionMsg {
	return HuobiSubscriptionMsg{
		Sub: currencyPairToHuobiTickerPair(cp),
	}
}

// currencyPairToHuobiTickerPair returns the channel name in the following format:
// "market.$symbol.ticker".
func currencyPairToHuobiTickerPair(cp types.CurrencyPair) string {
	return strings.ToLower("market." + cp.String() + ".ticker")
}

// newHuobiSubscriptionMsg returns a new candle subscription Msg.
func newHuobiCandleSubscriptionMsg(cp types.CurrencyPair) HuobiSubscriptionMsg {
	return HuobiSubscriptionMsg{
		Sub: currencyPairToHuobiCandlePair(cp),
	}
}

// currencyPairToHuobiCandlePair returns the channel name in the following format:
// "market.$symbol.line.$period".
func currencyPairToHuobiCandlePair(cp types.CurrencyPair) string {
	return strings.ToLower("market." + cp.String() + ".kline.1min")
}
