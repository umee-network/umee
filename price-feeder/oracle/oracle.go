package oracle

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	pfsync "github.com/umee-network/umee/price-feeder/pkg/sync"
)

type Oracle struct {
	logger zerolog.Logger
	closer *pfsync.Closer
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
