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

	"github.com/umee-network/umee/v3/util/coin"
)

const (
	mexcWSHost   = "wbs.mexc.com"
	mexcWSPath   = "/raw/ws"
	mexcRestHost = "https://www.mexc.com"
	mexcRestPath = "/open/api/v2/market/ticker"
)

var _ Provider = (*MexcProvider)(nil)

type (
	// MexcProvider defines an Oracle provider implemented by the Mexc public
	// API.
	//
	// REF: https://mxcdevelop.github.io/apidocs/spot_v2_en/#ticker-information
	// REF: https://mxcdevelop.github.io/apidocs/spot_v2_en/#k-line
	// REF: https://mxcdevelop.github.io/apidocs/spot_v2_en/#overview
	MexcProvider struct {
		wsURL           url.URL
		wsClient        *websocket.Conn
		logger          zerolog.Logger
		mtx             sync.RWMutex
		endpoints       Endpoint
		tickers         map[string]types.TickerPrice   // Symbol => TickerPrice
		candles         map[string][]types.CandlePrice // Symbol => CandlePrice
		subscribedPairs map[string]types.CurrencyPair  // Symbol => types.CurrencyPair
	}

	// MexcTickerResponse is the ticker price response object.
	MexcTickerResponse struct {
		Symbol map[string]MexcTicker `json:"data"` // e.x. ATOM_USDT
	}
	MexcTicker struct {
		LastPrice float64 `json:"p"` // Last price ex.: 0.0025
		Volume    float64 `json:"v"` // Total traded base asset volume ex.: 1000
	}

	// MexcCandle is the candle websocket response object.
	MexcCandleResponse struct {
		Symbol   string     `json:"symbol"` // Symbol ex.: ATOM_USDT
		Metadata MexcCandle `json:"data"`   // Metadata for candle
	}
	MexcCandle struct {
		Close     float64 `json:"c"` // Price at close
		TimeStamp int64   `json:"t"` // Close time in unix epoch ex.: 1645756200000
		Volume    float64 `json:"v"` // Volume during period
	}

	// MexcCandleSubscription Msg to subscribe all the candle channels.
	MexcCandleSubscription struct {
		OP       string `json:"op"`       // kline
		Symbol   string `json:"symbol"`   // streams to subscribe ex.: atom_usdt
		Interval string `json:"interval"` // Min1、Min5、Min15、Min30
	}

	// MexcTickerSubscription Msg to subscribe all the ticker channels.
	MexcTickerSubscription struct {
		OP string `json:"op"` // kline
	}

	// MexcPairSummary defines the response structure for a Mexc pair
	// summary.
	MexcPairSummary struct {
		Symbol string `json:"symbol"`
	}
)

func NewMexcProvider(
	ctx context.Context,
	logger zerolog.Logger,
	endpoints Endpoint,
	pairs ...types.CurrencyPair,
) (*MexcProvider, error) {
	if (endpoints.Name) != ProviderMexc {
		endpoints = Endpoint{
			Name:      ProviderMexc,
			Rest:      mexcRestHost,
			Websocket: mexcWSHost,
		}
	}

	wsURL := url.URL{
		Scheme: "wss",
		Host:   endpoints.Websocket,
		Path:   mexcWSPath,
	}

	wsConn, resp, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error connecting to mexc websocket: %w", err)
	}

	provider := &MexcProvider{
		wsURL:           wsURL,
		wsClient:        wsConn,
		logger:          logger.With().Str("provider", "mexc").Logger(),
		endpoints:       endpoints,
		tickers:         map[string]types.TickerPrice{},
		candles:         map[string][]types.CandlePrice{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	if err := provider.SubscribeCurrencyPairs(pairs...); err != nil {
		return nil, err
	}

	go provider.handleWebSocketMsgs(ctx)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the provided pairs.
func (p *MexcProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	tickerPrices := make(map[string]types.TickerPrice, len(pairs))

	for _, cp := range pairs {
		key := currencyPairToMexcPair(cp)
		price, err := p.getTickerPrice(key)
		if err != nil {
			return nil, err
		}
		tickerPrices[cp.String()] = price
	}

	return tickerPrices, nil
}

// GetCandlePrices returns the candlePrices based on the provided pairs.
func (p *MexcProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candlePrices := make(map[string][]types.CandlePrice, len(pairs))

	for _, cp := range pairs {
		key := currencyPairToMexcPair(cp)
		prices, err := p.getCandlePrices(key)
		if err != nil {
			return nil, err
		}
		candlePrices[cp.String()] = prices
	}

	return candlePrices, nil
}

// SubscribeCurrencyPairs subscribe all currency pairs into ticker and candle channels.
func (p *MexcProvider) SubscribeCurrencyPairs(cps ...types.CurrencyPair) error {
	if len(cps) == 0 {
		return fmt.Errorf("currency pairs is empty")
	}

	if err := p.subscribeChannels(cps...); err != nil {
		return err
	}

	p.setSubscribedPairs(cps...)
	return nil
}

// subscribeChannels subscribe to the ticker and candle channels for all currency pairs.
func (p *MexcProvider) subscribeChannels(cps ...types.CurrencyPair) error {
	if err := p.subscribeTickers(cps...); err != nil {
		return err
	}

	return p.subscribeCandles(cps...)
}

// subscribeTickers subscribe to the ticker channel for all currency pairs.
func (p *MexcProvider) subscribeTickers(cps ...types.CurrencyPair) error {
	pairs := make([]string, len(cps))

	for i, cp := range cps {
		pairs[i] = currencyPairToMexcPair(cp)
	}

	return p.subscribePairs(pairs...)
}

// subscribeCandles subscribe to the candle channel for all currency pairs.
func (p *MexcProvider) subscribeCandles(cps ...types.CurrencyPair) error {
	pairs := make([]string, len(cps))

	for i, cp := range cps {
		pairs[i] = currencyPairToMexcPair(cp)
	}

	return p.subscribePairs(pairs...)
}

// subscribedPairsToSlice returns the map of subscribed pairs as a slice.
func (p *MexcProvider) subscribedPairsToSlice() []types.CurrencyPair {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	return types.MapPairsToSlice(p.subscribedPairs)
}

func (p *MexcProvider) getTickerPrice(key string) (types.TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	ticker, ok := p.tickers[key]
	if !ok {
		return types.TickerPrice{}, fmt.Errorf(
			types.ErrTickerNotFound.Error(),
			ProviderMexc,
			key,
		)
	}

	return ticker, nil
}

func (p *MexcProvider) getCandlePrices(key string) ([]types.CandlePrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	candles, ok := p.candles[key]
	if !ok {
		return []types.CandlePrice{}, fmt.Errorf(
			types.ErrCandleNotFound.Error(),
			ProviderMexc,
			key,
		)
	}

	return candles, nil
}

func (p *MexcProvider) messageReceived(messageType int, bz []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var (
		tickerResp MexcTickerResponse
		tickerErr  error
		candleResp MexcCandleResponse
		candleErr  error
	)

	tickerErr = json.Unmarshal(bz, &tickerResp)
	for _, cp := range p.subscribedPairs {
		mexcPair := currencyPairToMexcPair(cp)
		if tickerResp.Symbol[mexcPair].LastPrice != 0 {
			p.setTickerPair(
				mexcPair,
				tickerResp.Symbol[mexcPair],
			)
			telemetryWebsocketMessage(ProviderMexc, MessageTypeTicker)
			return
		}
	}

	candleErr = json.Unmarshal(bz, &candleResp)
	if candleResp.Metadata.Close != 0 {
		p.setCandlePair(candleResp)
		telemetryWebsocketMessage(ProviderMexc, MessageTypeCandle)
		return
	}

	if tickerErr != nil || candleErr != nil {
		p.logger.Error().
			Int("length", len(bz)).
			AnErr("ticker", tickerErr).
			AnErr("candle", candleErr).
			Msg("mexc: Error on receive message")
	}
}

func (p *MexcProvider) setTickerPair(symbol string, ticker MexcTicker) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	price, err := coin.NewDecFromFloat(ticker.LastPrice)
	if err != nil {
		p.logger.Warn().Err(err).Msg("mexc: failed to parse ticker price")
	}
	volume, err := coin.NewDecFromFloat(ticker.Volume)
	if err != nil {
		p.logger.Warn().Err(err).Msg("mexc: failed to parse ticker volume")
	}

	p.tickers[symbol] = types.TickerPrice{
		Price:  price,
		Volume: volume,
	}
}

func (p *MexcProvider) setCandlePair(candleResp MexcCandleResponse) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	close, err := coin.NewDecFromFloat(candleResp.Metadata.Close)
	if err != nil {
		p.logger.Warn().Err(err).Msg("mexc: failed to parse candle close")
	}
	volume, err := coin.NewDecFromFloat(candleResp.Metadata.Volume)
	if err != nil {
		p.logger.Warn().Err(err).Msg("mexc: failed to parse candle volume")
	}
	candle := types.CandlePrice{
		Price:  close,
		Volume: volume,
		// convert seconds -> milli
		TimeStamp: SecondsToMilli(candleResp.Metadata.TimeStamp),
	}

	staleTime := PastUnixTime(providerCandlePeriod)
	candleList := []types.CandlePrice{}
	candleList = append(candleList, candle)

	for _, c := range p.candles[candleResp.Symbol] {
		if staleTime < c.TimeStamp {
			candleList = append(candleList, c)
		}
	}

	p.candles[candleResp.Symbol] = candleList
}

func (p *MexcProvider) handleWebSocketMsgs(ctx context.Context) {
	reconnectTicker := time.NewTicker(defaultMaxConnectionTime)
	defer reconnectTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(defaultReadNewWSMessage):
			messageType, bz, err := p.wsClient.ReadMessage()
			if err != nil {
				// if some error occurs continue to try to read the next message.
				p.logger.Err(err).Msg("mexc: could not read message")
				continue
			}

			if len(bz) == 0 {
				continue
			}

			p.messageReceived(messageType, bz)

		case <-reconnectTicker.C:
			if err := p.reconnect(); err != nil {
				p.logger.Err(err).Msg("error reconnecting")
			}
		}
	}
}

// reconnect closes the last WS connection then create a new one and subscribes to
// all subscribed pairs in the ticker and candle pairs. If no ping is received
// within 1 minute, the connection will be disconnected. It is recommended to
// send a ping for 10-20 seconds
func (p *MexcProvider) reconnect() error {
	err := p.wsClient.Close()
	if err != nil {
		return types.ErrProviderConnection.Wrapf("error closing Mecx websocket %v", err)
	}

	p.logger.Debug().Msg("mexc: reconnecting websocket")

	wsConn, resp, err := websocket.DefaultDialer.Dial(p.wsURL.String(), nil)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("mexc: error reconnect to mexc websocket: %w", err)
	}
	p.wsClient = wsConn
	telemetryWebsocketReconnect(ProviderMexc)

	return p.subscribeChannels(p.subscribedPairsToSlice()...)
}

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *MexcProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

// subscribePairs write the subscription msg to the provider.
func (p *MexcProvider) subscribePairs(pairs ...string) error {
	for _, cp := range pairs {
		subsMsg := newMexcCandleSubscriptionMsg(cp)
		err := p.wsClient.WriteJSON(subsMsg)
		if err != nil {
			return err
		}
	}
	subsMsg := newMexcTickerSubscriptionMsg()
	return p.wsClient.WriteJSON(subsMsg)
}

// GetAvailablePairs returns all pairs to which the provider can subscribe.
// ex.: map["ATOMUSDT" => {}, "UMEEUSDC" => {}].
func (p *MexcProvider) GetAvailablePairs() (map[string]struct{}, error) {
	resp, err := http.Get(p.endpoints.Rest + mexcRestPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary []MexcPairSummary
	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return nil, err
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary))
	for _, pairName := range pairsSummary {
		availablePairs[strings.ToUpper(pairName.Symbol)] = struct{}{}
	}

	return availablePairs, nil
}

// currencyPairToMexcPair receives a currency pair and return mexc
// ticker symbol atomusdt@ticker.
func currencyPairToMexcPair(cp types.CurrencyPair) string {
	return strings.ToUpper(cp.Base + "_" + cp.Quote)
}

// newMexcCandleSubscriptionMsg returns a new candle subscription Msg.
func newMexcCandleSubscriptionMsg(param string) MexcCandleSubscription {
	return MexcCandleSubscription{
		OP:       "sub.kline",
		Symbol:   param,
		Interval: "Min1",
	}
}

// newMexcTickerSubscriptionMsg returns a new ticker subscription Msg.
func newMexcTickerSubscriptionMsg() MexcTickerSubscription {
	return MexcTickerSubscription{
		OP: "sub.overview",
	}
}
