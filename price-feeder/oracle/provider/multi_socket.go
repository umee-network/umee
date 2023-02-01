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

	controllers []*SingleSocket
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
	ms.AddAndStartControllers(subscriptionMsgs)
	return ms
}

func (ms *MultiSocket) AddAndStartControllers(msgs []interface{}) {
	for _, sub := range msgs {
		controller := NewWebsocketController(
			ms.parentCtx,
			ms.providerName,
			ms.websocketURL,
			[]interface{}{sub},
			ms.messageHandler,
			ms.pingDuration,
			ms.pingMessageType,
			ms.logger,
		)
		go controller.Start()
		ms.controllers = append(ms.controllers, controller)
	}
}

func (ms *MultiSocket) AddSubscriptionMsgs(msgs []interface{}) error {
	ms.AddAndStartControllers(msgs)
	return nil
}

func (ms *MultiSocket) Start() {
	// not needed for multi-socket because they are auto started
}

func (ms *MultiSocket) SendJSON(interface{}) error {
	// currently incompatible with multi-socket connections
	return nil
}
