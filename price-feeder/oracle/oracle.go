package oracle

import (
	"context"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	pfsync "github.com/umee-network/umee/price-feeder/pkg/sync"
)

type Oracle struct {
	logger zerolog.Logger
	closer *pfsync.Closer

	mtx             sync.RWMutex
	lastPriceSyncTS time.Time
	prices          map[string]sdk.Dec
}

func New() *Oracle {
	return &Oracle{
		logger: log.With().Str("module", "oracle").Logger(),
		closer: pfsync.NewCloser(),
	}
}

func (o *Oracle) Stop() {
	o.closer.Close()
	<-o.closer.Done()
}

func (o *Oracle) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			o.closer.Close()
			return

		default:
			// TODO: ...
		}
	}
}

// GetLastPriceSyncTimestamp returns the latest timestamp at which prices where
// fetched from the oracle's set of exchange rate providers.
func (o *Oracle) GetLastPriceSyncTimestamp() time.Time {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.lastPriceSyncTS
}

// GetPrices returns a copy of the current prices fetched from the oracle's
// set of exchange rate providers.
func (o *Oracle) GetPrices() map[string]sdk.Dec {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	prices := make(map[string]sdk.Dec, len(o.prices))
	for k, v := range o.prices {
		prices[k] = v
	}

	return prices
}
