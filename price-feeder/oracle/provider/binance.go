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
	binanceHost = "stream.binance.com:9443"
	binancePath = "/ws/umeestream"
)

var _ Provider = (*BinanceProvider)(nil)

type (
	// BinanceProvider defines an Oracle provider implemented by the Binance public
	// API.
	//
	// REF: https://binance-docs.github.io/apidocs/spot/en/#aggregate-trade-streams
	BinanceProvider struct {
		wsURL    url.URL
		wsClient *websocket.Conn
		logger   zerolog.Logger
		mu       sync.Mutex
		tickers  map[string]BinanceTicker // Symbol => BinanceTicker
	}

	// BinanceTicker defines the response structure of a Binance ticker
	// request. https://pkg.go.dev/encoding/json#Unmarshal
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

	BinanceSubscribeMsg struct {
		Method string   `json:"method"`
		Params []string `json:"params"`
		Id     uint16   `json:"id"`
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
		wsURL:    wsURL,
		wsClient: wsConn,
		logger:   logger.With().Str("module", "oracle").Logger(),
		tickers:  map[string]BinanceTicker{},
	}

	if err := provider.subscribeTickers(pairs...); err != nil {
		return nil, err
	}

	go provider.handleReceivedTickers(ctx)

	return provider, nil
}

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
		return TickerPrice{}, fmt.Errorf("failed to get %s", key)
	}

	return ticker.toTickerPrice()
}

func (p *BinanceProvider) messageReceived(messageType int, bz []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var tickerRespWS BinanceTicker
	if err := json.Unmarshal(bz, &tickerRespWS); err != nil {
		// sometimes it returns other messages which are not ticker responses
		p.logger.Err(err).Msg("Binance provider could not unmarshal")
		return
	}

	if len(tickerRespWS.LastPrice) == 0 {
		return
	}

	p.setTickerPair(tickerRespWS)
}

func (p *BinanceProvider) setTickerPair(ticker BinanceTicker) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tickers[ticker.Symbol] = ticker
}

func (ticker BinanceTicker) toTickerPrice() (TickerPrice, error) {
	return newTickerPrice("Binance", ticker.Symbol, ticker.LastPrice, ticker.Volume)
}

// subscribeTickers subscribe to all currency pairs
func (p *BinanceProvider) subscribeTickers(cps ...types.CurrencyPair) error {
	params := make([]string, len(cps))

	for i, cp := range cps {
		params[i] = strings.ToLower(cp.Base + cp.Quote + "@ticker")
	}

	subsMsg := newBinanceSubscriptionMsg(params...)
	return p.wsClient.WriteJSON(subsMsg)
}

func (p *BinanceProvider) handleReceivedTickers(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(defaultReadNewMessage):
			// time after to avoid asking for prices too frequently
			messageType, bz, err := p.wsClient.ReadMessage()
			if err != nil {
				// if some error occurs continue to try to read the next message
				p.logger.Err(err).Msg("Binance provider could not read message")
				continue
			}

			p.messageReceived(messageType, bz)
		}
	}
}

// newSubscriptionMsg returns a new subscription Msg
func newBinanceSubscriptionMsg(params ...string) BinanceSubscribeMsg {
	return BinanceSubscribeMsg{
		Method: "SUBSCRIBE",
		Params: params,
		Id:     1,
	}
}
