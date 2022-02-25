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
	krakenHost                    = "ws.kraken.com"
	krakenEventSystemStatus       = "systemStatus"
	krakenEventSubscriptionStatus = "subscriptionStatus"
)

var _ Provider = (*KrakenProvider)(nil)

type (
	// KrakenProvider defines an Oracle provider implemented by the Kraken public
	// API.
	//
	// REF: https://docs.kraken.com/websockets/#overview
	KrakenProvider struct {
		wsURL           url.URL
		wsClient        *websocket.Conn
		logger          zerolog.Logger
		mtx             sync.Mutex
		tickers         map[string]TickerPrice        // Symbol => TickerPrice
		subscribedPairs map[string]types.CurrencyPair // Symbol => types.CurrencyPair
	}

	// KrakenTicker ticker price response from Kraken ticker channel.
	// https://docs.kraken.com/websockets/#message-ticker
	KrakenTicker struct {
		C []string `json:"c"` // Close with Price in the first position
		V []string `json:"v"` // Volume with the value over last 24 hours in the second position
	}

	// KrakenSubscriptionMsg Msg to subscribe to all the pairs at once.
	KrakenSubscriptionMsg struct {
		Event        string                    `json:"event"`        // subscribe/unsubscribe
		Pair         []string                  `json:"pair"`         // Array of currency pairs ex.: "BTC/USDT",
		Subscription KrakenSubscriptionChannel `json:"subscription"` // subscription object
	}

	// KrakenSubscriptionChannel Msg with the channel name to be subscribed.
	KrakenSubscriptionChannel struct {
		Name string `json:"name"` // channel to be subscribed ex.: ticker
	}

	// KrakenEvent wraps the possible events from the provider.
	KrakenEvent struct {
		Event string `json:"event"` // events from kraken ex.: systemStatus | subscriptionStatus
	}

	// KrakenEventSystemStatus parse the systemStatus event message.
	KrakenEventSystemStatus struct {
		Status string `json:"status"` // online|maintenance|cancel_only|limit_only|post_only
	}

	// KrakenEventSubscriptionStatus parse the subscriptionStatus event message.
	KrakenEventSubscriptionStatus struct {
		Status       string `json:"status"`       // subscribed|unsubscribed|error
		Pair         string `json:"pair"`         // Pair symbol base/quote ex.: "XBT/USD"
		ErrorMessage string `json:"errorMessage"` // error description
	}
)

// NewKrakenProvider returns a new Kraken provider with the WS connection and msg handler.
func NewKrakenProvider(ctx context.Context, logger zerolog.Logger, pairs ...types.CurrencyPair) (*KrakenProvider, error) {
	wsURL := url.URL{
		Scheme: "wss",
		Host:   krakenHost,
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error connecting to websocket: %w", err)
	}

	provider := &KrakenProvider{
		wsURL:           wsURL,
		wsClient:        wsConn,
		logger:          logger.With().Str("provider", "kraken").Logger(),
		tickers:         map[string]TickerPrice{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	if err := provider.SubscribeTickers(pairs...); err != nil {
		return nil, err
	}

	go provider.handleWebSocketMsgs(ctx)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the saved map.
func (p *KrakenProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
	tickerPrices := make(map[string]TickerPrice, len(pairs))

	for _, cp := range pairs {
		key := cp.String()
		tickerPrice, ok := p.tickers[key]
		if !ok {
			return nil, fmt.Errorf("kraken provider failed to get ticker price for %s", key)
		}
		tickerPrices[key] = tickerPrice
	}

	return tickerPrices, nil
}

// SubscribeTickers subscribe to all currency pairs and
// add the new ones into the provider subscribed pairs.
func (p *KrakenProvider) SubscribeTickers(cps ...types.CurrencyPair) error {
	pairs := make([]string, len(cps))

	for i, cp := range cps {
		pairs[i] = currencyPairToKrakenPair(cp)
	}

	if err := p.subscribePairs(pairs...); err != nil {
		return err
	}

	p.setSubscribedPairs(cps...)
	return nil
}

// handleWebSocketMsgs receive all the messages from the provider
// and controls to reconnect the web socket.
func (p *KrakenProvider) handleWebSocketMsgs(ctx context.Context) {
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
				p.logger.Err(err).Msg("provider could not read message")
				if err := p.ping(); err != nil {
					p.logger.Err(err).Msg("failed to send ping")
					p.keepReconnecting()
				}
				continue
			}

			if len(bz) == 0 {
				continue
			}

			p.messageReceived(messageType, bz)

		case <-reconnectTicker.C:
			if err := p.reconnect(); err != nil {
				p.logger.Err(err).Msg("provider attempted to reconnect")
				p.keepReconnecting()
			}
		}
	}
}

// messageReceived handles any message sent by the provider.
func (p *KrakenProvider) messageReceived(messageType int, bz []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var krakenEvent KrakenEvent
	if err := json.Unmarshal(bz, &krakenEvent); err != nil {
		p.logger.Debug().Msg("provider received a message that is not an event")
		// msg is not an event, it will try to marshal to ticker message
		p.messageReceivedTickerPrice(bz)
		return
	}

	switch krakenEvent.Event {
	case krakenEventSystemStatus:
		p.messageReceivedSystemStatus(bz)
		return
	case krakenEventSubscriptionStatus:
		p.messageReceivedSubscriptionStatus(bz)
		return
	}
}

// messageReceivedTickerPrice handles the ticker price msg.
func (p *KrakenProvider) messageReceivedTickerPrice(bz []byte) {
	// the provider response is an array with different types at each index
	// kraken documentation https://docs.kraken.com/websockets/#message-ticker
	var tickerMessage []interface{}
	if err := json.Unmarshal(bz, &tickerMessage); err != nil {
		p.logger.Err(err).Msg("provider could not unmarshal")
		return
	}

	if len(tickerMessage) != 4 {
		p.logger.Debug().Msg("provider sent something different than ticker")
		return
	}

	channelName, ok := tickerMessage[2].(string)
	if !ok || channelName != "ticker" {
		p.logger.Debug().Msg("provider sent an unexpected channel name")
		return
	}

	tickerBz, err := json.Marshal(tickerMessage[1])
	if err != nil {
		p.logger.Err(err).Msg("provider could not marshal ticker message")
		return
	}

	var krakenTicker KrakenTicker
	if err := json.Unmarshal(tickerBz, &krakenTicker); err != nil {
		p.logger.Err(err).Msg("provider could not unmarshal ticker")
		return
	}

	krakenPair, ok := tickerMessage[3].(string)
	if !ok {
		p.logger.Debug().Msg("provider sent an unexpected pair")
		return
	}

	krakenPair = normalizeKrakenBTCPair(krakenPair)
	currencyPairSymbol := krakenPairToCurrencyPairSymbol(krakenPair)

	tickerPrice, err := krakenTicker.toTickerPrice(currencyPairSymbol)
	if err != nil {
		p.logger.Err(err).Msg("provider could not parse kraken ticker to ticker price")
		return
	}

	p.setTickerPair(currencyPairSymbol, tickerPrice)
}

// reconnect closes the last WS connection and create a new one.
func (p *KrakenProvider) reconnect() error {
	p.wsClient.Close()

	wsConn, _, err := websocket.DefaultDialer.Dial(p.wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("error connecting to Kraken websocket: %w", err)
	}
	p.wsClient = wsConn

	pairs := make([]string, len(p.subscribedPairs))
	iterator := 0
	for _, cp := range p.subscribedPairs {
		pairs[iterator] = currencyPairToKrakenPair(cp)
		iterator++
	}

	return p.subscribePairs(pairs...)
}

// keepReconnecting keeps trying to reconnect if an error occurs in recconnect.
func (p *KrakenProvider) keepReconnecting() {
	reconnectTicker := time.NewTicker(defaultReconnectTime)
	defer reconnectTicker.Stop()
	connectionTries := 1

	for time := range reconnectTicker.C {
		if err := p.reconnect(); err != nil {
			p.logger.Err(err).Msgf("provider attempted to reconnect %d times at %s", connectionTries, time.String())
			continue
		}

		if connectionTries > maxReconnectionTries {
			p.logger.Warn().Msgf("provider failed to reconnect %d times", connectionTries)
		}
		connectionTries++
		return
	}
}

// messageReceivedSubscriptionStatus handle the subscription status message
// sent by the provider.
func (p *KrakenProvider) messageReceivedSubscriptionStatus(bz []byte) {
	var subscriptionStatus KrakenEventSubscriptionStatus
	if err := json.Unmarshal(bz, &subscriptionStatus); err != nil {
		p.logger.Err(err).Msg("provider could not unmarshal KrakenEventSubscriptionStatus")
		return
	}

	switch subscriptionStatus.Status {
	case "error":
		p.logger.Error().Msg(subscriptionStatus.ErrorMessage)
		p.removeSubscribedTickers(krakenPairToCurrencyPairSymbol(subscriptionStatus.Pair))
		return
	case "unsubscribed":
		p.logger.Debug().Msgf("ticker %s was unsubscribed", subscriptionStatus.Pair)
		p.removeSubscribedTickers(krakenPairToCurrencyPairSymbol(subscriptionStatus.Pair))
		return
	}
}

// messageReceivedSystemStatus handle the system status and try to reconnect if it is not online.
func (p *KrakenProvider) messageReceivedSystemStatus(bz []byte) {
	var systemStatus KrakenEventSystemStatus
	if err := json.Unmarshal(bz, &systemStatus); err != nil {
		p.logger.Err(err).Msg("provider could not unmarshal KrakenEventSystemStatus")
		return
	}

	if strings.EqualFold(systemStatus.Status, "online") {
		return
	}

	p.keepReconnecting()
}

// setTickerPair sets an ticker to the map thread safe by the mutex.
func (p *KrakenProvider) setTickerPair(symbol string, ticker TickerPrice) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.tickers[symbol] = ticker
}

// ping to check websocket connection.
func (p *KrakenProvider) ping() error {
	return p.wsClient.WriteMessage(websocket.PingMessage, ping)
}

// subscribePairs write the subscription msg to the provider.
func (p *KrakenProvider) subscribePairs(pairs ...string) error {
	subsMsg := newKrakenSubscriptionMsg(pairs...)
	return p.wsClient.WriteJSON(subsMsg)
}

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *KrakenProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

// removeSubscribedTickers delete N pairs from the subscribed map.
func (p *KrakenProvider) removeSubscribedTickers(tickerSymbols ...string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, tickerSymbol := range tickerSymbols {
		delete(p.subscribedPairs, tickerSymbol)
	}
}

// toTickerPrice return a TickerPrice based on the KrakenTicker.
func (ticker KrakenTicker) toTickerPrice(symbol string) (TickerPrice, error) {
	if len(ticker.C) != 2 || len(ticker.V) != 2 {
		return TickerPrice{}, fmt.Errorf("error converting KrakenTicker to TickerPrice")
	}
	// ticker.C has the Price in the first position
	// ticker.V has the totla	Value over last 24 hours in the second position
	return newTickerPrice("Kraken", symbol, ticker.C[0], ticker.V[1])
}

// newKrakenSubscriptionMsg returns a new subscription Msg.
func newKrakenSubscriptionMsg(pairs ...string) KrakenSubscriptionMsg {
	return KrakenSubscriptionMsg{
		Event: "subscribe",
		Pair:  pairs,
		Subscription: KrakenSubscriptionChannel{
			Name: "ticker",
		},
	}
}

// krakenPairToCurrencyPairSymbol receives a kraken pair formated
// ex.: ATOM/USDT and return currencyPair Symbol ATOMUSDT.
func krakenPairToCurrencyPairSymbol(krakenPair string) string {
	return strings.Replace(krakenPair, "/", "", -1)
}

// currencyPairToKrakenPair receives a currency pair
// and return kraken ticker symbol ATOM/USDT.
func currencyPairToKrakenPair(cp types.CurrencyPair) string {
	return strings.ToUpper(cp.Base + "/" + cp.Quote)
}

// normalizeKrakenBTCPair changes XBT pairs to BTC,
// since other providers list bitcoin as BTC
func normalizeKrakenBTCPair(ticker string) string {
	return strings.Replace(ticker, "XBT", "BTC", 1)
}
