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

	WebsocketController struct {
		ctx            context.Context
		providerName   Name
		websocketUrl   url.URL
		tickerSubMsg   string
		candleSubMsg   string
		messageHandler MessageHandler
		logger         zerolog.Logger
		client         *websocket.Conn
		mu             sync.Mutex
	}
)

func NewWebsocketHandler(
	ctx context.Context,
	providerName Name,
	websocketUrl url.URL,
	tickerSubMsg string,
	candleSubMsg string,
	messageHandler MessageHandler,
	logger zerolog.Logger) *WebsocketController {
	return &WebsocketController{
		ctx:            ctx,
		providerName:   providerName,
		websocketUrl:   websocketUrl,
		tickerSubMsg:   tickerSubMsg,
		candleSubMsg:   candleSubMsg,
		messageHandler: messageHandler,
		logger:         logger,
	}
}

func (c *WebsocketController) start() {
	connectTicker := time.NewTicker(defaultReconnectTime)
	defer connectTicker.Stop()

	for {
		err := c.connect() // attempt first connection immediately
		if err == nil {
			c.subscribe()
			go c.readWebSocket()
			return
		}
		select {
		case <-c.ctx.Done():
			return
		case <-connectTicker.C:
			continue
		}
	}
}

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

	if err := c.client.WriteJSON(c.tickerSubMsg); err != nil {
		return fmt.Errorf("error sending ticker subscription message: %w", err)
	}

	if err := c.client.WriteJSON(c.candleSubMsg); err != nil {
		return fmt.Errorf("error sending candle subscription message: %w", err)
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

func (c *WebsocketController) close() {
	err := c.client.Close()
	if err != nil {
		c.logger.Err(err).Msg(fmt.Sprintf("error closing websocket"))
	}
	c.mu.Lock()
	c.client = nil
	c.mu.Unlock()
}

func (c *WebsocketController) reconnect() {
	c.close()
	go c.start()
	telemetryWebsocketReconnect(c.providerName)
}
