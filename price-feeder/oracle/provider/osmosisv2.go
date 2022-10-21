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

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	osmosisV2WSHost   = "api.osmo-api.network.umee.cc"
	osmosisV2WSPath   = "ws"
	osmosisV2RestHost = "https://api-osmosis.imperator.co"
	osmosisV2RestPath = "/pairs/v1/summary"
)

var _ Provider = (*OsmosisV2Provider)(nil)

type (
	// OsmosisV2Provider defines an Oracle provider implemented by UMEE's
	// Osmosis API.
	//
	// REF: https://github.com/umee-network/osmosis-api
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

	OsmosisV2Ticker struct {
		Price  string `json:"Price"`
		Volume string `json:"Volume"`
	}

	OsmosisV2Candle struct {
		Close   string `json:"Close"`
		Volume  string `json:"Volume"`
		EndTime int64  `json:"EndTime"`
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
		logger:          logger.With().Str("provider", "osmosisv2").Logger(),
		endpoints:       endpoints,
		tickers:         map[string]types.TickerPrice{},
		candles:         map[string][]types.CandlePrice{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	go provider.handleWebSocketMsgs(ctx, pairs...)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the saved map.
func (p *OsmosisV2Provider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	tickerPrices := make(map[string]types.TickerPrice, len(pairs))

	for _, cp := range pairs {
		key := currencyPairToOsmosisV2Pair(cp)
		price, err := p.getTickerPrice(key)
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
		key := currencyPairToOsmosisV2Pair(cp)
		prices, err := p.getCandlePrices(key)
		if err != nil {
			return nil, err
		}
		candlePrices[cp.String()] = prices
	}

	return candlePrices, nil
}

func (p *OsmosisV2Provider) getTickerPrice(key string) (types.TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	ticker, ok := p.tickers[key]
	if !ok {
		return types.TickerPrice{}, fmt.Errorf(
			types.ErrTickerNotFound.Error(),
			ProviderOsmosisV2,
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
			ProviderOsmosisV2,
			key,
		)
	}

	return candles, nil
}

// SubscribeCurrencyPairs performs a no-op since the osmosis-api does not
// have specific currency channels.
func (p *OsmosisV2Provider) SubscribeCurrencyPairs(pairs ...types.CurrencyPair) error {
	return nil
}

func (p *OsmosisV2Provider) messageReceived(messageType int, bz []byte, reconnectTicker *time.Ticker, pairs ...types.CurrencyPair) {
	if messageType != websocket.TextMessage {
		return
	}

	// check if message is an ack first
	if string(bz) == "ack" {
		return
	}

	var (
		messageResp map[string]interface{}
		messageErr  error
		tickerResp  OsmosisV2Ticker
		tickerErr   error
		candleResp  OsmosisV2Candle
		candleErr   error
	)

	messageErr = json.Unmarshal(bz, &messageResp)

	// Check the response for currency pairs that the provider is subscribed
	// to and determine whether it is a ticker or candle.
	for _, pair := range pairs {
		osmosisV2Pair := currencyPairToOsmosisV2Pair(pair)
		if msg, ok := messageResp[osmosisV2Pair]; ok {
			switch v := msg.(type) {
			// ticker response
			case map[string]interface{}:
				tickerString, _ := json.Marshal(v)
				tickerErr = json.Unmarshal(tickerString, &tickerResp)
				if tickerErr != nil {
					p.logger.Error().
						Int("length", len(bz)).
						AnErr("ticker", tickerErr).
						Msg("Error on receive message")

					return
				}
				p.setTickerPair(
					osmosisV2Pair,
					tickerResp,
				)
				telemetryWebsocketMessage(ProviderOsmosisV2, MessageTypeTicker)

				return

			// candle response
			case []interface{}:
				// use latest candlestick in list
				candleString, _ := json.Marshal(v[len(v)-1].(map[string]interface{}))
				candleErr = json.Unmarshal(candleString, &candleResp)
				if candleErr != nil {
					p.logger.Error().
						Int("length", len(bz)).
						AnErr("candle", candleErr).
						Msg("Error on receive message")

					return
				}
				p.setCandlePair(
					osmosisV2Pair,
					candleResp,
				)
				telemetryWebsocketMessage(ProviderOsmosisV2, MessageTypeCandle)

				return
			}
		}
	}

	p.logger.Error().
		Int("length", len(bz)).
		AnErr("message", messageErr).
		Msg("Error on receive message")
}

// ping to check websocket connection
func (p *OsmosisV2Provider) ping() error {
	return p.wsClient.WriteMessage(websocket.PingMessage, ping)
}

func (p *OsmosisV2Provider) setTickerPair(symbol string, tickerPair OsmosisV2Ticker) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	price, err := sdk.NewDecFromStr(tickerPair.Price)
	if err != nil {
		p.logger.Warn().Err(err).Msg("osmosisv2: failed to parse ticker price")
		return
	}
	volume, err := sdk.NewDecFromStr(tickerPair.Volume)
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

	close, err := sdk.NewDecFromStr(candlePair.Close)
	if err != nil {
		p.logger.Warn().Err(err).Msg("osmosisv2: failed to parse candle close")
		return
	}
	volume, err := sdk.NewDecFromStr(candlePair.Volume)
	if err != nil {
		p.logger.Warn().Err(err).Msg("osmosisv2: failed to parse candle volume")
		return
	}
	candle := types.CandlePrice{
		Price:  close,
		Volume: volume,
		// convert seconds -> milli
		TimeStamp: SecondsToMilli(candlePair.EndTime),
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

func (p *OsmosisV2Provider) handleWebSocketMsgs(ctx context.Context, pairs ...types.CurrencyPair) {
	reconnectTicker := time.NewTicker(defaultMaxConnectionTime)
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

			p.messageReceived(messageType, bz, reconnectTicker, pairs...)

		case <-reconnectTicker.C:
			if err := p.reconnect(); err != nil {
				p.logger.Err(err).Msg("error reconnecting")
			}
		}
	}
}

// reconnect closes the last WS connection then create a new one
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

	return err
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

// currencyPairToOsmosisV2Pair receives a currency pair and return osmosisv2
// ticker symbol atomusdt@ticker.
func currencyPairToOsmosisV2Pair(cp types.CurrencyPair) string {
	return strings.ToUpper(cp.Base + "/" + cp.Quote)
}
