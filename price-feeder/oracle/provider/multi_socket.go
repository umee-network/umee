package provider

import (
	"context"
	"net/url"
	"time"

	"github.com/rs/zerolog"
)

type MultiSocket struct {
	parentCtx       context.Context
	providerName    Name
	websocketURL    url.URL
	messageHandler  MessageHandler
	pingDuration    time.Duration
	pingMessageType uint
	logger          zerolog.Logger

	controllers []*WebsocketController
}

func NewMultiSocket(
	ctx context.Context,
	providerName Name,
	websocketURL url.URL,
	subscriptionMsgs []interface{},
	messageHandler MessageHandler,
	pingDuration time.Duration,
	pingMessageType uint,
	logger zerolog.Logger,
) *MultiSocket {
	ms := &MultiSocket{
		parentCtx:       ctx,
		providerName:    providerName,
		websocketURL:    websocketURL,
		messageHandler:  messageHandler,
		pingDuration:    pingDuration,
		pingMessageType: pingMessageType,
		logger:          logger,
	}
	ms.AddControllers(subscriptionMsgs)
	return ms
}

func (ms *MultiSocket) AddControllers(msgs []interface{}) {
	for _, sub := range msgs {
		ms.controllers = append(ms.controllers, NewWebsocketController(
			ms.parentCtx,
			ms.providerName,
			ms.websocketURL,
			[]interface{}{sub},
			ms.messageHandler,
			ms.pingDuration,
			ms.pingMessageType,
			ms.logger,
		))
	}
}

func (ms *MultiSocket) AddSubscriptionMsgs(msgs []interface{}) error {
	ms.AddControllers(msgs)
	return nil
}
