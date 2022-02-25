package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	binanceHost           = "stream.binance.com:9443"
	binancePath           = "/ws/umeestream"
	binanceConnectionTime = time.Hour * 23 //  should be < 24
)

var _ Provider = (*BinanceProvider)(nil)

type (
	// BinanceProvider defines an Oracle provider implemented by the Binance public
	// API.
	//
	// REF: https://binance-docs.github.io/apidocs/spot/en/#individual-symbol-mini-ticker-stream
	// REF: https://binance-docs.github.io/apidocs/spot/en/#kline-candlestick-streams
	BinanceProvider struct {
		wsURL           url.URL
		wsClient        *websocket.Conn
		logger          zerolog.Logger
		mu              sync.Mutex
		tickers         map[string]BinanceTicker // Symbol => BinanceTicker
		candles         map[string]BinanceCandle // Symbol => BinanceCandle
		subscribedPairs []types.CurrencyPair
	}

	// BinanceTicker ticker price response
	// https://pkg.go.dev/encoding/json#Unmarshal
	// Unmarshal matches incoming object keys to the keys
	// used by Marshal (either the struct field name or its tag),
	// preferring an exact match but also accepting a case-insensitive match
	// C is not used, but it avoids to implement specific UnmarshalJSON
	BinanceTicker struct {
		Symbol    string `json:"s"` // Symbol ex.: BTCUSDT
		LastPrice string `json:"c"` // Last price ex.: 0.0025
		Volume    string `json:"v"` // Total traded base asset volume ex.: 1000
		C         uint64 `json:"C"` // Statistics close time
	}

	BinanceCandleMetadata struct {
		StartTime  int64  `json:"t"` // Start time in unix epoch ex.: 1645756200000
		CloseTime  int64  `json:"T"` // Close time in unix epoch ex.: 1645756200000
		Interval   string `json:"i"` // Interval ex.: 1m/15m/1h
		OpenPrice  string `json:"o"` // Price at open
		ClosePrice string `json:"c"` // Price at close
		Closed     string `json:"x"` // Whether or not this kline is closed
	}
	BinanceCandle struct {
		Symbol   string                `json:"s"` // Symbol ex.: BTCUSDT
		Metadata BinanceCandleMetadata `json:"k"` // Metadata for candle
	}

	// BinanceSubscribeMsg Msg to subscribe all the tickers channels
	BinanceSubscriptionMsg struct {
		Method string   `json:"method"` // SUBSCRIBE/UNSUBSCRIBE
		Params []string `json:"params"` // streams to subscribe ex.: usdtatom@ticker
		ID     uint16   `json:"id"`     // identify messages going back and forth
	}
)

func NewBinanceProvider(ctx context.Context, logger zerolog.Logger, pairs ...types.CurrencyPair) (*BinanceProvider, error) {
	wsURL := url.URL{
		Scheme: "wss",
		Host:   binanceHost,
		Path:   binancePath,
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Binance websocket: %w", err)
	}

	provider := &BinanceProvider{
		wsURL:           wsURL,
		wsClient:        wsConn,
		logger:          logger.With().Str("module", "oracle").Logger(),
		tickers:         map[string]BinanceTicker{},
		candles:         map[string]BinanceCandle{},
		subscribedPairs: pairs,
	}

	if err := provider.subscribeTickers(pairs...); err != nil {
		return nil, err
	}
	if err := provider.subscribeCandles(pairs...); err != nil {
		return nil, err
	}

	go provider.handleWebSocketMsgs(ctx)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the provided pairs.
func (p *BinanceProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
	tickerPrices := make(map[string]TickerPrice, len(pairs))

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

func (p *BinanceProvider) getTickerPrice(key string) (TickerPrice, error) {
	ticker, ok := p.tickers[key]
	if !ok {
		return TickerPrice{}, fmt.Errorf("binance provider failed to get ticker price for %s", key)
	}

	return ticker.toTickerPrice()
}

func (p *BinanceProvider) getTickerTrades(key string) (BinanceCandle, error) {
	candle, ok := p.candles[key]
	if !ok {
		return BinanceCandle{}, fmt.Errorf("failed to get ticker trades for %s", key)
	}

	return candle, nil
}

func (p *BinanceProvider) messageReceived(messageType int, bz []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var tickerResp BinanceTicker
	var candleResp BinanceCandle

	if err := json.Unmarshal(bz, &candleResp); err != nil {
		// sometimes it returns other messages which are not ticker responses
		if err := json.Unmarshal(bz, &tickerResp); err != nil {
			p.logger.Err(err).Msg("Binance provider could not unmarshal")
			return
		}
	}

	if len(tickerResp.LastPrice) != 0 {
		p.setTickerPair(tickerResp)
	}
	if len(candleResp.Metadata.ClosePrice) != 0 {
		p.setCandlePair(candleResp)
	}
}

func (p *BinanceProvider) setTickerPair(ticker BinanceTicker) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tickers[ticker.Symbol] = ticker
}

func (p *BinanceProvider) setCandlePair(candle BinanceCandle) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.candles[candle.Symbol] = candle
}

func (ticker BinanceTicker) toTickerPrice() (TickerPrice, error) {
	return newTickerPrice("Binance", ticker.Symbol, ticker.LastPrice, ticker.Volume)
}

// subscribeTickers subscribe to all currency pairs
func (p *BinanceProvider) subscribeTickers(cps ...types.CurrencyPair) error {
	params := make([]string, len(cps))

	for i, cp := range cps {
		params[i] = strings.ToLower(cp.String() + "@ticker")
	}

	subsMsg := newBinanceSubscriptionMsg(params...)
	return p.wsClient.WriteJSON(subsMsg)
}

// subscribeTickers subscribe to all candles
func (p *BinanceProvider) subscribeCandles(cps ...types.CurrencyPair) error {
	params := make([]string, len(cps))

	for i, cp := range cps {
		params[i] = strings.ToLower(cp.String() + "@kline_30m")
	}

	subsMsg := newBinanceSubscriptionMsg(params...)
	return p.wsClient.WriteJSON(subsMsg)
}

func (p *BinanceProvider) handleWebSocketMsgs(ctx context.Context) {
	reconnectTicker := time.NewTicker(defaultMaxConnectionTime)
	defer reconnectTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(defaultReadNewWSMessage):
			messageType, bz, err := p.wsClient.ReadMessage()
			if err != nil {
				// if some error occurs continue to try to read the next message
				p.logger.Err(err).Msg("Binance provider could not read message")
				continue
			}

			if len(bz) == 0 {
				continue
			}

			p.messageReceived(messageType, bz)

		case <-reconnectTicker.C:
			if err := p.reconnect(); err != nil {
				p.logger.Err(err).Msg("binance provider error reconnecting")
				p.keepReconnecting()
			}
		}
	}
}

// reconnect closes the last WS connection and create a new one
// A single connection to stream.binance.com is only valid for 24 hours;
// expect to be disconnected at the 24 hour mark
// The websocket server will send a ping frame every 3 minutes.
// If the websocket server does not receive a pong frame back from
// the connection within a 10 minute period, the connection will be disconnected.
// Unsolicited pong frames are allowed.
func (p *BinanceProvider) reconnect() error {
	p.wsClient.Close()

	p.logger.Debug().Msg("binance reconnecting websocket")
	wsConn, _, err := websocket.DefaultDialer.Dial(p.wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("error reconnect to binance websocket: %w", err)
	}
	p.wsClient = wsConn

	if err := p.subscribeCandles(p.subscribedPairs...); err != nil {
		return err
	}
	return p.subscribeTickers(p.subscribedPairs...)
}

// keepReconnecting keeps trying to reconnect if an error occurs in reconnect.
func (p *BinanceProvider) keepReconnecting() {
	reconnectTicker := time.NewTicker(defaultReconnectTime)
	defer reconnectTicker.Stop()
	connectionTries := 1

	for time := range reconnectTicker.C {
		if err := p.reconnect(); err != nil {
			p.logger.Err(err).Msgf("binance provider attempted to reconnect %d times at %s", connectionTries, time.String())
			continue
		}

		if connectionTries > maxReconnectionTries {
			p.logger.Warn().Msgf("binance provider failed to reconnect %d times", connectionTries)
		}
		connectionTries++
		return
	}
}

// newBinanceSubscriptionMsg returns a new subscription Msg
func newBinanceSubscriptionMsg(params ...string) BinanceSubscriptionMsg {
	return BinanceSubscriptionMsg{
		Method: "SUBSCRIBE",
		Params: params,
		ID:     1,
	}
}
