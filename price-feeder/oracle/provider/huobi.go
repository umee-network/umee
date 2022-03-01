package provider

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	huobiHost          = "api-aws.huobi.pro"
	huobiPath          = "/ws"
	huobiReconnectTime = time.Minute * 2
)

var _ Provider = (*HuobiProvider)(nil)

type (
	// HuobiProvider defines an Oracle provider implemented by the Huobi public
	// API.
	//
	// REF: https://huobiapi.github.io/docs/spot/v1/en/#market-ticker
	HuobiProvider struct {
		wsURL           url.URL
		wsClient        *websocket.Conn
		logger          zerolog.Logger
		mtx             sync.Mutex
		tickers         map[string]HuobiTicker        // market.$symbol.ticker => HuobiTicker
		subscribedPairs map[string]types.CurrencyPair // Symbol => types.CurrencyPair
	}

	// HuobiTicker defines the response type for the channel and
	// the tick object for a given ticker/symbol.
	HuobiTicker struct {
		CH   string    `json:"ch"` // Data belonged channel，Format：market.$symbol.ticker
		Tick HuobiTick `json:"tick"`
	}

	// HuobiTick defines the response type for the last 24h market summary
	// and the last traded price for a given ticker/symbol.
	HuobiTick struct {
		Vol       float64 `json:"vol"`       // Accumulated trading value of last 24 hours
		LastPrice float64 `json:"lastPrice"` // Last traded price
	}

	// HuobiSubscriptionMsg Msg to subscribe to one ticker channel at time
	HuobiSubscriptionMsg struct {
		Sub string `json:"sub"` // channel to subscribe market.$symbol.ticker
	}
)

// NewHuobiProvider returns a new Huobi provider with the WS connection and msg handler.
func NewHuobiProvider(ctx context.Context, logger zerolog.Logger, pairs ...types.CurrencyPair) (*HuobiProvider, error) {
	wsURL := url.URL{
		Scheme: "wss",
		Host:   huobiHost,
		Path:   huobiPath,
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Huobi websocket: %w", err)
	}

	provider := &HuobiProvider{
		wsURL:           wsURL,
		wsClient:        wsConn,
		logger:          logger.With().Str("provider", "huobi").Logger(),
		tickers:         map[string]HuobiTicker{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	if err := provider.SubscribeTickers(pairs...); err != nil {
		return nil, err
	}

	go provider.handleWebSocketMsgs(ctx)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the saved map.
func (p *HuobiProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
	tickerPrices := make(map[string]TickerPrice, len(pairs))

	for _, cp := range pairs {
		price, err := p.getTickerPrice(cp)
		if err != nil {
			return nil, err
		}
		tickerPrices[cp.String()] = price
	}

	return tickerPrices, nil
}

// SubscribeTickers subscribe to all currency pairs and
// add the new ones into the provider subscribed pairs.
func (p *HuobiProvider) SubscribeTickers(cps ...types.CurrencyPair) error {
	for _, cp := range cps {
		if err := p.subscribePair(cp); err != nil {
			return err
		}
	}

	p.setSubscribedPairs(cps...)
	return nil
}

func (p *HuobiProvider) handleWebSocketMsgs(ctx context.Context) {
	reconnectTicker := time.NewTicker(huobiReconnectTime)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(defaultReadNewWSMessage):
			messageType, bz, err := p.wsClient.ReadMessage()
			if err != nil {
				// If some error occurs, check if connection is alive
				// and continue to try to read the next message.
				p.logger.Err(err).Msg("failed to read message")
				if err := p.ping(); err != nil {
					p.logger.Err(err).Msg("failed to send ping")
					if err := p.reconnect(); err != nil {
						p.logger.Err(err).Msg("error reconnecting")
					}
				}
				continue
			}

			if len(bz) == 0 {
				continue
			}

			p.messageReceived(messageType, bz, reconnectTicker)

		case <-reconnectTicker.C:
			if err := p.reconnect(); err != nil {
				p.logger.Err(err).Msg("error reconnecting")
			}
		}
	}
}

// messageReceived handles the received data from the Huobi websocket.
// All return data of websocket Market APIs are compressed with
// GZIP so they need to be decompressed.
func (p *HuobiProvider) messageReceived(messageType int, bz []byte, reconnectTicker *time.Ticker) {
	if messageType != websocket.BinaryMessage {
		return
	}

	bz, err := decompressGzip(bz)
	if err != nil {
		p.logger.Err(err).Msg("failed to decompress gziped message")
		return
	}

	if bytes.Contains(bz, ping) {
		p.pong(bz, reconnectTicker)
		return
	}

	var tickerResp HuobiTicker
	if err := json.Unmarshal(bz, &tickerResp); err != nil {
		// sometimes it returns other messages which are not ticker responses
		p.logger.Err(err).Msg("failed to unmarshal message")
		return
	}

	if tickerResp.Tick.LastPrice == 0 {
		return
	}

	p.setTickerPair(tickerResp)
}

// pong return a heartbeat message when a "ping" is received
// and reset the recconnect ticker because the connection is alive
// After connected to Huobi's Websocket server,
// the server will send heartbeat periodically (5s interval).
// When client receives an heartbeat message, it should respond
// with a matching "pong" message which has the same integer in it, e.g.
// {"ping": 1492420473027} and the return should be
// {"pong": 1492420473027}
func (p *HuobiProvider) pong(bz []byte, reconnectTicker *time.Ticker) {
	reconnectTicker.Reset(huobiReconnectTime)
	var heartbeat struct {
		Ping uint64 `json:"ping"`
	}

	if err := json.Unmarshal(bz, &heartbeat); err != nil {
		p.logger.Err(err).Msg("could not unmarshal heartbeat")
		return
	}

	if err := p.wsClient.WriteJSON(struct {
		Pong uint64 `json:"pong"`
	}{Pong: heartbeat.Ping}); err != nil {
		p.logger.Err(err).Msg("could not send pong message back")
	}
}

// ping to check websocket connection
func (p *HuobiProvider) ping() error {
	return p.wsClient.WriteMessage(websocket.PingMessage, ping)
}

func (p *HuobiProvider) setTickerPair(ticker HuobiTicker) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.tickers[ticker.CH] = ticker
}

// reconnect closes the last WS connection and create a new one
func (p *HuobiProvider) reconnect() error {
	p.wsClient.Close()

	p.logger.Debug().Msg("reconnecting websocket")
	wsConn, _, err := websocket.DefaultDialer.Dial(p.wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("error reconnecting to Huobi websocket: %w", err)
	}
	p.wsClient = wsConn

	for _, cp := range p.subscribedPairs {
		if err := p.subscribePair(cp); err != nil {
			return err
		}
	}

	return nil
}

// subscribePair write the subscription msg to the provider.
func (p *HuobiProvider) subscribePair(cp types.CurrencyPair) error {
	huobiSubscriptionMsg := newHuobiSubscriptionMsg(cp)
	return p.wsClient.WriteJSON(huobiSubscriptionMsg)
}

func (p *HuobiProvider) getTickerPrice(cp types.CurrencyPair) (TickerPrice, error) {
	ticker, ok := p.tickers[currencyPairToHuobiPair(cp)]
	if !ok {
		return TickerPrice{}, fmt.Errorf("huobi provider failed to get ticker price for %s", cp.String())
	}

	return ticker.toTickerPrice()
}

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *HuobiProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

// decompressGzip uncompress gzip compressed messages
// All data returned from the websocket Market APIs is compressed
// with GZIP, so it needs to be unzipped.
func decompressGzip(bz []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(bz))
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(r)
}

func (ticker HuobiTicker) toTickerPrice() (TickerPrice, error) {
	return newTickerPrice(
		"Huobi",
		ticker.CH,
		strconv.FormatFloat(ticker.Tick.LastPrice, 'f', -1, 64),
		strconv.FormatFloat(ticker.Tick.Vol, 'f', -1, 64),
	)
}

// newHuobiSubscriptionMsg returns a new subscription Msg
func newHuobiSubscriptionMsg(cp types.CurrencyPair) HuobiSubscriptionMsg {
	return HuobiSubscriptionMsg{
		Sub: currencyPairToHuobiPair(cp),
	}
}

// currencyPairToHuobiPair returns the channel name in the Format：market.$symbol.ticker
func currencyPairToHuobiPair(cp types.CurrencyPair) string {
	return strings.ToLower("market." + cp.String() + ".ticker")
}
