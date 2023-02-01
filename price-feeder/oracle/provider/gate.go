package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"github.com/umee-network/umee/price-feeder/v2/oracle/types"
)

const (
	gateWSHost    = "ws.gate.io"
	gateWSPath    = "/v4"
	gatePingCheck = time.Second * 28 // should be < 30
	gateRestHost  = "https://api.gateio.ws"
	gateRestPath  = "/api/v4/spot/currency_pairs"
)

var _ Provider = (*GateProvider)(nil)

type (
	// GateProvider defines an Oracle provider implemented by the Gate public
	// API.
	//
	// REF: https://www.gate.io/docs/websocket/index.html
	GateProvider struct {
		wsc             WebsocketController
		logger          zerolog.Logger
		reconnectTimer  *time.Ticker
		mtx             sync.RWMutex
		endpoints       Endpoint
		tickers         map[string]GateTicker         // Symbol => GateTicker
		candles         map[string][]GateCandle       // Symbol => GateCandle
		subscribedPairs map[string]types.CurrencyPair // Symbol => types.CurrencyPair
	}

	GateTicker struct {
		Last   string `json:"last"`       // Last traded price ex.: 43508.9
		Vol    string `json:"baseVolume"` // Trading volume ex.: 11159.87127845
		Symbol string `json:"symbol"`     // Symbol ex.: ATOM_UDST
	}

	GateCandle struct {
		Close     string // Closing price
		TimeStamp int64  // Unix timestamp
		Volume    string // Total candle volume
		Symbol    string // Total symbol
	}

	// GateTickerSubscriptionMsg Msg to subscribe all the tickers channels.
	GateTickerSubscriptionMsg struct {
		Method string   `json:"method"` // ticker.subscribe
		Params []string `json:"params"` // streams to subscribe ex.: BOT_USDT
		ID     uint16   `json:"id"`     // identify messages going back and forth
	}

	// GateCandleSubscriptionMsg Msg to subscribe to a candle channel.
	GateCandleSubscriptionMsg struct {
		Method string        `json:"method"` // ticker.subscribe
		Params []interface{} `json:"params"` // streams to subscribe ex.: ["BOT_USDT": 1800]
		ID     uint16        `json:"id"`     // identify messages going back and forth
	}

	// GateTickerResponse defines the response body for gate tickers.
	GateTickerResponse struct {
		Method string        `json:"method"`
		Params []interface{} `json:"params"`
	}

	// GateTickerResponse defines the response body for gate tickers.
	// The Params response is a 2D slice of multiple candles and their data.
	//
	// REF: https://www.gate.io/docs/websocket/index.html
	GateCandleResponse struct {
		Method string          `json:"method"`
		Params [][]interface{} `json:"params"`
	}

	// GateEvent defines the response body for gate subscription statuses.
	GateEvent struct {
		ID     int             `json:"id"`     // subscription id, ex.: 123
		Result GateEventResult `json:"result"` // event result body
	}
	// GateEventResult defines the Result body for the GateEvent response.
	GateEventResult struct {
		Status string `json:"status"` // ex. "successful"
	}

	// GatePairSummary defines the response structure for a Gate pair summary.
	GatePairSummary struct {
		Base  string `json:"base"`
		Quote string `json:"quote"`
	}
)

// NewGateProvider creates a new GateProvider.
func NewGateProvider(
	ctx context.Context,
	logger zerolog.Logger,
	endpoints Endpoint,
	pairs ...types.CurrencyPair,
) (*GateProvider, error) {
	if endpoints.Name != ProviderGate {
		endpoints = Endpoint{
			Name:      ProviderGate,
			Rest:      gateRestHost,
			Websocket: gateWSHost,
		}
	}

	wsURL := url.URL{
		Scheme: "wss",
		Host:   endpoints.Websocket,
		Path:   gateWSPath,
	}

	gateLogger := logger.With().Str("provider", string(ProviderGate)).Logger()

	provider := &GateProvider{
		logger:          gateLogger,
		reconnectTimer:  time.NewTicker(gatePingCheck),
		endpoints:       endpoints,
		tickers:         map[string]GateTicker{},
		candles:         map[string][]GateCandle{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	provider.setSubscribedPairs(pairs...)

	provider.wsc = NewWebsocketController(
		ctx,
		ProviderGate,
		wsURL,
		provider.getSubscriptionMsgs(pairs...),
		provider.messageReceived,
		defaultPingDuration,
		websocket.PingMessage,
		gateLogger,
	)
	go provider.wsc.Start()

	return provider, nil
}

func (p *GateProvider) getSubscriptionMsgs(cps ...types.CurrencyPair) []interface{} {
	subscriptionMsgs := make([]interface{}, 0, len(cps)*2)
	for _, cp := range cps {
		gatePair := currencyPairToGatePair(cp)
		subscriptionMsgs = append(subscriptionMsgs, newGateTickerSubscription(gatePair))
		subscriptionMsgs = append(subscriptionMsgs, newGateCandleSubscription(gatePair))
	}
	return subscriptionMsgs
}

// SubscribeCurrencyPairs sends the new subscription messages to the websocket
// and adds them to the providers subscribedPairs array
func (p *GateProvider) SubscribeCurrencyPairs(cps ...types.CurrencyPair) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	newPairs := []types.CurrencyPair{}
	for _, cp := range cps {
		if _, ok := p.subscribedPairs[cp.String()]; !ok {
			newPairs = append(newPairs, cp)
		}
	}

	newSubscriptionMsgs := p.getSubscriptionMsgs(newPairs...)
	if err := p.wsc.AddSubscriptionMsgs(newSubscriptionMsgs); err != nil {
		return err
	}
	p.setSubscribedPairs(newPairs...)
	return nil
}

// GetTickerPrices returns the tickerPrices based on the saved map.
func (p *GateProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	tickerPrices := make(map[string]types.TickerPrice, len(pairs))

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
func (p *GateProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candlePrices := make(map[string][]types.CandlePrice, len(pairs))

	for _, currencyPair := range pairs {
		gp := currencyPairToGatePair(currencyPair)
		price, err := p.getCandlePrices(gp)
		if err != nil {
			return nil, err
		}

		candlePrices[currencyPair.String()] = price
	}

	return candlePrices, nil
}

func (p *GateProvider) getCandlePrices(key string) ([]types.CandlePrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	candles, ok := p.candles[key]
	if !ok {
		return []types.CandlePrice{}, fmt.Errorf("gate failed to get candle prices for %s", key)
	}

	candleList := []types.CandlePrice{}
	for _, candle := range candles {
		cp, err := candle.toCandlePrice()
		if err != nil {
			return []types.CandlePrice{}, err
		}

		candleList = append(candleList, cp)
	}

	return candleList, nil
}

func (p *GateProvider) getTickerPrice(cp types.CurrencyPair) (types.TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	gp := currencyPairToGatePair(cp)
	if tickerPair, ok := p.tickers[gp]; ok {
		return tickerPair.toTickerPrice()
	}

	return types.TickerPrice{}, fmt.Errorf("gate failed to get ticker price for %s", gp)
}

func (p *GateProvider) messageReceived(_ int, bz []byte) {
	var (
		gateEvent GateEvent
		gateErr   error
		tickerErr error
		candleErr error
	)

	gateErr = json.Unmarshal(bz, &gateEvent)
	if gateErr == nil {
		switch gateEvent.Result.Status {
		case "success":
			return
		case "":
		default:
			return
		}
	}

	tickerErr = p.messageReceivedTickerPrice(bz)
	if tickerErr == nil {
		return
	}

	candleErr = p.messageReceivedCandle(bz)
	if candleErr == nil {
		return
	}

	p.logger.Error().
		Int("length", len(bz)).
		AnErr("ticker", tickerErr).
		AnErr("candle", candleErr).
		AnErr("event", gateErr).
		Msg("Error on receive message")
}

// messageReceivedTickerPrice handles the ticker price msg.
// The provider response is a slice with different types at each index.
//
// REF: https://www.gate.io/docs/websocket/index.html
func (p *GateProvider) messageReceivedTickerPrice(bz []byte) error {
	var tickerMessage GateTickerResponse
	if err := json.Unmarshal(bz, &tickerMessage); err != nil {
		return err
	}

	if tickerMessage.Method != "ticker.update" {
		return fmt.Errorf("message is not a ticker update")
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

	symbol, ok := tickerMessage.Params[0].(string)
	if !ok {
		return fmt.Errorf("symbol should be a string")
	}
	gateTicker.Symbol = symbol

	p.setTickerPair(gateTicker)
	telemetryWebsocketMessage(ProviderGate, MessageTypeTicker)
	return nil
}

// UnmarshalParams is a helper function which unmarshals the 2d slice of interfaces
// from a GateCandleResponse into the GateCandle.
func (candle *GateCandle) UnmarshalParams(params [][]interface{}) error {
	var tmp []interface{}

	if len(params) == 0 {
		return fmt.Errorf("no candles in response")
	}

	// use the most recent candle
	tmp = params[len(params)-1]
	if len(tmp) != 8 {
		return fmt.Errorf("wrong number of fields in candle")
	}

	time := int64(tmp[0].(float64))
	if time == 0 {
		return fmt.Errorf("time field must be a float")
	}
	candle.TimeStamp = time

	close, ok := tmp[1].(string)
	if !ok {
		return fmt.Errorf("close field must be a string")
	}
	candle.Close = close

	volume, ok := tmp[5].(string)
	if !ok {
		return fmt.Errorf("volume field must be a string")
	}
	candle.Volume = volume

	symbol, ok := tmp[7].(string)
	if !ok {
		return fmt.Errorf("symbol field must be a string")
	}
	candle.Symbol = symbol

	return nil
}

// messageReceivedCandle handles the candle price msg.
// The provider response is a slice with different types at each index.
//
// REF: https://www.gate.io/docs/websocket/index.html
func (p *GateProvider) messageReceivedCandle(bz []byte) error {
	var candleMessage GateCandleResponse
	if err := json.Unmarshal(bz, &candleMessage); err != nil {
		return err
	}

	if candleMessage.Method != "kline.update" {
		return fmt.Errorf("message is not a kline update")
	}

	var gateCandle GateCandle
	if err := gateCandle.UnmarshalParams(candleMessage.Params); err != nil {
		return err
	}

	p.setCandlePair(gateCandle)
	telemetryWebsocketMessage(ProviderGate, MessageTypeCandle)
	return nil
}

func (p *GateProvider) setTickerPair(ticker GateTicker) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.tickers[ticker.Symbol] = ticker
}

func (p *GateProvider) setCandlePair(candle GateCandle) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	// convert gate timestamp seconds -> milliseconds
	candle.TimeStamp = SecondsToMilli(candle.TimeStamp)
	staleTime := PastUnixTime(providerCandlePeriod)
	candleList := []GateCandle{}

	candleList = append(candleList, candle)
	for _, c := range p.candles[candle.Symbol] {
		if staleTime < c.TimeStamp {
			candleList = append(candleList, c)
		}
	}
	p.candles[candle.Symbol] = candleList
}

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *GateProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

// GetAvailablePairs returns all pairs to which the provider can subscribe.
func (p *GateProvider) GetAvailablePairs() (map[string]struct{}, error) {
	resp, err := http.Get(p.endpoints.Rest + gateRestPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary []GatePairSummary
	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return nil, err
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary))
	for _, pair := range pairsSummary {
		cp := types.CurrencyPair{
			Base:  strings.ToUpper(pair.Base),
			Quote: strings.ToUpper(pair.Quote),
		}
		availablePairs[cp.String()] = struct{}{}
	}

	return availablePairs, nil
}

func (ticker GateTicker) toTickerPrice() (types.TickerPrice, error) {
	return types.NewTickerPrice(string(ProviderGate), ticker.Symbol, ticker.Last, ticker.Vol)
}

func (candle GateCandle) toCandlePrice() (types.CandlePrice, error) {
	return types.NewCandlePrice(
		string(ProviderGate),
		candle.Symbol,
		candle.Close,
		candle.Volume,
		candle.TimeStamp,
	)
}

// currencyPairToGatePair returns the expected pair for Gate
// ex.: "ATOM_USDT".
func currencyPairToGatePair(pair types.CurrencyPair) string {
	return pair.Base + "_" + pair.Quote
}

// newGateTickerSubscription returns a new subscription topic for tickers.
func newGateTickerSubscription(cp ...string) GateTickerSubscriptionMsg {
	return GateTickerSubscriptionMsg{
		Method: "ticker.subscribe",
		Params: cp,
		ID:     1,
	}
}

// newGateCandleSubscription returns a new subscription topic for candles.
func newGateCandleSubscription(gatePair string) GateCandleSubscriptionMsg {
	params := []interface{}{
		gatePair, // currency pair ex. "ATOM_USDT"
		60,       // time interval in seconds
	}
	return GateCandleSubscriptionMsg{
		Method: "kline.subscribe",
		Params: params,
		ID:     2,
	}
}
