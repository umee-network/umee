package provider

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type (
	MessageHandler func(int, []byte)

	// WebsocketController defines a provider agnostic websocket handler
	// that manages reconnecting, subscribing, and receiving messages
	WebsocketController struct {
		ctx              context.Context
		providerName     Name
		websocketUrl     url.URL
		subscriptionMsgs []interface{}
		messageHandler   MessageHandler
		logger           zerolog.Logger
		client           *websocket.Conn
		mu               sync.Mutex
	}
)

func NewWebsocketController(
	ctx context.Context,
	providerName Name,
	websocketUrl url.URL,
	subscriptionMsgs []interface{},
	messageHandler MessageHandler,
	logger zerolog.Logger,
) *WebsocketController {
	return &WebsocketController{
		ctx:              ctx,
		providerName:     providerName,
		websocketUrl:     websocketUrl,
		subscriptionMsgs: subscriptionMsgs,
		messageHandler:   messageHandler,
		logger:           logger,
	}
}

// Start will contniously loop and attempt connecting to the websocket
// until a successsful connection is made. It then starts the ping
// service and read listener in new go routines and sends subscription
// messages  using the passed in subscription messages
func (c *WebsocketController) Start() {
	connectTicker := time.NewTicker(defaultReconnectTime)
	defer connectTicker.Stop()

	for {
		err := c.connect()
		if err == nil {
			go c.readWebSocket()
			go c.ping()
			err = c.subscribe()
			if err == nil {
				return
			}
			c.close()
		}
		select {
		case <-c.ctx.Done():
			return
		case <-connectTicker.C:
			continue
		}
	}
}

// connect dials the websocket and sets the client to the established connection
func (c *WebsocketController) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Info().Msg("connecting to websocket")
	conn, resp, err := websocket.DefaultDialer.Dial(c.websocketUrl.String(), nil)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error connecting to websocket: %w", err)
	}
	c.client = conn
	return nil
}

// subscribe sends the WebsocketControllers subscription messages to the websocket
func (c *WebsocketController) subscribe() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, jsonMessage := range c.subscriptionMsgs {
		c.logger.Info().Msg(fmt.Sprintf("sending message: %+v", jsonMessage))
		if err := c.client.WriteJSON(jsonMessage); err != nil {
			return fmt.Errorf("error sending subscription message: %w", err)
		}
	}
	return nil
}

// ping sends a ping to the server every defaultPingDuration
func (c *WebsocketController) ping() {
	pingTicker := time.NewTicker(defaultPingDuration)
	defer pingTicker.Stop()

	for {
		if c.client == nil {
			return
		}
		c.logger.Info().Msg("ping")
		c.mu.Lock()
		err := c.client.WriteMessage(1, []byte("ping"))
		c.mu.Unlock()
		if err != nil {
			c.logger.Err(err).Msg("error sending ping message")
		}
		select {
		case <-c.ctx.Done():
			return
		case <-pingTicker.C:
			continue
		}
	}
}

// readWebSocket contiously reads from the websocket and relays messages
// to the passed in messageHandler. On websocket error this function
// terminates and starts the reconnect process
func (c *WebsocketController) readWebSocket() {
	reconnectTicker := time.NewTicker(defaultMaxConnectionTime)
	defer reconnectTicker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			c.close()
			return
		case <-time.After(defaultReadNewWSMessage):
			messageType, bz, err := c.client.ReadMessage()
			if err != nil {
				c.logger.Err(err).Msg("error reading websocket message")
				c.reconnect()
				return
			}
			c.readSuccess(messageType, bz)
		case <-reconnectTicker.C:
			c.reconnect()
			return
		}
	}
}

func (c *WebsocketController) readSuccess(messageType int, bz []byte) {
	c.logger.Info().Msg(fmt.Sprintf("%d: %s", messageType, string(bz)))

	if messageType != websocket.TextMessage || len(bz) <= 0 {
		return
	}
	if string(bz) == "pong" {
		c.client.PongHandler()
		return
	}
	c.messageHandler(messageType, bz)
}

// close sends a close message to the websocket and sets the client to nil
func (c *WebsocketController) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Info().Msg("closing websocket")
	err := c.client.Close()
	if err != nil {
		c.logger.Err(err).Msg("error closing websocket")
	}
	c.client = nil
}

// reconnect closes the current websocket and starts a new connection process
func (c *WebsocketController) reconnect() {
	c.close()
	go c.Start()
	telemetryWebsocketReconnect(c.providerName)
}
