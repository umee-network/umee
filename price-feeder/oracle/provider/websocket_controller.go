package provider

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/umee-network/umee/price-feeder/v2/oracle/types"
)

const (
	defaultReadNewWSMessage   = 50 * time.Millisecond
	defaultMaxConnectionTime  = time.Hour * 23 // should be < 24h
	defaultPingDuration       = 15 * time.Second
	disabledPingDuration      = time.Duration(0)
	startingReconnectDuration = 5 * time.Second
	maxRetryMultiplier        = 25 // max retry duration: 52m5s
)

type (
	MessageHandler func(int, []byte)

	// WebsocketController defines a provider agnostic websocket handler
	// that manages reconnecting, subscribing, and receiving messages
	WebsocketConnection struct {
		parentCtx           context.Context
		websocketCtx        context.Context
		websocketCancelFunc context.CancelFunc
		providerName        Name
		websocketURL        url.URL
		subscriptionMsg     interface{}
		messageHandler      MessageHandler
		pingDuration        time.Duration
		pingMessageType     uint
		logger              zerolog.Logger

		mtx              sync.Mutex
		client           *websocket.Conn
		reconnectCounter uint
	}

	WebsocketController struct {
		parentCtx    context.Context
		providerName Name
		websocketURL url.URL
		logger       zerolog.Logger
		connections  []*WebsocketConnection
	}
)

func NewWebsocketController(
	ctx context.Context,
	providerName Name,
	websocketURL url.URL,
	subscriptionMsgs []interface{},
	messageHandler MessageHandler,
	pingDuration time.Duration,
	pingMessageType uint,
	logger zerolog.Logger,
) *WebsocketController {
	connections := make([]*WebsocketConnection, 0)

	for _, subMsg := range subscriptionMsgs {
		connection := &WebsocketConnection{
			parentCtx:       ctx,
			providerName:    providerName,
			websocketURL:    websocketURL,
			subscriptionMsg: subMsg,
			messageHandler:  messageHandler,
			pingDuration:    pingDuration,
			pingMessageType: pingMessageType,
			logger:          logger,
		}
		connections = append(connections, connection)
	}

	return &WebsocketController{
		parentCtx:    ctx,
		providerName: providerName,
		websocketURL: websocketURL,
		logger:       logger,
		connections:  connections,
	}
}

func (wsc *WebsocketController) StartConnections() {
	for _, conn := range wsc.connections {
		go conn.start()
	}
}

// AddSubscriptionMsgs immediately sends the new subscription messages and
// adds them to the subscriptionMsgs array if successful
func (wsc *WebsocketController) AddWebsocketConnection(
	msgs []interface{},
	messageHandler MessageHandler,
	pingDuration time.Duration,
	pingMessageType uint,
) {
	for _, msg := range msgs {
		conn := &WebsocketConnection{
			parentCtx:       wsc.parentCtx,
			providerName:    wsc.providerName,
			websocketURL:    wsc.websocketURL,
			subscriptionMsg: msg,
			messageHandler:  messageHandler,
			pingDuration:    pingDuration,
			pingMessageType: pingMessageType,
			logger:          wsc.logger,
		}
		wsc.connections = append(wsc.connections, conn)
		go conn.start()
	}
}

// Start will continuously loop and attempt connecting to the websocket
// until a successful connection is made. It then starts the ping
// service and read listener in new go routines and sends subscription
// messages  using the passed in subscription messages
func (conn *WebsocketConnection) start() {
	connectTicker := time.NewTicker(time.Millisecond)
	defer connectTicker.Stop()

	for {
		if err := conn.connect(); err != nil {
			conn.logger.Err(err).Send()
			select {
			case <-conn.parentCtx.Done():
				return
			case <-connectTicker.C:
				connectTicker.Reset(conn.iterateRetryCounter())
				continue
			}
		}

		go conn.readWebSocket()
		go conn.pingLoop()

		if err := conn.subscribe(conn.subscriptionMsg); err != nil {
			conn.logger.Err(err).Send()
			conn.close()
			continue
		}
		return
	}
}

// connect dials the websocket and sets the client to the established connection
func (conn *WebsocketConnection) connect() error {
	conn.mtx.Lock()
	defer conn.mtx.Unlock()

	conn.logger.Debug().Msg("connecting to websocket")
	connection, resp, err := websocket.DefaultDialer.Dial(conn.websocketURL.String(), nil)
	if err != nil {
		return fmt.Errorf(types.ErrWebsocketDial.Error(), conn.providerName, err)
	}
	defer resp.Body.Close()
	conn.client = connection
	conn.websocketCtx, conn.websocketCancelFunc = context.WithCancel(conn.parentCtx)
	conn.client.SetPingHandler(conn.pingHandler)
	conn.reconnectCounter = 0
	return nil
}

func (conn *WebsocketConnection) iterateRetryCounter() time.Duration {
	if conn.reconnectCounter < 25 {
		conn.reconnectCounter++
	}
	multiplier := math.Pow(float64(conn.reconnectCounter), 2)
	return startingReconnectDuration * time.Duration(multiplier)
}

// subscribe sends the WebsocketControllers subscription messages to the websocket
func (conn *WebsocketConnection) subscribe(msg interface{}) error {
	telemetryWebsocketSubscribeCurrencyPairs(conn.providerName, 1)
	if err := conn.SendJSON(msg); err != nil {
		return fmt.Errorf(types.ErrWebsocketSend.Error(), conn.providerName, err)
	}
	return nil
}

// SendJSON sends a json message to the websocket connection using the Websocket
// Controller mutex to ensure multiple writes do not happen at once
func (conn *WebsocketConnection) SendJSON(msg interface{}) error {
	conn.mtx.Lock()
	defer conn.mtx.Unlock()

	if conn.client == nil {
		return fmt.Errorf("unable to send JSON on a closed connection")
	}
	conn.logger.Debug().Interface("msg", msg).Msg("sending websocket message")
	if err := conn.client.WriteJSON(msg); err != nil {
		return fmt.Errorf(types.ErrWebsocketSend.Error(), conn.providerName, err)
	}
	return nil
}

// ping sends a ping to the server every defaultPingDuration
func (conn *WebsocketConnection) pingLoop() {
	if conn.pingDuration == disabledPingDuration {
		return // disable ping loop if disabledPingDuration
	}
	pingTicker := time.NewTicker(conn.pingDuration)
	defer pingTicker.Stop()

	for {
		err := conn.ping()
		if err != nil {
			return
		}
		select {
		case <-conn.websocketCtx.Done():
			return
		case <-pingTicker.C:
			continue
		}
	}
}

func (conn *WebsocketConnection) ping() error {
	conn.mtx.Lock()
	defer conn.mtx.Unlock()

	if conn.client == nil {
		return fmt.Errorf("unable to ping closed connection")
	}
	err := conn.client.WriteMessage(int(conn.pingMessageType), ping)
	if err != nil {
		conn.logger.Err(fmt.Errorf(types.ErrWebsocketSend.Error(), conn.providerName, err)).Send()
	}
	return err
}

// readWebSocket continuously reads from the websocket and relays messages
// to the passed in messageHandler. On websocket error this function
// terminates and starts the reconnect process.
// Some providers (Binance) will only allow a valid connection for 24 hours
// so we manually disconnect and reconnect every 23 hours (defaultMaxConnectionTime)
func (conn *WebsocketConnection) readWebSocket() {
	reconnectTicker := time.NewTicker(defaultMaxConnectionTime)
	defer reconnectTicker.Stop()

	for {
		select {
		case <-conn.websocketCtx.Done():
			conn.close()
			return
		case <-time.After(defaultReadNewWSMessage):
			messageType, bz, err := conn.client.ReadMessage()
			if err != nil {
				conn.logger.Err(fmt.Errorf(types.ErrWebsocketRead.Error(), conn.providerName, err)).Send()
				conn.reconnect()
				return
			}
			conn.readSuccess(messageType, bz)
		case <-reconnectTicker.C:
			conn.reconnect()
			return
		}
	}
}

func (conn *WebsocketConnection) readSuccess(messageType int, bz []byte) {
	if len(bz) == 0 {
		return
	}
	// mexc and bitget do not send a valid pong response code so check for it here
	if string(bz) == "pong" {
		return
	}
	conn.messageHandler(messageType, bz)
}

// close sends a close message to the websocket and sets the client to nil
func (conn *WebsocketConnection) close() {
	conn.mtx.Lock()
	defer conn.mtx.Unlock()

	conn.logger.Debug().Msg("closing websocket")
	conn.websocketCancelFunc()
	if err := conn.client.Close(); err != nil {
		conn.logger.Err(fmt.Errorf(types.ErrWebsocketClose.Error(), conn.providerName, err)).Send()
	}
	conn.client = nil
}

// reconnect closes the current websocket and starts a new connection process
func (conn *WebsocketConnection) reconnect() {
	conn.close()
	go conn.start()
	telemetryWebsocketReconnect(conn.providerName)
}

// pingHandler is called by the websocket library whenever a ping message is received
// and responds with a pong message to the server
func (conn *WebsocketConnection) pingHandler(string) error {
	if err := conn.client.WriteMessage(websocket.PongMessage, []byte("pong")); err != nil {
		conn.logger.Error().Err(err).Msg("error sending pong")
	}
	return nil
}
