package provider

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	defaultReadNewWSMessage  = 50 * time.Millisecond
	defaultMaxConnectionTime = time.Hour * 23 // should be < 24h
	defaultReconnectTime     = 2 * time.Minute
	defaultPingDuration      = 15 * time.Second
)

type (
	MessageHandler func(int, []byte)

	// WebsocketController defines a provider agnostic websocket handler
	// that manages reconnecting, subscribing, and receiving messages
	WebsocketController struct {
		parentCtx           context.Context
		websocketCtx        context.Context
		websocketCancelFunc context.CancelFunc
		providerName        Name
		websocketURL        url.URL
		subscriptionMsgs    []interface{}
		messageHandler      MessageHandler
		pingDuration        time.Duration
		pingMessageType     uint
		logger              zerolog.Logger

		mtx    sync.Mutex
		client *websocket.Conn
	}
)

// NewWebsocketController does nothing except initialize a new WebsocketController
// and provider a reminder for what fields need to be passed in.
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
	return &WebsocketController{
		parentCtx:        ctx,
		providerName:     providerName,
		websocketURL:     websocketURL,
		subscriptionMsgs: subscriptionMsgs,
		messageHandler:   messageHandler,
		pingDuration:     pingDuration,
		pingMessageType:  pingMessageType,
		logger:           logger,
	}
}

// Start will continuously loop and attempt connecting to the websocket
// until a successful connection is made. It then starts the ping
// service and read listener in new go routines and sends subscription
// messages  using the passed in subscription messages
func (wsc *WebsocketController) Start() {
	connectTicker := time.NewTicker(defaultReconnectTime)
	defer connectTicker.Stop()

	for {
		if err := wsc.connect(); err != nil {
			wsc.logger.Err(err).Send()
			select {
			case <-wsc.parentCtx.Done():
				return
			case <-connectTicker.C:
				continue
			}
		}

		go wsc.readWebSocket()
		go wsc.pingLoop()

		if err := wsc.subscribe(); err != nil {
			wsc.logger.Err(err).Send()
			wsc.close()
			continue
		}
		return
	}
}

func (wsc *WebsocketController) pongHandler(appData string) error {
	wsc.logger.Debug().Str("msg", appData).Msg("pong received")
	return nil
}

func (wsc *WebsocketController) pingHandler(appData string) error {
	wsc.logger.Debug().Str("msg", appData).Msg("ping received")
	if err := wsc.client.WriteMessage(websocket.PongMessage, []byte("pong")); err != nil {
		wsc.logger.Error().Err(err).Msg("error sending pong")
	}
	return nil
}

// connect dials the websocket and sets the client to the established connection
func (wsc *WebsocketController) connect() error {
	wsc.mtx.Lock()
	defer wsc.mtx.Unlock()

	wsc.logger.Debug().Msg("connecting to websocket")
	conn, resp, err := websocket.DefaultDialer.Dial(wsc.websocketURL.String(), nil)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf(types.ErrWebsocketDial.Error(), wsc.providerName, err)
	}
	wsc.client = conn
	wsc.websocketCtx, wsc.websocketCancelFunc = context.WithCancel(wsc.parentCtx)
	wsc.client.SetPingHandler((wsc.pingHandler))
	wsc.client.SetPongHandler(wsc.pongHandler)
	return nil
}

// subscribe sends the WebsocketControllers subscription messages to the websocket
func (wsc *WebsocketController) subscribe() error {
	for _, jsonMessage := range wsc.subscriptionMsgs {
		wsc.logger.Debug().Interface("msg", jsonMessage).Msg("sending websocket message")
		if err := wsc.SendJSON(jsonMessage); err != nil {
			return fmt.Errorf(types.ErrWebsocketSend.Error(), wsc.providerName, err)
		}
	}
	return nil
}

// SendJSON sends a json message to the websocket connection using the Websocket
// Controller mutex to ensure multiple writes do not happen at once
func (wsc *WebsocketController) SendJSON(msg interface{}) error {
	wsc.mtx.Lock()
	defer wsc.mtx.Unlock()

	if wsc.client == nil {
		return fmt.Errorf("unable to send JSON on a closed connection")
	}
	wsc.logger.Debug().Interface("msg", msg).Msg("sending websocket message")
	if err := wsc.client.WriteJSON(msg); err != nil {
		return fmt.Errorf(types.ErrWebsocketSend.Error(), wsc.providerName, err)
	}
	return nil
}

// ping sends a ping to the server every defaultPingDuration
func (wsc *WebsocketController) pingLoop() {
	if wsc.pingDuration == time.Duration(0) {
		return // disable ping loop if duration is zero
	}
	pingTicker := time.NewTicker(wsc.pingDuration)
	defer pingTicker.Stop()

	for {
		err := wsc.ping()
		if err != nil {
			return
		}
		wsc.logger.Debug().Msg("ping sent")
		select {
		case <-wsc.websocketCtx.Done():
			return
		case <-pingTicker.C:
			continue
		}
	}
}

func (wsc *WebsocketController) ping() error {
	wsc.mtx.Lock()
	defer wsc.mtx.Unlock()

	if wsc.client == nil {
		return fmt.Errorf("unable to ping closed connection")
	}
	err := wsc.client.WriteMessage(int(wsc.pingMessageType), []byte("ping"))
	if err != nil {
		wsc.logger.Err(fmt.Errorf(types.ErrWebsocketSend.Error(), wsc.providerName, err)).Send()
	}
	return err
}

// readWebSocket continuously reads from the websocket and relays messages
// to the passed in messageHandler. On websocket error this function
// terminates and starts the reconnect process.
// Some providers (Binance) will only allow a valid connection for 24 hours
// so we manually disconnect and reconnect every 23 hours (defaultMaxConnectionTime)
func (wsc *WebsocketController) readWebSocket() {
	reconnectTicker := time.NewTicker(defaultMaxConnectionTime)
	defer reconnectTicker.Stop()

	for {
		select {
		case <-wsc.websocketCtx.Done():
			wsc.close()
			return
		case <-time.After(defaultReadNewWSMessage):
			messageType, bz, err := wsc.client.ReadMessage()
			if err != nil {
				wsc.logger.Err(fmt.Errorf(types.ErrWebsocketRead.Error(), wsc.providerName, err)).Send()
				wsc.reconnect()
				return
			}
			wsc.readSuccess(messageType, bz)
		case <-reconnectTicker.C:
			wsc.reconnect()
			return
		}
	}
}

func (wsc *WebsocketController) readSuccess(messageType int, bz []byte) {
	if len(bz) == 0 {
		return
	}
	// mexc and bitget do not send a valid pong response code so check for it here
	if string(bz) == "pong" {
		wsc.pongHandler(string(bz))
		return
	}
	wsc.messageHandler(messageType, bz)
}

// close sends a close message to the websocket and sets the client to nil
func (wsc *WebsocketController) close() {
	wsc.mtx.Lock()
	defer wsc.mtx.Unlock()

	wsc.logger.Debug().Msg("closing websocket")
	wsc.websocketCancelFunc()
	if err := wsc.client.Close(); err != nil {
		wsc.logger.Err(fmt.Errorf(types.ErrWebsocketClose.Error(), wsc.providerName, err)).Send()
	}
	wsc.client = nil
}

// reconnect closes the current websocket and starts a new connection process
func (wsc *WebsocketController) reconnect() {
	wsc.close()
	go wsc.Start()
	telemetryWebsocketReconnect(wsc.providerName)
}
