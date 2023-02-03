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
	cryptoWSHost             = "stream.crypto.com"
	cryptoWSPath             = "/v2/market"
	cryptoReconnectTime      = time.Second * 30
	cryptoRestHost           = "https://api.crypto.com"
	cryptoRestPath           = "/v2/public/get-ticker"
	cryptoTickerChannel      = "ticker"
	cryptoCandleChannel      = "candlestick"
	cryptoHeartbeatMethod    = "public/heartbeat"
	cryptoHeartbeatReqMethod = "public/respond-heartbeat"
	cryptoTickerMsgPrefix    = "ticker."
	cryptoCandleMsgPrefix    = "candlestick.5m."
)

var _ Provider = (*CryptoProvider)(nil)

type (
	// CryptoProvider defines an Oracle provider implemented by the Crypto.com public
	// API.
	//
	// REF: https://exchange-docs.crypto.com/spot/index.html#introduction
	CryptoProvider struct {
		wsc             *WebsocketController
		logger          zerolog.Logger
		mtx             sync.RWMutex
		endpoints       Endpoint
		tickers         map[string]types.TickerPrice   // Symbol => TickerPrice
		candles         map[string][]types.CandlePrice // Symbol => CandlePrice
		subscribedPairs map[string]types.CurrencyPair  // Symbol => types.CurrencyPair
	}

	CryptoTickerResponse struct {
		Result CryptoTickerResult `json:"result"`
	}
	CryptoTickerResult struct {
		InstrumentName string         `json:"instrument_name"` // ex.: ATOM_USDT
		Channel        string         `json:"channel"`         // ex.: ticker
		Data           []CryptoTicker `json:"data"`            // ticker data
	}
	CryptoTicker struct {
		InstrumentName string `json:"i"` // Instrument Name, e.g. BTC_USDT, ETH_CRO, etc.
		Volume         string `json:"v"` // The total 24h traded volume
		LatestTrade    string `json:"a"` // The price of the latest trade, null if there weren't any trades
	}

	CryptoCandleResponse struct {
		Result CryptoCandleResult `json:"result"`
	}
	CryptoCandleResult struct {
		InstrumentName string         `json:"instrument_name"` // ex.: ATOM_USDT
		Channel        string         `json:"channel"`         // ex.: candlestick
		Data           []CryptoCandle `json:"data"`            // candlestick data
	}
	CryptoCandle struct {
		Close     string `json:"c"` // Price at close
		Volume    string `json:"v"` // Volume during interval
		Timestamp int64  `json:"t"` // End time of candlestick (Unix timestamp)
	}

	CryptoSubscriptionMsg struct {
		ID     int64                    `json:"id"`
		Method string                   `json:"method"` // subscribe, unsubscribe
		Params CryptoSubscriptionParams `json:"params"`
		Nonce  int64                    `json:"nonce"` // Current timestamp (milliseconds since the Unix epoch)
	}
	CryptoSubscriptionParams struct {
		Channels []string `json:"channels"` // Channels to be subscribed ex. ticker.ATOM_USDT
	}

	CryptoPairsSummary struct {
		Result CryptoInstruments `json:"result"`
	}
	CryptoInstruments struct {
		Data []CryptoTicker `json:"data"`
	}

	CryptoHeartbeatResponse struct {
		ID     int64  `json:"id"`
		Method string `json:"method"` // public/heartbeat
	}
	CryptoHeartbeatRequest struct {
		ID     int64  `json:"id"`
		Method string `json:"method"` // public/respond-heartbeat
	}
)

func NewCryptoProvider(
	ctx context.Context,
	logger zerolog.Logger,
	endpoints Endpoint,
	pairs ...types.CurrencyPair,
) (*CryptoProvider, error) {
	if endpoints.Name != ProviderCrypto {
		endpoints = Endpoint{
			Name:      ProviderCrypto,
			Rest:      cryptoRestHost,
			Websocket: cryptoWSHost,
		}
	}

	wsURL := url.URL{
		Scheme: "wss",
		Host:   endpoints.Websocket,
		Path:   cryptoWSPath,
	}

	cryptoLogger := logger.With().Str("provider", "crypto").Logger()

	provider := &CryptoProvider{
		logger:          cryptoLogger,
		endpoints:       endpoints,
		tickers:         map[string]types.TickerPrice{},
		candles:         map[string][]types.CandlePrice{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	provider.setSubscribedPairs(pairs...)

	provider.wsc = NewWebsocketController(
		ctx,
		endpoints.Name,
		wsURL,
		provider.getSubscriptionMsgs(pairs...),
		provider.messageReceived,
		disabledPingDuration,
		websocket.PingMessage,
		cryptoLogger,
	)
	provider.wsc.StartConnections()

	return provider, nil
}

func (p *CryptoProvider) getSubscriptionMsgs(cps ...types.CurrencyPair) []interface{} {
	subscriptionMsgs := make([]interface{}, 0, len(cps)*2)
	for _, cp := range cps {
		cryptoPair := currencyPairToCryptoPair(cp)
		channel := cryptoTickerMsgPrefix + cryptoPair
		msg := newCryptoSubscriptionMsg([]string{channel})
		subscriptionMsgs = append(subscriptionMsgs, msg)

		cryptoPair = currencyPairToCryptoPair(cp)
		channel = cryptoCandleMsgPrefix + cryptoPair
		msg = newCryptoSubscriptionMsg([]string{channel})
		subscriptionMsgs = append(subscriptionMsgs, msg)
	}
	return subscriptionMsgs
}

// SubscribeCurrencyPairs sends the new subscription messages to the websocket
// and adds them to the providers subscribedPairs array
func (p *CryptoProvider) SubscribeCurrencyPairs(cps ...types.CurrencyPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	newPairs := []types.CurrencyPair{}
	for _, cp := range cps {
		if _, ok := p.subscribedPairs[cp.String()]; !ok {
			newPairs = append(newPairs, cp)
		}
	}

	newSubscriptionMsgs := p.getSubscriptionMsgs(newPairs...)
	p.wsc.AddWebsocketConnection(
		newSubscriptionMsgs,
		p.messageReceived,
		disabledPingDuration,
		websocket.PingMessage,
	)
	p.setSubscribedPairs(newPairs...)
}

// GetTickerPrices returns the tickerPrices based on the provided pairs.
func (p *CryptoProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	tickerPrices := make(map[string]types.TickerPrice, len(pairs))

	tickerErrs := 0
	for _, cp := range pairs {
		key := currencyPairToCryptoPair(cp)
		price, err := p.getTickerPrice(key)
		if err != nil {
			p.logger.Warn().Err(err)
			tickerErrs++
			continue
		}
		tickerPrices[cp.String()] = price
	}

	if tickerErrs == len(pairs) {
		return nil, fmt.Errorf(
			types.ErrNoTickers.Error(),
			p.endpoints.Name,
			pairs,
		)
	}
	return tickerPrices, nil
}

// GetCandlePrices returns the candlePrices based on the provided pairs.
func (p *CryptoProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candlePrices := make(map[string][]types.CandlePrice, len(pairs))

	candleErrs := 0
	for _, cp := range pairs {
		key := currencyPairToCryptoPair(cp)
		prices, err := p.getCandlePrices(key)
		if err != nil {
			p.logger.Warn().Err(err)
			candleErrs++
			continue
		}
		candlePrices[cp.String()] = prices
	}

	if candleErrs == len(pairs) {
		return nil, fmt.Errorf(
			types.ErrNoCandles.Error(),
			p.endpoints.Name,
			pairs,
		)
	}
	return candlePrices, nil
}

func (p *CryptoProvider) getTickerPrice(key string) (types.TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	ticker, ok := p.tickers[key]
	if !ok {
		return types.TickerPrice{}, fmt.Errorf(
			types.ErrTickerNotFound.Error(),
			p.endpoints.Name,
			key,
		)
	}

	return ticker, nil
}

func (p *CryptoProvider) getCandlePrices(key string) ([]types.CandlePrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	candles, ok := p.candles[key]
	if !ok {
		return []types.CandlePrice{}, fmt.Errorf(
			types.ErrCandleNotFound.Error(),
			p.endpoints.Name,
			key,
		)
	}

	candleList := []types.CandlePrice{}
	candleList = append(candleList, candles...)

	return candleList, nil
}

func (p *CryptoProvider) messageReceived(messageType int, conn *WebsocketConnection, bz []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var (
		heartbeatResp CryptoHeartbeatResponse
		heartbeatErr  error
		tickerResp    CryptoTickerResponse
		tickerErr     error
		candleResp    CryptoCandleResponse
		candleErr     error
	)

	// sometimes the message received is not a ticker or a candle response.
	heartbeatErr = json.Unmarshal(bz, &heartbeatResp)
	if heartbeatResp.Method == cryptoHeartbeatMethod {
		p.pong(conn, heartbeatResp)
		return
	}

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
		AnErr("heartbeat", heartbeatErr).
		AnErr("ticker", tickerErr).
		AnErr("candle", candleErr).
		Msg("Error on receive message")
}

// pongReceived return a heartbeat message when a "ping" is received and reset the
// recconnect ticker because the connection is alive. After connected to crypto.com's
// Websocket server, the server will send heartbeat periodically (30s interval).
// When client receives an heartbeat message, it must respond back with the
// public/respond-heartbeat method, using the same matching id,
// within 5 seconds, or the connection will break.
func (p *CryptoProvider) pong(conn *WebsocketConnection, heartbeatResp CryptoHeartbeatResponse) {
	heartbeatReq := CryptoHeartbeatRequest{
		ID:     heartbeatResp.ID,
		Method: cryptoHeartbeatReqMethod,
	}

	if err := conn.SendJSON(heartbeatReq); err != nil {
		p.logger.Err(err).Msg("could not send pong message back")
	}
}

func (p *CryptoProvider) setTickerPair(symbol string, tickerPair CryptoTicker) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	tickerPrice, err := types.NewTickerPrice(
		string(ProviderCrypto),
		symbol,
		tickerPair.LatestTrade,
		tickerPair.Volume,
	)
	if err != nil {
		p.logger.Warn().Err(err).Msg("crypto: failed to parse ticker")
		return
	}

	p.tickers[symbol] = tickerPrice
}

func (p *CryptoProvider) setCandlePair(symbol string, candlePair CryptoCandle) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	candle, err := types.NewCandlePrice(
		string(ProviderCrypto),
		symbol,
		candlePair.Close,
		candlePair.Volume,
		SecondsToMilli(candlePair.Timestamp),
	)
	if err != nil {
		p.logger.Warn().Err(err).Msg("crypto: failed to parse candle")
		return
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

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *CryptoProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

// GetAvailablePairs returns all pairs to which the provider can subscribe.
// ex.: map["ATOMUSDT" => {}, "UMEEUSDC" => {}].
func (p *CryptoProvider) GetAvailablePairs() (map[string]struct{}, error) {
	resp, err := http.Get(p.endpoints.Rest + cryptoRestPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary CryptoPairsSummary
	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return nil, err
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary.Result.Data))
	for _, pair := range pairsSummary.Result.Data {
		splitInstName := strings.Split(pair.InstrumentName, "_")
		if len(splitInstName) != 2 {
			continue
		}

		cp := types.CurrencyPair{
			Base:  strings.ToUpper(splitInstName[0]),
			Quote: strings.ToUpper(splitInstName[1]),
		}

		availablePairs[cp.String()] = struct{}{}
	}

	return availablePairs, nil
}

// currencyPairToCryptoPair receives a currency pair and return crypto
// ticker symbol atomusdt@ticker.
func currencyPairToCryptoPair(cp types.CurrencyPair) string {
	return strings.ToUpper(cp.Base + "_" + cp.Quote)
}

// newCryptoSubscriptionMsg returns a new subscription Msg.
func newCryptoSubscriptionMsg(channels []string) CryptoSubscriptionMsg {
	return CryptoSubscriptionMsg{
		ID:     1,
		Method: "subscribe",
		Params: CryptoSubscriptionParams{
			Channels: channels,
		},
		Nonce: time.Now().UnixMilli(),
	}
}
