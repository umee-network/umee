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
	okxHost          = "ws.okx.com:8443"
	okxPath          = "/ws/v5/public"
	msReadNewMessage = 50
)

var _ Provider = (*OkxProvider)(nil)

type (
	// OkxProvider defines an Oracle provider implemented by the Okx public
	// API.
	//
	// REF: https://www.okx.com/docs-v5/en/#websocket-api-public-channel-tickers-channel
	OkxProvider struct {
		wsURL            url.URL
		wsClient         *websocket.Conn
		logger           zerolog.Logger
		mu               sync.Mutex
		tickers          map[string]OkxTickerPair // InstId => OkxTickerPair
		msReadNewMessage uint16
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
		wsURL:            wsURL,
		wsClient:         wsConn,
		logger:           logger.With().Str("module", "oracle").Logger(),
		tickers:          map[string]OkxTickerPair{},
		msReadNewMessage: msReadNewMessage,
	}

	if err := provider.subscribeTickers(pairs...); err != nil {
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
		return TickerPrice{}, fmt.Errorf("failed to get %s", instrumentId)
	}

	return tickerPair.toTickerPrice()
}

func (p *OkxProvider) handleReceivedTickers(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(p.msReadNewMessage) * time.Millisecond):
			// time after to avoid asking for prices too frequently
			messageType, bz, err := p.wsClient.ReadMessage()
			if err != nil {
				// if some error occurs continue to try to read the next message
				p.logger.Err(err).Msg("Okx provider could not read message")
				continue
			}
			p.messageReceived(messageType, bz)
		}
	}
}

func (p *OkxProvider) messageReceived(messageType int, bz []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var tickerRespWS OkxTickerResponse
	if err := json.Unmarshal(bz, &tickerRespWS); err != nil {
		// sometimes it returns other messages which are not tickerResponses
		p.logger.Err(err).Msg("Okx provider could not unmarshal")
		return
	}

	for _, tickerPair := range tickerRespWS.Data {
		p.setTickerPair(tickerPair)
	}
}

func (p *OkxProvider) setTickerPair(tickerPair OkxTickerPair) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tickers[tickerPair.InstId] = tickerPair
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

// newOkxSubscriptionMsg returns a new subscription Msg for Okx
func newOkxSubscriptionMsg(args ...OkxSubscriptionTopic) OkxSubscriptionMsg {
	return OkxSubscriptionMsg{
		Op:   "subscribe",
		Args: args,
	}
}
