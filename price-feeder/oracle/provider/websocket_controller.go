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
	// that manages reconnecting, subscribing, and recieiving messages
	WebsocketController struct {
		ctx              context.Context
		providerName     Name
		websocketUrl     url.URL
		subscriptionMsgs []string
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
	subscriptionMsgs []string,
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

func (c *WebsocketController) start() {
	connectTicker := time.NewTicker(defaultReconnectTime)
	defer connectTicker.Stop()

	for {
		err := c.connect() // attempt first connection immediately
		if err == nil {
			err = c.subscribe()
			if err == nil {
				go c.readWebSocket()
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
	conn, resp, err := websocket.DefaultDialer.Dial(c.websocketUrl.String(), nil)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error connecting to websocket: %w", err)
	}
	c.mu.Lock()
	c.client = conn
	c.mu.Unlock()
	return nil
}

func (c *WebsocketController) subscribe() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for message := range c.subscriptionMsgs {
		if err := c.client.WriteJSON(message); err != nil {
			return fmt.Errorf("error sending ticker subscription message: %w", err)
		}
	}
	return nil
}

func (c *WebsocketController) readWebSocket() {
	reconnectTicker := time.NewTicker(defaultMaxConnectionTime)
	defer reconnectTicker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(defaultReadNewWSMessage):
			messageType, bz, err := c.client.ReadMessage()
			if err != nil {
				c.logger.Err(err).Msg("error reading websocket message")
				if websocket.IsCloseError(err) {
					c.reconnect()
					return // stop reading from websocket after close error
				}
				continue // for other errors continue to try and read the next message
			}
			if len(bz) > 0 {
				c.messageHandler(messageType, bz)
			}
		case <-reconnectTicker.C:
			c.reconnect()
		}
	}
}

// close sends a close message to the websocket and sets the client to nil
func (c *WebsocketController) close() {
	err := c.client.Close()
	if err != nil {
		c.logger.Err(err).Msg(fmt.Sprintf("error closing websocket"))
	}
	c.mu.Lock()
	c.client = nil
	c.mu.Unlock()
}

// reconnect closes the current websocket and starts a new connection process
func (c *WebsocketController) reconnect() {
	c.close()
	go c.start()
	telemetryWebsocketReconnect(c.providerName)
}
