package provider

import (
	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

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
				Value: string(n),
			},
		},
	)
}
