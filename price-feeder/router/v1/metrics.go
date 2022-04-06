package v1

import "github.com/umee-network/umee/v2/price-feeder/telemetry"

type Metrics interface {
	Gather(format string) (telemetry.GatherResponse, error)
}
