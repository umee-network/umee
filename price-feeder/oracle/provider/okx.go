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
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/umee-network/umee/price-feeder/telemetry"

	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	okxHost          = "ws.okx.com:8443"
	okxPath          = "/ws/v5/public"
	okxPingCheck     = time.Second * 28 // should be < 30
	okxPairsEndpoint = "https://www.okx.com/api/v5/market/tickers?instType=SPOT"
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
		mtx             sync.RWMutex
		tickers         map[string]OkxTickerPair      // InstId => OkxTickerPair
		candles         map[string][]OkxCandlePair    // InstId => 0kxCandlePair
		subscribedPairs map[string]types.CurrencyPair // Symbol => types.CurrencyPair
	}

	// OkxInstId defines the id Symbol of an pair.
	OkxInstId struct {
		InstID string `json:"instId"` // Instrument ID ex.: BTC-USDT
	}

	// OkxTickerPair defines a ticker pair of Okx.
	OkxTickerPair struct {
		OkxInstId
		Last   string `json:"last"`   // Last traded price ex.: 43508.9
		Vol24h string `json:"vol24h"` // 24h trading volume ex.: 11159.87127845
	}

	// OkxInst defines the structure containing ID information for the OkxResponses.
	OkxID struct {
		OkxInstId
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
		InstId    string `json:"instId"` // Instrument ID ex.: BTC-USDT
	}

	// OkxCandleResponse defines the response structure of a Okx candle request.
	OkxCandleResponse struct {
		Data [][]string `json:"data"`
		ID   OkxID      `json:"arg"`
	}

	// OkxSubscriptionTopic Topic with the ticker to be subscribed/unsubscribed.
	OkxSubscriptionTopic struct {
		Channel string `json:"channel"` // Channel name ex.: tickers
		InstId  string `json:"instId"`  // Instrument ID ex.: BTC-USDT
	}

	// OkxSubscriptionMsg Message to subscribe/unsubscribe with N Topics.
	OkxSubscriptionMsg struct {
		Op   string                 `json:"op"` // Operation ex.: subscribe
		Args []OkxSubscriptionTopic `json:"args"`
	}

	// OkxPairsSummary defines the response structure for an Okx pairs summary.
	OkxPairsSummary struct {
		Data []OkxInstId `json:"data"`
	}
)

// NewOkxProvider creates a new OkxProvider.
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
		reconnectTimer:  time.NewTicker(okxPingCheck),
		tickers:         map[string]OkxTickerPair{},
		candles:         map[string][]OkxCandlePair{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}
	provider.wsClient.SetPongHandler(provider.pongHandler)

	if err := provider.SubscribeCurrencyPairs(pairs...); err != nil {
		return nil, err
	}

	go provider.handleReceivedTickers(ctx)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the saved map.
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

// GetCandlePrices returns the candlePrices based on the saved map
func (p *OkxProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]CandlePrice, error) {
	candlePrices := make(map[string][]CandlePrice, len(pairs))

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
	if len(cps) == 0 {
		return fmt.Errorf("currency pairs is empty")
	}

	if err := p.subscribeChannels(cps...); err != nil {
		return err
	}

	p.setSubscribedPairs(cps...)
	telemetry.IncrCounter(
		float32(len(cps)),
		"websocket",
		"subscribe",
		"currency_pairs",
		"provider",
		config.ProviderOkx,
	)
	return nil
}

// subscribeChannels subscribe all currency pairs into ticker and candle channels.
func (p *OkxProvider) subscribeChannels(cps ...types.CurrencyPair) error {
	if err := p.subscribeTickers(cps...); err != nil {
		return err
	}

	return p.subscribeCandles(cps...)
}

// subscribeTickers subscribe all currency pairs into ticker channel.
func (p *OkxProvider) subscribeTickers(cps ...types.CurrencyPair) error {
	topics := make([]OkxSubscriptionTopic, len(cps))

	for i, cp := range cps {
		topics[i] = newOkxTickerSubscriptionTopic(currencyPairToOkxPair(cp))
	}

	return p.subscribePairs(topics...)
}

// subscribeCandles subscribe all currency pairs into candle channel.
func (p *OkxProvider) subscribeCandles(cps ...types.CurrencyPair) error {
	topics := make([]OkxSubscriptionTopic, len(cps))

	for i, cp := range cps {
		topics[i] = newOkxCandleSubscriptionTopic(currencyPairToOkxPair(cp))
	}

	return p.subscribePairs(topics...)
}

// subscribedPairsToSlice returns the map of subscribed pairs as slice
func (p *OkxProvider) subscribedPairsToSlice() []types.CurrencyPair {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	return types.MapPairsToSlice(p.subscribedPairs)
}

func (p *OkxProvider) getTickerPrice(cp types.CurrencyPair) (TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	instrumentId := currencyPairToOkxPair(cp)
	tickerPair, ok := p.tickers[instrumentId]
	if !ok {
		return TickerPrice{}, fmt.Errorf("okx provider failed to get ticker price for %s", instrumentId)
	}

	return tickerPair.toTickerPrice()
}

func (p *OkxProvider) getCandlePrices(cp types.CurrencyPair) ([]CandlePrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	instrumentId := currencyPairToOkxPair(cp)
	candles, ok := p.candles[instrumentId]
	if !ok {
		return []CandlePrice{}, fmt.Errorf("failed to get candle prices for %s", instrumentId)
	}
	candleList := []CandlePrice{}
	for _, candle := range candles {
		cp, err := candle.toCandlePrice()
		if err != nil {
			return []CandlePrice{}, err
		}
		candleList = append(candleList, cp)
	}

	return candleList, nil
}

func (p *OkxProvider) handleReceivedTickers(ctx context.Context) {
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

func (p *OkxProvider) messageReceived(messageType int, bz []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var (
		tickerResp OkxTickerResponse
		tickerErr  error
		candleResp OkxCandleResponse
	)

	// sometimes the message received is not a ticker or a candle response.
	tickerErr = json.Unmarshal(bz, &tickerResp)
	if tickerResp.ID.Channel == "tickers" {
		// succeeds to unmarshal ticker
		for _, tickerPair := range tickerResp.Data {
			p.setTickerPair(tickerPair)
			telemetry.IncrCounter(
				1,
				"websocket",
				"message",
				"type",
				"ticker",
				"provider",
				config.ProviderOkx,
			)
		}
		return
	}

	if err := json.Unmarshal(bz, &candleResp); err != nil {
		p.logger.Error().Err(err).Msg("could not unmarshal candle resp")
		return
	}
	if candleResp.ID.Channel == "candle1m" {
		for _, candlePair := range candleResp.Data {
			p.setCandlePair(candlePair, candleResp.ID.InstID)
			telemetry.IncrCounter(
				1,
				"websocket",
				"message",
				"type",
				"candle",
				"provider",
				config.ProviderOkx,
			)
		}
		return
	}

	if tickerErr != nil {
		p.logger.Error().Err(tickerErr).Msg("could not unmarshal ticker resp")
	}
}

func (p *OkxProvider) setTickerPair(tickerPair OkxTickerPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.tickers[tickerPair.InstID] = tickerPair
}

// subscribePairs write the subscription msg to the provider.
func (p *OkxProvider) subscribePairs(pairs ...OkxSubscriptionTopic) error {
	subsMsg := newOkxSubscriptionMsg(pairs...)
	return p.wsClient.WriteJSON(subsMsg)
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
		InstId:    instID,
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

func (p *OkxProvider) resetReconnectTimer() {
	p.reconnectTimer.Reset(okxPingCheck)
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
func (p *OkxProvider) reconnect() error {
	p.wsClient.Close()

	p.logger.Debug().Msg("reconnecting websocket")
	wsConn, _, err := websocket.DefaultDialer.Dial(p.wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("error reconnecting to Okx websocket: %w", err)
	}
	wsConn.SetPongHandler(p.pongHandler)
	p.wsClient = wsConn

	currencyPairs := p.subscribedPairsToSlice()

	telemetry.IncrCounter(
		1,
		"websocket",
		"reconnect",
		"provider",
		config.ProviderOkx,
	)
	return p.subscribeChannels(currencyPairs...)
}

// ping to check websocket connection.
func (p *OkxProvider) ping() error {
	return p.wsClient.WriteMessage(websocket.PingMessage, ping)
}

func (p *OkxProvider) pongHandler(appData string) error {
	p.resetReconnectTimer()
	return nil
}

// GetAvailablePairs return all available pairs symbol to susbscribe.
func (p *OkxProvider) GetAvailablePairs() (map[string]struct{}, error) {
	resp, err := http.Get(okxPairsEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary struct {
		Data []OkxInstId `json:"data"`
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

func (ticker OkxTickerPair) toTickerPrice() (TickerPrice, error) {
	return newTickerPrice("Okx", ticker.InstID, ticker.Last, ticker.Vol24h)
}

func (candle OkxCandlePair) toCandlePrice() (CandlePrice, error) {
	return newCandlePrice("Okx", candle.InstId, candle.Close, candle.Volume, candle.TimeStamp)
}

// currencyPairToOkxPair returns the expected pair instrument ID for Okx
// ex.: "BTC-USDT".
func currencyPairToOkxPair(pair types.CurrencyPair) string {
	return pair.Base + "-" + pair.Quote
}

// newOkxTickerSubscriptionTopic returns a new subscription topic.
func newOkxTickerSubscriptionTopic(instId string) OkxSubscriptionTopic {
	return OkxSubscriptionTopic{
		Channel: "tickers",
		InstId:  instId,
	}
}

// newOkxSubscriptionTopic returns a new subscription topic.
func newOkxCandleSubscriptionTopic(instId string) OkxSubscriptionTopic {
	return OkxSubscriptionTopic{
		Channel: "candle1m",
		InstId:  instId,
	}
}

// newOkxSubscriptionMsg returns a new subscription Msg for Okx.
func newOkxSubscriptionMsg(args ...OkxSubscriptionTopic) OkxSubscriptionMsg {
	return OkxSubscriptionMsg{
		Op:   "subscribe",
		Args: args,
	}
}
