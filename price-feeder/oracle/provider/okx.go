package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	okxWSHost   = "ws.okx.com:8443"
	okxWSPath   = "/ws/v5/public"
	okxRestHost = "https://www.okx.com"
	okxRestPath = "/api/v5/market/tickers?instType=SPOT"
)

var _ Provider = (*OkxProvider)(nil)

type (
	// OkxProvider defines an Oracle provider implemented by the Okx public
	// API.
	//
	// REF: https://www.okx.com/docs-v5/en/#websocket-api-public-channel-tickers-channel
	OkxProvider struct {
		logger          zerolog.Logger
		mtx             sync.RWMutex
		endpoints       Endpoint
		tickers         map[string]OkxTickerPair      // InstId => OkxTickerPair
		candles         map[string][]OkxCandlePair    // InstId => 0kxCandlePair
		subscribedPairs map[string]types.CurrencyPair // Symbol => types.CurrencyPair
	}

	// OkxInstId defines the id Symbol of an pair.
	OkxInstID struct {
		InstID string `json:"instId"` // Instrument ID ex.: BTC-USDT
	}

	// OkxTickerPair defines a ticker pair of Okx.
	OkxTickerPair struct {
		OkxInstID
		Last   string `json:"last"`   // Last traded price ex.: 43508.9
		Vol24h string `json:"vol24h"` // 24h trading volume ex.: 11159.87127845
	}

	// OkxInst defines the structure containing ID information for the OkxResponses.
	OkxID struct {
		OkxInstID
		Channel string `json:"channel"`
	}

	// OkxTickerResponse defines the response structure of a Okx ticker request.
	OkxTickerResponse struct {
		Data []OkxTickerPair `json:"data"`
		ID   OkxID           `json:"arg"`
	}

	// OkxCandlePair defines a candle for Okx.
	OkxCandlePair struct {
		Close     string `json:"c"`      // Close price for this time period
		TimeStamp int64  `json:"ts"`     // Linux epoch timestamp
		Volume    string `json:"vol"`    // Volume for this time period
		InstID    string `json:"instId"` // Instrument ID ex.: BTC-USDT
	}

	// OkxCandleResponse defines the response structure of a Okx candle request.
	OkxCandleResponse struct {
		Data [][]string `json:"data"`
		ID   OkxID      `json:"arg"`
	}

	// OkxSubscriptionTopic Topic with the ticker to be subscribed/unsubscribed.
	OkxSubscriptionTopic struct {
		Channel string `json:"channel"` // Channel name ex.: tickers
		InstID  string `json:"instId"`  // Instrument ID ex.: BTC-USDT
	}

	// OkxSubscriptionMsg Message to subscribe/unsubscribe with N Topics.
	OkxSubscriptionMsg struct {
		Op   string                 `json:"op"` // Operation ex.: subscribe
		Args []OkxSubscriptionTopic `json:"args"`
	}

	// OkxPairsSummary defines the response structure for an Okx pairs summary.
	OkxPairsSummary struct {
		Data []OkxInstID `json:"data"`
	}
)

// NewOkxProvider creates a new OkxProvider.
func NewOkxProvider(
	ctx context.Context,
	logger zerolog.Logger,
	endpoints Endpoint,
	pairs ...types.CurrencyPair,
) (*OkxProvider, error) {
	if endpoints.Name != ProviderOkx {
		endpoints = Endpoint{
			Name:      ProviderOkx,
			Rest:      okxRestHost,
			Websocket: okxWSHost,
		}
	}

	wsURL := url.URL{
		Scheme: "wss",
		Host:   endpoints.Websocket,
		Path:   okxWSPath,
	}

	okxLogger := logger.With().Str("provider", string(ProviderOkx)).Logger()

	provider := &OkxProvider{
		logger:          okxLogger,
		endpoints:       endpoints,
		tickers:         map[string]OkxTickerPair{},
		candles:         map[string][]OkxCandlePair{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	provider.setSubscribedPairs(pairs...)

	controller := NewWebsocketController(
		ctx,
		ProviderOkx,
		wsURL,
		provider.getSubscriptionMsgs(),
		provider.messageReceived,
		defaultPingDuration,
		websocket.PingMessage,
		okxLogger,
	)
	go controller.Start()

	return provider, nil
}

func (p *OkxProvider) getSubscriptionMsgs() []interface{} {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	subscriptionMsgs := make([]interface{}, 0, len(p.subscribedPairs)*2)
	for _, cp := range p.subscribedPairs {
		okxPair := currencyPairToOkxPair(cp)
		okxTopic := newOkxCandleSubscriptionTopic(okxPair)
		subscriptionMsgs = append(subscriptionMsgs, newOkxSubscriptionMsg(okxTopic))

		okxTopic = newOkxTickerSubscriptionTopic(okxPair)
		subscriptionMsgs = append(subscriptionMsgs, newOkxSubscriptionMsg(okxTopic))
	}
	return subscriptionMsgs
}

// GetTickerPrices returns the tickerPrices based on the saved map.
func (p *OkxProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
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
func (p *OkxProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candlePrices := make(map[string][]types.CandlePrice, len(pairs))

	for _, currencyPair := range pairs {
		candles, err := p.getCandlePrices(currencyPair)
		if err != nil {
			return nil, err
		}

		candlePrices[currencyPair.String()] = candles
	}

	return candlePrices, nil
}

// SubscribeCurrencyPairs subscribe all currency pairs into ticker and candle channels.
func (p *OkxProvider) SubscribeCurrencyPairs(cps ...types.CurrencyPair) error {
	return nil // handled by the websocket controller
}

func (p *OkxProvider) getTickerPrice(cp types.CurrencyPair) (types.TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	instrumentID := currencyPairToOkxPair(cp)
	tickerPair, ok := p.tickers[instrumentID]
	if !ok {
		return types.TickerPrice{}, fmt.Errorf("okx failed to get ticker price for %s", instrumentID)
	}

	return tickerPair.toTickerPrice()
}

func (p *OkxProvider) getCandlePrices(cp types.CurrencyPair) ([]types.CandlePrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	instrumentID := currencyPairToOkxPair(cp)
	candles, ok := p.candles[instrumentID]
	if !ok {
		return []types.CandlePrice{}, fmt.Errorf("okx failed to get candle prices for %s", instrumentID)
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

func (p *OkxProvider) messageReceived(messageType int, bz []byte) {
	var (
		tickerResp OkxTickerResponse
		tickerErr  error
		candleResp OkxCandleResponse
		candleErr  error
	)

	// sometimes the message received is not a ticker or a candle response.
	tickerErr = json.Unmarshal(bz, &tickerResp)
	if tickerResp.ID.Channel == "tickers" {
		for _, tickerPair := range tickerResp.Data {
			p.setTickerPair(tickerPair)
			telemetryWebsocketMessage(ProviderOkx, MessageTypeTicker)
		}
		return
	}

	candleErr = json.Unmarshal(bz, &candleResp)
	if candleResp.ID.Channel == "candle1m" {
		for _, candlePair := range candleResp.Data {
			p.setCandlePair(candlePair, candleResp.ID.InstID)
			telemetryWebsocketMessage(ProviderOkx, MessageTypeCandle)
		}
		return
	}

	p.logger.Error().
		Int("length", len(bz)).
		AnErr("ticker", tickerErr).
		AnErr("candle", candleErr).
		Msg("Error on receive message")
}

func (p *OkxProvider) setTickerPair(tickerPair OkxTickerPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.tickers[tickerPair.InstID] = tickerPair
}

func (p *OkxProvider) setCandlePair(pairData []string, instID string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	ts, err := strconv.ParseInt(pairData[0], 10, 64)
	if err != nil {
		return
	}
	// the candlesticks channel uses an array of strings.
	candle := OkxCandlePair{
		Close:     pairData[4],
		InstID:    instID,
		Volume:    pairData[5],
		TimeStamp: ts,
	}
	staleTime := PastUnixTime(providerCandlePeriod)
	candleList := []OkxCandlePair{}

	candleList = append(candleList, candle)
	for _, c := range p.candles[instID] {
		if staleTime < c.TimeStamp {
			candleList = append(candleList, c)
		}
	}
	p.candles[instID] = candleList
}

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *OkxProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

// GetAvailablePairs return all available pairs symbol to subscribe.
func (p *OkxProvider) GetAvailablePairs() (map[string]struct{}, error) {
	resp, err := http.Get(p.endpoints.Rest + okxRestPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary struct {
		Data []OkxInstID `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return nil, err
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary.Data))
	for _, pair := range pairsSummary.Data {
		splitInstID := strings.Split(pair.InstID, "-")
		if len(splitInstID) != 2 {
			continue
		}

		cp := types.CurrencyPair{
			Base:  strings.ToUpper(splitInstID[0]),
			Quote: strings.ToUpper(splitInstID[1]),
		}
		availablePairs[cp.String()] = struct{}{}
	}

	return availablePairs, nil
}

func (ticker OkxTickerPair) toTickerPrice() (types.TickerPrice, error) {
	return types.NewTickerPrice(string(ProviderOkx), ticker.InstID, ticker.Last, ticker.Vol24h)
}

func (candle OkxCandlePair) toCandlePrice() (types.CandlePrice, error) {
	return types.NewCandlePrice(string(ProviderOkx), candle.InstID, candle.Close, candle.Volume, candle.TimeStamp)
}

// currencyPairToOkxPair returns the expected pair instrument ID for Okx
// ex.: "BTC-USDT".
func currencyPairToOkxPair(pair types.CurrencyPair) string {
	return pair.Base + "-" + pair.Quote
}

// newOkxTickerSubscriptionTopic returns a new subscription topic.
func newOkxTickerSubscriptionTopic(instID string) OkxSubscriptionTopic {
	return OkxSubscriptionTopic{
		Channel: "tickers",
		InstID:  instID,
	}
}

// newOkxSubscriptionTopic returns a new subscription topic.
func newOkxCandleSubscriptionTopic(instID string) OkxSubscriptionTopic {
	return OkxSubscriptionTopic{
		Channel: "candle1m",
		InstID:  instID,
	}
}

// newOkxSubscriptionMsg returns a new subscription Msg for Okx.
func newOkxSubscriptionMsg(args ...OkxSubscriptionTopic) OkxSubscriptionMsg {
	return OkxSubscriptionMsg{
		Op:   "subscribe",
		Args: args,
	}
}
