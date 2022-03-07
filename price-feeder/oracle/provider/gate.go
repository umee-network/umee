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
	gateHost      = "ws.gate.io"
	gatePath      = "/v3"
	gatePingCheck = time.Second * 28 // should be < 30
)

var _ Provider = (*GateProvider)(nil)

type (
	// GateProvider defines an Oracle provider implemented by the Gate public
	// API.
	//
	// REF: https://www.gate.com/docs-v5/en/#websocket-api-public-channel-tickers-channel
	GateProvider struct {
		wsURL           url.URL
		wsClient        *websocket.Conn
		logger          zerolog.Logger
		reconnectTimer  *time.Ticker
		mtx             sync.RWMutex
		tickers         map[string]GateTicker         // Symbol => GateTickerPair
		subscribedPairs map[string]types.CurrencyPair // Symbol => types.CurrencyPair
	}

	// GateTickerPair defines a ticker pair of Gate.
	GateTicker struct {
		Last   string `json:"last"`       // Last traded price ex.: 43508.9
		Vol    string `json:"baseVolume"` // Trading volume ex.: 11159.87127845
		Symbol string // Symbol ex.: ATOM_UDST
	}

	// GateSubscriptionMsg Msg to subscribe all the tickers channels.
	GateSubscriptionMsg struct {
		Method string   `json:"method"` // ticker.subscribe
		Params []string `json:"params"` // streams to subscribe ex.: BOT_USDT
		ID     uint16   `json:"id"`     // identify messages going back and forth
	}

	// GateTickerResponse defines the response body for gate tickers
	GateTickerResponse struct {
		Method string        `json:"method"`
		Params []interface{} `json:"params"`
	}

	// GateEvent defines the response body for gate subscription statuses
	GateEvent struct {
		ID     int             `json:"id"`
		Result GateEventResult `json:"result"`
	}
	GateEventResult struct {
		Status string `json:"status"`
	}
)

// NewGateProvider creates a new GateProvider.
func NewGateProvider(ctx context.Context, logger zerolog.Logger, pairs ...types.CurrencyPair) (*GateProvider, error) {
	wsURL := url.URL{
		Scheme: "wss",
		Host:   gateHost,
		Path:   gatePath,
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Gate websocket: %w", err)
	}

	provider := &GateProvider{
		wsURL:           wsURL,
		wsClient:        wsConn,
		logger:          logger.With().Str("provider", "gate").Logger(),
		reconnectTimer:  time.NewTicker(gatePingCheck),
		tickers:         map[string]GateTicker{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}
	provider.wsClient.SetPongHandler(provider.pongHandler)

	if err := provider.SubscribeTickers(pairs...); err != nil {
		return nil, err
	}

	go provider.handleReceivedTickers(ctx)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the saved map.
func (p *GateProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
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

// GetCandlePrices returns the candlePrices based on the saved map
func (p *GateProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]CandlePrice, error) {
	return nil, nil
}

// SubscribeTickers subscribe all currency pairs into ticker and candle channels.
func (p *GateProvider) SubscribeTickers(cps ...types.CurrencyPair) error {
	topics := []string{}

	for _, cp := range cps {
		topics = append(topics, currencyPairToGatePair(cp))
	}

	tickerMsg := newGateTickerSubscription(topics...)
	if err := p.subscribePairs(tickerMsg); err != nil {
		return err
	}

	p.setSubscribedPairs(cps...)
	return nil
}

func (p *GateProvider) getTickerPrice(cp types.CurrencyPair) (TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	gp := currencyPairToGatePair(cp)
	tickerPair, ok := p.tickers[gp]
	if !ok {
		return TickerPrice{}, fmt.Errorf("gate provider failed to get ticker price for %s", gp)
	}

	return tickerPair.toTickerPrice()
}

func (p *GateProvider) handleReceivedTickers(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(defaultReadNewWSMessage):
			messageType, bz, err := p.wsClient.ReadMessage()
			if err != nil {
				// if some error occurs continue to try to read the next message.
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

		case <-p.reconnectTimer.C: // reset by the pongHandler.
			if err := p.reconnect(); err != nil {
				p.logger.Err(err).Msg("error reconnecting")
			}
		}
	}
}

func (p *GateProvider) messageReceived(messageType int, bz []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var gateEvent GateEvent
	if err := json.Unmarshal(bz, &gateEvent); err != nil {
		// msg is not an event, it will try to marshal to ticker message.
		p.logger.Debug().Msg("received a message that is not an event")
	} else {
		switch gateEvent.Result.Status {
		case "success":
			return
		default:
			p.reconnect()
		}
	}

	if err := p.messageReceivedTickerPrice(bz); err != nil {
		// msg is not a ticker, it will try to marshal to candle message.
		p.logger.Debug().Err(err).Msg("unable to unmarshal ticker")
	}
}

// messageReceivedTickerPrice handles the ticker price msg.
func (p *GateProvider) messageReceivedTickerPrice(bz []byte) error {
	// the provider response is an array with different types at each index
	// gate documentation https://www.gate.io/docs/websocket/index.html
	var tickerMessage GateTickerResponse
	if err := json.Unmarshal(bz, &tickerMessage); err != nil {
		return err
	}

	tickerBz, err := json.Marshal(tickerMessage.Params[1])
	if err != nil {
		p.logger.Err(err).Msg("could not marshal ticker message")
		return err
	}

	var gateTicker GateTicker
	if err := json.Unmarshal(tickerBz, &gateTicker); err != nil {
		p.logger.Err(err).Msg("could not unmarshal ticker message")
		return err
	}

	pairName, ok := tickerMessage.Params[0].(string)
	if !ok {
		return fmt.Errorf("unable to unmarshal pair name")
	}
	gateTicker.Symbol = pairName

	p.setTickerPair(gateTicker)
	return nil
}

func (p *GateProvider) setTickerPair(ticker GateTicker) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.tickers[ticker.Symbol] = ticker
}

// subscribePairs write the subscription msg to the provider.
func (p *GateProvider) subscribePairs(msg GateSubscriptionMsg) error {
	return p.wsClient.WriteJSON(msg)
}

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *GateProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

func (p *GateProvider) resetReconnectTimer() {
	p.reconnectTimer.Reset(gatePingCheck)
}

// reconnect closes the last WS connection and creates a new one. If there’s a
// network problem, the system will automatically disable the connection. The
// connection will break automatically if the subscription is not established or
// data has not been pushed for more than 30 seconds. To keep the connection stable:
// 1. Set a timer of N seconds whenever a response message is received, where N is
// less than 30.
// 2. If the timer is triggered, which means that no new message is received within
// N seconds, send the String 'ping'.
// 3. Expect a 'pong' as a response. If the response message is not received within
// N seconds, please raise an error or reconnect.
func (p *GateProvider) reconnect() error {
	p.wsClient.Close()

	p.logger.Debug().Msg("reconnecting websocket")
	wsConn, _, err := websocket.DefaultDialer.Dial(p.wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("error reconnecting to Gate websocket: %w", err)
	}
	wsConn.SetPongHandler(p.pongHandler)
	p.wsClient = wsConn

	topics := []string{}
	for _, cp := range p.subscribedPairs {
		topics = append(topics, currencyPairToGatePair(cp))
	}

	tickerMsg := newGateTickerSubscription(topics...)
	if err := p.subscribePairs(tickerMsg); err != nil {
		return err
	}

	return p.subscribePairs(tickerMsg)
}

// ping to check websocket connection.
func (p *GateProvider) ping() error {
	return p.wsClient.WriteMessage(websocket.PingMessage, ping)
}

func (p *GateProvider) pongHandler(appData string) error {
	p.resetReconnectTimer()
	return nil
}

func (ticker GateTicker) toTickerPrice() (TickerPrice, error) {
	return newTickerPrice("Gate", ticker.Symbol, ticker.Last, ticker.Vol)
}

// currencyPairToGatePair returns the expected pair for Gate
// ex.: "ATOM_USDT".
func currencyPairToGatePair(pair types.CurrencyPair) string {
	return pair.Base + "_" + pair.Quote
}

// newGateTickerSubscription returns a new subscription topic.
func newGateTickerSubscription(cp ...string) GateSubscriptionMsg {
	return GateSubscriptionMsg{
		Method: "ticker.subscribe",
		Params: cp,
		ID:     1,
	}
}
