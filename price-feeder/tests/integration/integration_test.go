package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/oracle"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

func TestBinance(t *testing.T) {
	cps := make([]types.CurrencyPair, 0)
	cps = append(cps, types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
	providerTest(t, provider.ProviderBinance, cps...)
}

func TestMexc(t *testing.T) {
	cps := make([]types.CurrencyPair, 0)
	cps = append(cps, types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
	providerTest(t, provider.ProviderMexc, cps...)
}

func providerTest(t *testing.T, providerName provider.Name, currencyPairs ...types.CurrencyPair) {
	ctx, cancel := context.WithCancel(context.Background())

	cps := make([]types.CurrencyPair, 0)
	cps = append(cps, types.CurrencyPair{Base: "ATOM", Quote: "USDT"})

	pvd, _ := oracle.NewProvider(ctx, providerName, loggerFixture(), provider.Endpoint{}, cps...)

	time.Sleep(15 * time.Second)

	candlePrices, _ := pvd.GetCandlePrices(cps...)
	fmt.Printf("candle prices: %+v\n", candlePrices)
	require.True(t, len(candlePrices["ATOMUSDT"]) > 0)

	tickerPrices, _ := pvd.GetTickerPrices(cps...)
	fmt.Printf("ticker prices: %+v\n", tickerPrices)

	if _, ok := tickerPrices["ATOMUSDT"]; !ok {
		t.Error("ticker prices did not contain required currency pair")
	}
	cancel()
	time.Sleep(5 * time.Second)
}

func loggerFixture() zerolog.Logger {
	logWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	logLvl := zerolog.DebugLevel
	return zerolog.New(logWriter).Level(logLvl).With().Timestamp().Logger()
}
