package provider

import (
	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

const (
	messageTypeCandle = messageType("candle")
	messageTypeTicker = messageType("ticker")
	messageTypeTrade  = messageType("trade")
)

type (
	messageType string
)

// String cast provider MessageType to string.
func (mt messageType) String() string {
	return string(mt)
}

// telemetryWebsocketReconnect gives an standard way to add
// `price_feeder_websocket_reconnect` metric.
func telemetryWebsocketReconnect(n Name) {
	telemetry.IncrCounterWithLabels(
		[]string{
			"websocket",
			"reconnect",
		},
		1,
		[]metrics.Label{
			{
				Name:  "provider",
				Value: n.String(),
			},
		},
	)
}

// telemetryWebsocketMessage gives an standard way to add
// `price_feeder_websocket_message{type="x", provider="x"}` metric.
func telemetryWebsocketMessage(n Name, mt messageType) {
	telemetry.IncrCounterWithLabels(
		[]string{
			"websocket",
			"message",
		},
		1,
		[]metrics.Label{
			{
				Name:  "provider",
				Value: n.String(),
			},
			{
				Name:  "type",
				Value: mt.String(),
			},
		},
	)
}
