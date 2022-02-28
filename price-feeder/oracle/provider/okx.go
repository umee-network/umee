package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	okxHost      = "ws.okx.com:8443"
	okxPath      = "/ws/v5/public"
	okxPingCheck = time.Second * 28 // should be < 30
)

var _ Provider = (*OkxProvider)(nil)

type (
	// OkxProvider defines an Oracle provider implemented by the Okx public
	// API.
	//
	// REF: https://www.okx.com/docs-v5/en/#websocket-api-public-channel-tickers-channel
	OkxProvider struct {
		wsURL           url.URL
		wsClient        *websocket.Conn
		logger          zerolog.Logger
		mu              sync.Mutex
		tickers         map[string]OkxTickerPair // InstId => OkxTickerPair
		candles         map[string]OkxCandlePair // InstId => 0kxCandlePair
		reconnectTimer  *time.Ticker
		subscribedPairs []types.CurrencyPair
	}

	// OkxTickerPair defines a ticker pair of Okx
	OkxTickerPair struct {
		InstId string `json:"instId"` // Instrument ID ex.: BTC-USDT
		Last   string `json:"last"`   // Last traded price ex.: 43508.9
		Vol24h string `json:"vol24h"` // 24h trading volume ex.: 11159.87127845
	}

	// OkxTickerResponse defines the response structure of a Okx ticker
	// request.
	OkxTickerResponse struct {
		Data []OkxTickerPair `json:"data"`
	}

	OkxCandlePair struct {
		Open   string `json:"o"`
		Close  string `json:"c"`
		Volume string `json:"vol"`
	}

	OkxCandleArg struct {
		InstID string `json:"instId"`
	}

	OkxCandleResponse struct {
		Data [][]string   `json:"data"`
		Arg  OkxCandleArg `json:"arg"`
	}

	// OkxSubscriptionTopic Topic with the ticker to be subscribed/unsubscribed
	OkxSubscriptionTopic struct {
		Channel string `json:"channel"` // Channel name ex.: tickers
		InstId  string `json:"instId"`  // Instrument ID ex.: BTC-USDT
	}

	// OkxSubscriptionMsg Message to subscribe/unsubscribe with N Topics
	OkxSubscriptionMsg struct {
		Op   string                 `json:"op"` // Operation ex.: subscribe
		Args []OkxSubscriptionTopic `json:"args"`
	}
)

// NewOkxProvider creates a new OkxProvider
func NewOkxProvider(ctx context.Context, logger zerolog.Logger, pairs ...types.CurrencyPair) (*OkxProvider, error) {
	wsURL := url.URL{
		Scheme: "wss",
		Host:   okxHost,
		Path:   okxPath,
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Okx websocket: %w", err)
	}

	provider := &OkxProvider{
		wsURL:           wsURL,
		wsClient:        wsConn,
		logger:          logger.With().Str("provider", "okx").Logger(),
		tickers:         map[string]OkxTickerPair{},
		candles:         map[string]OkxCandlePair{},
		reconnectTimer:  time.NewTicker(okxPingCheck),
		subscribedPairs: pairs,
	}
	provider.wsClient.SetPongHandler(provider.pongHandler)

	if err := provider.subscribeTickers(pairs...); err != nil {
		return nil, err
	}

	if err := provider.subscribeCandles(pairs...); err != nil {
		return nil, err
	}

	go provider.handleReceivedTickers(ctx)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the saved map
func (p *OkxProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
	tickerPrices := make(map[string]TickerPrice, len(pairs))

	for _, currencyPair := range pairs {
		price, err := p.getTickerPrice(currencyPair)
		if err != nil {
			return nil, err
		}

		tickerPrices[currencyPair.String()] = price
	}

	return tickerPrices, nil
}

func (p *OkxProvider) getTickerPrice(cp types.CurrencyPair) (TickerPrice, error) {
	instrumentId := getInstrumentId(cp)
	tickerPair, ok := p.tickers[instrumentId]
	if !ok {
		return TickerPrice{}, fmt.Errorf("okx provider failed to get ticker price for %s", instrumentId)
	}

	return tickerPair.toTickerPrice()
}

func (p *OkxProvider) handleReceivedTickers(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(defaultReadNewWSMessage):
			messageType, bz, err := p.wsClient.ReadMessage()
			if err != nil {
				// if some error occurs continue to try to read the next message
				p.logger.Err(err).Msg("could not read message")
				if err := p.ping(); err != nil {
					p.logger.Err(err).Msg("could not send ping")
				}
				continue
			}

			if len(bz) == 0 {
				continue
			}

			p.resetReconnectTimer()
			p.messageReceived(messageType, bz)

		case <-p.reconnectTimer.C: // reset by the pongHandler
			if err := p.reconnect(); err != nil {
				p.logger.Err(err).Msg("error reconnecting")
			}
		}
	}
}

func (p *OkxProvider) messageReceived(messageType int, bz []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var tickerResp OkxTickerResponse
	var candleResp OkxCandleResponse

	if err := json.Unmarshal(bz, &tickerResp); err != nil {
		if err := json.Unmarshal(bz, &candleResp); err != nil {
			// sometimes it returns other messages which are not tickerResponses
			p.logger.Err(err).Msg("could not unmarshal")
			return
		}
	}

	if len(candleResp.Data) > 0 {
		for _, candlePair := range candleResp.Data {
			p.setCandlePair(candlePair, candleResp.Arg.InstID)
		}
		return
	}

	for _, tickerPair := range tickerResp.Data {
		p.setTickerPair(tickerPair)
	}
}

func (p *OkxProvider) setTickerPair(tickerPair OkxTickerPair) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tickers[tickerPair.InstId] = tickerPair
}

func (p *OkxProvider) setCandlePair(pairData []string, instID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// the candlesticks channel uses an array of stringss
	p.candles[instID] = OkxCandlePair{
		Open:   pairData[1],
		Close:  pairData[4],
		Volume: pairData[5],
	}
}

// subscribeTickers subscribe to all currency pairs
func (p *OkxProvider) subscribeTickers(cps ...types.CurrencyPair) error {
	topics := make([]OkxSubscriptionTopic, len(cps))

	for i, cp := range cps {
		instId := getInstrumentId(cp)
		topics[i] = newOkxSubscriptionTopic(instId)
	}

	subsMsg := newOkxSubscriptionMsg(topics...)
	return p.wsClient.WriteJSON(subsMsg)
}

// subscribeCandles subscribe to all candles pairs
func (p *OkxProvider) subscribeCandles(cps ...types.CurrencyPair) error {
	topics := make([]OkxSubscriptionTopic, len(cps))

	for i, cp := range cps {
		instId := getInstrumentId(cp)
		topics[i] = newOkxCandleSubscriptionTopic(instId)
	}

	subsMsg := newOkxSubscriptionMsg(topics...)
	return p.wsClient.WriteJSON(subsMsg)
}

func (p *OkxProvider) resetReconnectTimer() {
	p.reconnectTimer.Reset(okxPingCheck)
}

// reconnect closes the last WS connection and creates a new one.
// If there’s a network problem, the system will automatically disable the connection.
// The connection will break automatically if the subscription is not established or
// data has not been pushed for more than 30 seconds.
// To keep the connection stable:
// 1. Set a timer of N seconds whenever a response message is received, where N is less than 30.
// 2. If the timer is triggered, which means that no new message is received within N seconds, send the String 'ping'.
// 3. Expect a 'pong' as a response. If the response message is not received within N seconds, please raise an error or reconnect.
func (p *OkxProvider) reconnect() error {
	p.wsClient.Close()

	p.logger.Debug().Msg("reconnecting websocket")
	wsConn, _, err := websocket.DefaultDialer.Dial(p.wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("error reconnecting to Okx websocket: %w", err)
	}
	wsConn.SetPongHandler(p.pongHandler)
	p.wsClient = wsConn

	return p.subscribeTickers(p.subscribedPairs...)
}

// ping to check websocket connection
func (p *OkxProvider) ping() error {
	return p.wsClient.WriteMessage(websocket.PingMessage, ping)
}

func (p *OkxProvider) pongHandler(appData string) error {
	p.resetReconnectTimer()
	return nil
}

func (ticker OkxTickerPair) toTickerPrice() (TickerPrice, error) {
	return newTickerPrice("Okx", ticker.InstId, ticker.Last, ticker.Vol24h)
}

// getInstrumentId returns the expected pair instrument ID for Okx ex.: BTC-USDT
func getInstrumentId(pair types.CurrencyPair) string {
	return pair.Base + "-" + pair.Quote
}

// newOkxSubscriptionTopic returns a new subscription topic
func newOkxSubscriptionTopic(instId string) OkxSubscriptionTopic {
	return OkxSubscriptionTopic{
		Channel: "tickers",
		InstId:  instId,
	}
}

// newOkxSubscriptionTopic returns a new subscription topic
func newOkxCandleSubscriptionTopic(instId string) OkxSubscriptionTopic {
	return OkxSubscriptionTopic{
		Channel: "candle15m",
		InstId:  instId,
	}
}

// newOkxSubscriptionMsg returns a new subscription Msg for Okx
func newOkxSubscriptionMsg(args ...OkxSubscriptionTopic) OkxSubscriptionMsg {
	return OkxSubscriptionMsg{
		Op:   "subscribe",
		Args: args,
	}
}
