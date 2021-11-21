package oracle

import (
	"context"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
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

var denoms = []string{"ATOMUSDT"}

// Should return a prices map that we can set
// The oracle to. Needs to touch multiple providers,
// And then average out.
func GetPrices() map[string]sdk.Dec {

	binanceProvider := provider.NewBinanceProvider()
	krakenProvider := provider.NewKrakenProvider()

	wg := new(sync.WaitGroup)

	wg.Add(1)

	binancePrices, err := binanceProvider.GetTickerPrices(denoms...)

	// TODO : Fail silently when providers fail

	if binancePrices == nil || err != nil {
		panic("Unable to get binance prices")
	}

	wg.Done()

	wg.Add(1)

	krackenPrices, bigErr := krakenProvider.GetTickerPrices(denoms...)

	if krackenPrices == nil || bigErr != nil {
		panic("Unable to get kracken prices")
	}

	wg.Done()

	wg.Wait()

	// Create new map for averages

	var averages = make(map[string]sdk.Dec, len(denoms))

	half := sdk.MustNewDecFromStr("0.50")

	for _, v := range denoms {
		averages[v] = binancePrices[v]
		averages[v].Add(krackenPrices[v])
		averages[v].Mul(half)
	}

	return averages

}

func (o *Oracle) tick() {

	pricesArray := GetPrices()

	if pricesArray == nil {
		panic("Unable to get prices")
	}

	o.prices = pricesArray

}

func (o *Oracle) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			o.closer.Close()
			return

		default:
			// TODO : Finish main loop
			// ref : https://github.com/umee-network/umee/issues/178
			o.tick()
			o.lastPriceSyncTS = time.Now()
			time.Sleep(10 * time.Millisecond)

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

	// Creates a new array for the prices in the oracle
	prices := make(map[string]sdk.Dec, len(o.prices))
	for k, v := range o.prices {
		// Fills in the prices with each value in the oracle
		prices[k] = v
	}

	return prices
}
