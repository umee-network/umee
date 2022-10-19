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
	"github.com/umee-network/umee/price-feeder/oracle/types"

	"github.com/umee-network/umee/v3/util/coin"
)

const (
	osmosisV2WSHost    = ""
	osmosisV2WSPath    = ""
	osmosisV2RestHost  = "https://api-osmosis.imperator.co"
	osmosisV2RestPath  = "/pairs/v1/summary"
)

var _ Provider = (*OsmosisV2Provider)(nil)

type (
	// OsmosisProvider defines an Oracle provider implemented by the Osmosis public
	// API.
	//
	// REF: https://api-osmosis.imperator.co/swagger/
	OsmosisV2Provider struct {
		wsURL           url.URL
		wsClient        *websocket.Conn
		logger          zerolog.Logger
		mtx             sync.RWMutex
		endpoints       Endpoint
		tickers         map[string]types.TickerPrice   // Symbol => TickerPrice
		candles         map[string][]types.CandlePrice // Symbol => CandlePrice
		subscribedPairs map[string]types.CurrencyPair  // Symbol => types.CurrencyPair
	}

	OsmosisV2TickerResponse struct {
		Pool        string          `json:"pool"`         // ex.: ATOMUSDT
		TickerPrice OsmosisV2Ticker `json:"ticker_price"`
	}
	OsmosisV2Ticker struct {
		Price  float64 `json:"price"`
		Volume float64 `json:"volume"`
	}

	OsmosisV2CandleResponse struct {
		Pool        string          `json:"pool"`         // ex.: ATOMUSDT
		TickerPrice OsmosisV2Candle `json:"candle_price"`
	}
	OsmosisV2Candle struct {
		Time   int64   `json:"time"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}

	OsmosisV2SubscriptionMsg struct {

	}

	// OsmosisV2PairsSummary defines the response structure for an Osmosis pairs
	// summary.
	OsmosisV2PairsSummary struct {
		Data []OsmosisPairData `json:"data"`
	}

	// OsmosisV2PairData defines the data response structure for an Osmosis pair.
	OsmosisV2PairData struct {
		Base  string `json:"base_symbol"`
		Quote string `json:"quote_symbol"`
	}
)

func NewOsmosisV2Provider(
	ctx context.Context,
	logger zerolog.Logger,
	endpoints Endpoint,
	pairs ...types.CurrencyPair,
) (*OsmosisV2Provider, error) {
	if endpoints.Name != ProviderOsmosisV2 {
		endpoints = Endpoint{
			Name:      ProviderOsmosisV2,
			Rest:      osmosisV2RestHost,
			Websocket: osmosisV2WSHost,
		}
	}

	wsURL := url.URL{
		Scheme: "wss",
		Host:   endpoints.Websocket,
		Path:   osmosisV2WSPath,
	}

	wsConn, resp, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf(
			types.ErrWebsocketDial.Error(),
			ProviderOsmosisV2,
			err,
		)
	}
	defer resp.Body.Close()

	provider := &OsmosisV2Provider{
		wsURL:           wsURL,
		wsClient:        wsConn,
		logger:          logger.With().Str("provider", "crypto").Logger(),
		endpoints:       endpoints,
		tickers:         map[string]types.TickerPrice{},
		candles:         map[string][]types.CandlePrice{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	if err := provider.SubscribeCurrencyPairs(pairs...); err != nil {
		return nil, err
	}

	go provider.handleWebSocketMsgs(ctx)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the saved map.
func (p *OsmosisV2Provider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	tickerPrices := make(map[string]types.TickerPrice, len(pairs))

	for _, cp := range pairs {
		price, err := p.getTickerPrice(cp.String())
		if err != nil {
			return nil, err
		}
		tickerPrices[cp.String()] = price
	}

	return tickerPrices, nil
}

// GetCandlePrices returns the candlePrices based on the saved map
func (p *OsmosisV2Provider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candlePrices := make(map[string][]types.CandlePrice, len(pairs))

	for _, cp := range pairs {
		prices, err := p.getCandlePrices(cp.String())
		if err != nil {
			return nil, err
		}
		candlePrices[cp.String()] = prices
	}

	return candlePrices, nil
}

// SubscribeCurrencyPairs subscribe all currency pairs into ticker and candle channels.
func (p *OsmosisV2Provider) SubscribeCurrencyPairs(cps ...types.CurrencyPair) error {
	if len(cps) == 0 {
		return fmt.Errorf("currency pairs is empty")
	}

	if err := p.subscribeChannels(cps...); err != nil {
		return err
	}

	p.setSubscribedPairs(cps...)
	telemetryWebsocketSubscribeCurrencyPairs(ProviderCrypto, len(cps))
	return nil
}

// subscribeChannels subscribe all currency pairs into ticker and candle channels.
func (p *OsmosisV2Provider) subscribeChannels(cps ...types.CurrencyPair) error {
	if err := p.subscribeTickers(cps...); err != nil {
		return err
	}

	return p.subscribeCandles(cps...)
}

// subscribeTickers subscribe all currency pairs into ticker channel.
func (p *OsmosisV2Provider) subscribeTickers(cps ...types.CurrencyPair) error {
	pairs := make([]string, len(cps))

	for i, cp := range cps {
		pairs[i] = cp.String()
	}

	subsMsg := newOsmosisV2SubscriptionMsg()
	err := p.wsClient.WriteJSON(subsMsg)

	return err
}

// subscribeCandles subscribe all currency pairs into candle channel.
func (p *OsmosisV2Provider) subscribeCandles(cps ...types.CurrencyPair) error {
	pairs := make([]string, len(cps))

	for i, cp := range cps {
		pairs[i] = cp.String()
	}

	subsMsg := newOsmosisV2SubscriptionMsg()
	err := p.wsClient.WriteJSON(subsMsg)

	return err
}

// subscribedPairsToSlice returns the map of subscribed pairs as a slice.
func (p *OsmosisV2Provider) subscribedPairsToSlice() []types.CurrencyPair {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	return types.MapPairsToSlice(p.subscribedPairs)
}

func (p *OsmosisV2Provider) getTickerPrice(key string) (types.TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	ticker, ok := p.tickers[key]
	if !ok {
		return types.TickerPrice{}, fmt.Errorf(
			types.ErrTickerNotFound.Error(),
			ProviderCrypto,
			key,
		)
	}

	return ticker, nil
}

func (p *OsmosisV2Provider) getCandlePrices(key string) ([]types.CandlePrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	candles, ok := p.candles[key]
	if !ok {
		return []types.CandlePrice{}, fmt.Errorf(
			types.ErrCandleNotFound.Error(),
			ProviderCrypto,
			key,
		)
	}

	return candles, nil
}

func (p *OsmosisV2Provider) messageReceived(messageType int, bz []byte, reconnectTicker *time.Ticker) {
	if messageType != websocket.TextMessage {
		return
	}

	var (
		tickerResp    OsmosisV2TickerResponse
		tickerErr     error
		candleResp    OsmosisV2CandleResponse
		candleErr     error
	)

	tickerErr = json.Unmarshal(bz, &tickerResp)
	if tickerResp.Result.Channel == cryptoTickerChannel {
		for _, tickerPair := range tickerResp.Result.Data {
			p.setTickerPair(
				tickerResp.Result.InstrumentName,
				tickerPair,
			)
			telemetryWebsocketMessage(ProviderCrypto, MessageTypeTicker)
		}
		return
	}

	candleErr = json.Unmarshal(bz, &candleResp)
	if candleResp.Result.Channel == cryptoCandleChannel {
		for _, candlePair := range candleResp.Result.Data {
			p.setCandlePair(
				candleResp.Result.InstrumentName,
				candlePair,
			)
			telemetryWebsocketMessage(ProviderCrypto, MessageTypeCandle)
		}
		return
	}

	p.logger.Error().
		Int("length", len(bz)).
		AnErr("ticker", tickerErr).
		AnErr("candle", candleErr).
		Msg("Error on receive message")
}

// ping to check websocket connection
func (p *OsmosisV2Provider) ping() error {
	return p.wsClient.WriteMessage(websocket.PingMessage, ping)
}

func (p *OsmosisV2Provider) setTickerPair(symbol string, tickerPair OsmosisV2Ticker) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	price, err := coin.NewDecFromFloat(tickerPair.Price)
	if err != nil {
		p.logger.Warn().Err(err).Msg("osmosisv2: failed to parse ticker price")
		return
	}
	volume, err := coin.NewDecFromFloat(tickerPair.Volume)
	if err != nil {
		p.logger.Warn().Err(err).Msg("osmosisv2: failed to parse ticker volume")
		return
	}

	p.tickers[symbol] = types.TickerPrice{
		Price:  price,
		Volume: volume,
	}
}

func (p *OsmosisV2Provider) setCandlePair(symbol string, candlePair OsmosisV2Candle) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	close, err := coin.NewDecFromFloat(candlePair.Close)
	if err != nil {
		p.logger.Warn().Err(err).Msg("osmosisv2: failed to parse candle close")
		return
	}
	volume, err := coin.NewDecFromFloat(candlePair.Volume)
	if err != nil {
		p.logger.Warn().Err(err).Msg("osmosisv2: failed to parse candle volume")
		return
	}
	candle := types.CandlePrice{
		Price:  close,
		Volume: volume,
		// convert seconds -> milli
		TimeStamp: SecondsToMilli(candlePair.Time),
	}

	staleTime := PastUnixTime(providerCandlePeriod)
	candleList := []types.CandlePrice{}
	candleList = append(candleList, candle)

	for _, c := range p.candles[symbol] {
		if staleTime < c.TimeStamp {
			candleList = append(candleList, c)
		}
	}

	p.candles[symbol] = candleList
}

func (p *OsmosisV2Provider) handleWebSocketMsgs(ctx context.Context) {
	reconnectTicker := time.NewTicker(cryptoReconnectTime)
	defer reconnectTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(defaultReadNewWSMessage):
			messageType, bz, err := p.wsClient.ReadMessage()
			if err != nil {
				// If some error occurs, check if connection is alive
				// and continue to try to read the next message.
				p.logger.Err(err).Msg("failed to read message")
				if err := p.ping(); err != nil {
					p.logger.Err(err).Msg("failed to send ping")
					if err := p.reconnect(); err != nil {
						p.logger.Err(err).Msg("error reconnecting websocket")
					}
				}
				continue
			}

			if len(bz) == 0 {
				continue
			}

			p.messageReceived(messageType, bz, reconnectTicker)

		case <-reconnectTicker.C:
			if err := p.reconnect(); err != nil {
				p.logger.Err(err).Msg("error reconnecting")
			}
		}
	}
}

// reconnect closes the last WS connection then create a new one and subscribes to
// all subscribed pairs in the ticker and candle pairs. If no ping is received
// within 1 minute, the connection will be disconnected. It is recommended to
// send a ping for 10-20 seconds
func (p *OsmosisV2Provider) reconnect() error {
	err := p.wsClient.Close()
	if err != nil {
		p.logger.Err(err).Msg("error closing osmosisv2 websocket")
	}

	p.logger.Debug().Msg("osmosisv2: reconnecting websocket")

	wsConn, resp, err := websocket.DefaultDialer.Dial(p.wsURL.String(), nil)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf(
			types.ErrWebsocketDial.Error(),
			ProviderOsmosisV2,
			err,
		)
	}
	p.wsClient = wsConn
	telemetryWebsocketReconnect(ProviderOsmosisV2)

	return p.subscribeChannels(p.subscribedPairsToSlice()...)
}

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *OsmosisV2Provider) setSubscribedPairs(cps ...types.CurrencyPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

// GetAvailablePairs returns all pairs to which the provider can subscribe.
// ex.: map["ATOMUSDT" => {}, "UMEEUSDC" => {}].
func (p *OsmosisV2Provider) GetAvailablePairs() (map[string]struct{}, error) {
	resp, err := http.Get(p.endpoints.Rest + osmosisV2RestPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary OsmosisV2PairsSummary
	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return nil, err
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary.Data))
	for _, pair := range pairsSummary.Data {
		cp := types.CurrencyPair{
			Base:  strings.ToUpper(pair.Base),
			Quote: strings.ToUpper(pair.Quote),
		}
		availablePairs[cp.String()] = struct{}{}
	}

	return availablePairs, nil
}

// newOsmosisV2SubscriptionMsg returns a new subscription Msg.
func newOsmosisV2SubscriptionMsg() OsmosisV2SubscriptionMsg {
	return OsmosisV2SubscriptionMsg{

	}
}
