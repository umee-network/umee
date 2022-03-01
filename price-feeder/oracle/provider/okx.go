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
		reconnectTimer  *time.Ticker
		mtx             sync.Mutex
		tickers         map[string]OkxTickerPair      // InstId => OkxTickerPair
		subscribedPairs map[string]types.CurrencyPair // Symbol => types.CurrencyPair
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
		reconnectTimer:  time.NewTicker(okxPingCheck),
		subscribedPairs: map[string]types.CurrencyPair{},
	}
	provider.wsClient.SetPongHandler(provider.pongHandler)

	if err := provider.SubscribeTickers(pairs...); err != nil {
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

// SubscribeTickers subscribe to all currency pairs and
// add the new ones into the provider subscribed pairs.
func (p *OkxProvider) SubscribeTickers(cps ...types.CurrencyPair) error {
	topics := make([]OkxSubscriptionTopic, len(cps))

	for i, cp := range cps {
		topics[i] = newOkxSubscriptionTopic(currencyPairToOkxPair(cp))
	}

	if err := p.subscribePairs(topics...); err != nil {
		return err
	}

	p.setSubscribedPairs(cps...)
	return nil
}

func (p *OkxProvider) getTickerPrice(cp types.CurrencyPair) (TickerPrice, error) {
	instrumentId := currencyPairToOkxPair(cp)
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
	if err := json.Unmarshal(bz, &tickerResp); err != nil {
		// sometimes it returns other messages which are not tickerResponses
		p.logger.Err(err).Msg("could not unmarshal")
		return
	}

	for _, tickerPair := range tickerResp.Data {
		p.setTickerPair(tickerPair)
	}
}

func (p *OkxProvider) setTickerPair(tickerPair OkxTickerPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.tickers[tickerPair.InstId] = tickerPair
}

// subscribePairs write the subscription msg to the provider.
func (p *OkxProvider) subscribePairs(pairs ...OkxSubscriptionTopic) error {
	subsMsg := newOkxSubscriptionMsg(pairs...)
	return p.wsClient.WriteJSON(subsMsg)
}

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *OkxProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
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

	topics := make([]OkxSubscriptionTopic, len(p.subscribedPairs))
	iterator := 0
	for _, cp := range p.subscribedPairs {
		topics[iterator] = newOkxSubscriptionTopic(currencyPairToOkxPair(cp))
		iterator++
	}

	return p.subscribePairs(topics...)
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

// currencyPairToOkxPair returns the expected pair instrument ID for Okx ex.: BTC-USDT
func currencyPairToOkxPair(pair types.CurrencyPair) string {
	return pair.Base + "-" + pair.Quote
}

// newOkxSubscriptionTopic returns a new subscription topic
func newOkxSubscriptionTopic(instId string) OkxSubscriptionTopic {
	return OkxSubscriptionTopic{
		Channel: "tickers",
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
