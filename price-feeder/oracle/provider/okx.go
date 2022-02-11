package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/websocket"
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
		tickersMap       *sync.Map // InstId => OkxTickerPair
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
		Data []OkxTickerPair   `json:"data"`
		Arg  SubscriptionTopic `json:"arg"`
	}

	SubscriptionTopic struct {
		Channel string `json:"channel"` // Channel name ex.: tickers
		InstId  string `json:"instId"`  // Instrument ID ex.: BTC-USDT
	}

	SubscriptionMsg struct {
		Op   string              `json:"op"` // Operation ex.: subscribe
		Args []SubscriptionTopic `json:"args"`
	}
)

// NewOkxProvider creates a new OkxProvider
func NewOkxProvider(ctx context.Context, pairs ...types.CurrencyPair) (*OkxProvider, error) {
	wsURL := url.URL{
		Scheme: "wss",
		Host:   okxHost,
		Path:   okxPath,
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error connecting to ws: %+v", err)
	}

	provider := &OkxProvider{
		wsURL:            wsURL,
		wsClient:         wsConn,
		tickersMap:       &sync.Map{},
		msReadNewMessage: msReadNewMessage,
	}

	if err := provider.newTickerSubscription(pairs...); err != nil {
		return nil, err
	}

	go provider.handleTickers(ctx)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the saved map
func (p OkxProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
	tickerPrices := make(map[string]TickerPrice, len(pairs))

	for _, currencyPair := range pairs {
		cp := currencyPair
		price, err := p.getTickerPrice(cp)
		if err != nil {
			return nil, err
		}

		tickerPrices[cp.String()] = price
	}

	return tickerPrices, nil
}

func (p OkxProvider) getTickerPrice(cp types.CurrencyPair) (TickerPrice, error) {
	instrumentId := getInstrumentId(cp)
	tickerPair, err := p.getMapTicker(instrumentId)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to get %s - %+v", instrumentId, err)
	}

	return tickerPair.ToTickerPrice()
}

func (p OkxProvider) getMapTicker(instrumentId string) (OkxTickerPair, error) {
	value, ok := p.tickersMap.Load(instrumentId)
	if !ok {
		return OkxTickerPair{}, fmt.Errorf("ticker not found in map %s", instrumentId)
	}

	pair, ok := value.(OkxTickerPair)
	if !ok {
		return OkxTickerPair{}, fmt.Errorf("ticker found %s, but failed on casting to OkxTickerPair", instrumentId)
	}

	return pair, nil
}

func (p OkxProvider) handleTickers(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(p.msReadNewMessage) * time.Millisecond):
			// time after to avoid asking for prices too frequently
			messageType, bz, err := p.wsClient.ReadMessage()
			// if some error occurs continue to try to read the next message
			if err != nil {
				continue
			}
			go p.messageReceived(messageType, bz)
		}
	}
}

func (p OkxProvider) messageReceived(messageType int, bz []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var tickerRespWS OkxTickerResponse
	if err := json.Unmarshal(bz, &tickerRespWS); err != nil {
		// sometimes it returns other messages which are not tickerResponses
		return
	}

	for _, tickerPair := range tickerRespWS.Data {
		p.tickersMap.Store(tickerPair.InstId, tickerPair)
	}
}

// newTickerSubscription subscribe to all currency pairs
func (p OkxProvider) newTickerSubscription(cps ...types.CurrencyPair) error {
	topics := make([]SubscriptionTopic, len(cps))

	for _, cp := range cps {
		instId := getInstrumentId(cp)
		topics = append(topics, newSubscriptionTopic(instId))
	}

	subsMsg := newSubscriptionMsg(topics...)
	return p.wsClient.WriteJSON(subsMsg)
}

func (tickerPair OkxTickerPair) ToTickerPrice() (TickerPrice, error) {
	price, err := sdk.NewDecFromStr(tickerPair.Last)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse Okx price (%s) for %s", tickerPair.Last, tickerPair.InstId)
	}

	volume, err := sdk.NewDecFromStr(tickerPair.Vol24h)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse Okx volume (%s) for %s", tickerPair.Vol24h, tickerPair.InstId)
	}

	return TickerPrice{Price: price, Volume: volume}, nil
}

// getInstrumentId returns the expected pair instrument ID for Okx ex.: BTC-USDT
func getInstrumentId(pair types.CurrencyPair) string {
	return pair.Base + "-" + pair.Quote
}

// newSubscriptionTopic returns a new subscription topic
func newSubscriptionTopic(instId string) SubscriptionTopic {
	return SubscriptionTopic{
		Channel: "tickers",
		InstId:  instId,
	}
}

// newSubscriptionMsg returns a new subscription Msg
func newSubscriptionMsg(args ...SubscriptionTopic) SubscriptionMsg {
	return SubscriptionMsg{
		Op:   "subscribe",
		Args: args,
	}
}
