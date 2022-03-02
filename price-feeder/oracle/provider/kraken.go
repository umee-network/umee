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
		candles         map[string]KrakenCandle       // Symbol => KrakenCandle
		subscribedPairs map[string]types.CurrencyPair // Symbol => types.CurrencyPair
	}

	// KrakenTicker ticker price response from Kraken ticker channel.
	// https://docs.kraken.com/websockets/#message-ticker
	KrakenTicker struct {
		C []string `json:"c"` // Close with Price in the first position
		V []string `json:"v"` // Volume with the value over last 24 hours in the second position
	}

	// KrakenCandle candle response from Kraken candle channel.
	// REF : https://docs.kraken.com/websockets/#message-ohlc
	KrakenCandle struct {
		Close     string // Close price during this period
		TimeStamp int64  // Linux epoch timestamp
		Volume    string // Volume during this period
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
		candles:         map[string]KrakenCandle{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	if err := provider.SubscribeTickers(pairs...); err != nil {
		return nil, err
	}
	if err := provider.SubscribeCandles(pairs...); err != nil {
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
			return nil, fmt.Errorf("failed to get ticker price for %s", key)
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

	if err := p.subscribeTickerPairs(pairs...); err != nil {
		return err
	}
	if err := p.subscribeCandlePairs(pairs...); err != nil {
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
				p.logger.Err(err).Msg("could not read message")
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
				p.logger.Err(err).Msg("attempted to reconnect")
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
		// msg is not an event, it will try to marshal to ticker message
		p.logger.Debug().Msg("received a message that is not an event")
	} else {
		switch krakenEvent.Event {
		case krakenEventSystemStatus:
			p.messageReceivedSystemStatus(bz)
			return
		case krakenEventSubscriptionStatus:
			p.messageReceivedSubscriptionStatus(bz)
			return
		}
	}

	if err := p.messageReceivedTickerPrice(bz); err != nil {
		// msg is not a ticker, it will try to marshal to candle message
		p.logger.Debug().Err(err).Msg("unable to unmarshal ticker")
	} else {
		return
	}
	if err := p.messageReceivedCandle(bz); err != nil {
		p.logger.Debug().Err(err).Msg("unable to unmarshal candle")
	}
}

// messageReceivedTickerPrice handles the ticker price msg.
func (p *KrakenProvider) messageReceivedTickerPrice(bz []byte) error {
	// the provider response is an array with different types at each index
	// kraken documentation https://docs.kraken.com/websockets/#message-ticker
	var tickerMessage []interface{}
	if err := json.Unmarshal(bz, &tickerMessage); err != nil {
		return err
	}

	if len(tickerMessage) != 4 {
		return fmt.Errorf("received an unexpected structure")
	}

	channelName, ok := tickerMessage[1].(string)
	if !ok || channelName != "ticker" {
		return fmt.Errorf("received an unexpected channel name")
	}

	tickerBz, err := json.Marshal(tickerMessage[1])
	if err != nil {
		p.logger.Err(err).Msg("could not marshal ticker message")
		return err
	}

	var krakenTicker KrakenTicker
	if err := json.Unmarshal(tickerBz, &krakenTicker); err != nil {
		p.logger.Err(err).Msg("could not unmarshal ticker message")
		return err
	}

	krakenPair, ok := tickerMessage[3].(string)
	if !ok {
		p.logger.Debug().Msg("received an unexpected pair")
		return err
	}

	krakenPair = normalizeKrakenBTCPair(krakenPair)
	currencyPairSymbol := krakenPairToCurrencyPairSymbol(krakenPair)

	tickerPrice, err := krakenTicker.toTickerPrice(currencyPairSymbol)
	if err != nil {
		p.logger.Err(err).Msg("could not parse kraken ticker to ticker price")
		return err
	}

	p.setTickerPair(currencyPairSymbol, tickerPrice)
	return nil
}

func (kc *KrakenCandle) UnmarshalJSON(buf []byte) error {
	var tmp []interface{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 9 {
		return fmt.Errorf("wrong number of fields in candle")
	}

	time, ok := tmp[2].(int64)
	if !ok {
		return fmt.Errorf("time field must be an int64")
	}
	kc.TimeStamp = time

	close, ok := tmp[5].(string)
	if !ok {
		return fmt.Errorf("close field must be a string")
	}
	kc.Close = close

	volume, ok := tmp[7].(string)
	if !ok {
		return fmt.Errorf("volume field must be a string")
	}
	kc.Volume = volume

	return nil
}

// messageReceivedCandle handles the candle msg.
func (p *KrakenProvider) messageReceivedCandle(bz []byte) error {
	// the provider response is an array with different types at each index
	// kraken documentation https://docs.kraken.com/websockets/#message-ohlc
	var candleMessage []interface{}
	if err := json.Unmarshal(bz, &candleMessage); err != nil {
		return err
	}

	if len(candleMessage) != 4 {
		return fmt.Errorf("received something different than candle")
	}

	channelName, ok := candleMessage[2].(string)
	if !ok || channelName != "ohlc-1" {
		return fmt.Errorf("received an unexpected channel name")
	}

	tickerBz, err := json.Marshal(candleMessage[1])
	if err != nil {
		return fmt.Errorf("could not marshal candle message")
	}

	var krakenCandle KrakenCandle
	if err = krakenCandle.UnmarshalJSON(tickerBz); err != nil {
		return err
	}

	krakenPair, ok := candleMessage[3].(string)
	if !ok {
		return fmt.Errorf("received an unexpected pair")
	}

	krakenPair = normalizeKrakenBTCPair(krakenPair)
	currencyPairSymbol := krakenPairToCurrencyPairSymbol(krakenPair)

	p.setCandlePair(currencyPairSymbol, krakenCandle)
	return nil
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

	if err := p.subscribeTickerPairs(pairs...); err != nil {
		return err
	}
	return p.subscribeCandlePairs(pairs...)
}

// keepReconnecting keeps trying to reconnect if an error occurs in recconnect.
func (p *KrakenProvider) keepReconnecting() {
	reconnectTicker := time.NewTicker(defaultReconnectTime)
	defer reconnectTicker.Stop()
	connectionTries := 1

	for time := range reconnectTicker.C {
		if err := p.reconnect(); err != nil {
			p.logger.Err(err).Msgf("attempted to reconnect %d times at %s", connectionTries, time.String())
			continue
		}

		if connectionTries > maxReconnectionTries {
			p.logger.Warn().Msgf("failed to reconnect %d times", connectionTries)
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
		p.logger.Err(err).Msg("could not unmarshal event system status")
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

func (p *KrakenProvider) setCandlePair(symbol string, candle KrakenCandle) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.candles[symbol] = candle
}

// ping to check websocket connection.
func (p *KrakenProvider) ping() error {
	return p.wsClient.WriteMessage(websocket.PingMessage, ping)
}

// subscribeTickerPairs write the subscription msg to the provider.
func (p *KrakenProvider) subscribeTickerPairs(pairs ...string) error {
	subsMsg := newKrakenSubscriptionMsg(pairs...)
	return p.wsClient.WriteJSON(subsMsg)
}

// subscribeCandlePairs write the subscription msg to the provider.
func (p *KrakenProvider) subscribeCandlePairs(pairs ...string) error {
	subsMsg := newKrakenCandleSubscriptionMsg(pairs...)
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

// newKrakenSubscriptionMsg returns a new subscription Msg.
func newKrakenCandleSubscriptionMsg(pairs ...string) KrakenSubscriptionMsg {
	return KrakenSubscriptionMsg{
		Event: "subscribe",
		Pair:  pairs,
		Subscription: KrakenSubscriptionChannel{
			Name: "ohlc",
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
