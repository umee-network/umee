package integration

import (
	"context"
	"os"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/umee-network/umee/price-feeder/oracle"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	logger zerolog.Logger
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.logger = getLogger()
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func getLogger() zerolog.Logger {
	logWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	logLvl := zerolog.DebugLevel
	zerolog.SetGlobalLevel(logLvl)
	return zerolog.New(logWriter).Level(logLvl).With().Timestamp().Logger()
}

func (s *IntegrationTestSuite) TestWebsocketProviders() {

	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testCases := []struct {
		provider      provider.Name
		currencyPairs []types.CurrencyPair
	}{
		{
			provider:      provider.ProviderBinance,
			currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
		},
		{
			provider:      provider.ProviderMexc,
			currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
		},
		{
			provider:      provider.ProviderKraken,
			currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
		},
		{
			provider: provider.ProviderOsmosisV2,
			currencyPairs: []types.CurrencyPair{
				{Base: "OSMO", Quote: "ATOM"},
				{Base: "ATOM", Quote: "JUNO"},
				{Base: "ATOM", Quote: "STARGAZE"},
				{Base: "OSMO", Quote: "WBTC"},
				{Base: "OSMO", Quote: "WETH"},
				{Base: "OSMO", Quote: "CRO"},
			},
		},
		{
			provider:      provider.ProviderCoinbase,
			currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
		},
		{
			provider:      provider.ProviderBitget,
			currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
		},
		{
			provider:      provider.ProviderGate,
			currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
		},
		{
			provider:      provider.ProviderOkx,
			currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
		},
		{
			provider:      provider.ProviderHuobi,
			currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
		},
		{
			provider:      provider.ProviderCrypto,
			currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USD"}},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		s.T().Run(string(tc.provider), func(t *testing.T) {
			t.Parallel()
			s.runPriceTest(t, tc.provider, tc.currencyPairs...)
		})
	}
}

func (s *IntegrationTestSuite) runPriceTest(t *testing.T, providerName provider.Name, currencyPairs ...types.CurrencyPair) {
	ctx, cancel := context.WithCancel(context.Background())
	pvd, _ := oracle.NewProvider(ctx, providerName, s.logger, provider.Endpoint{}, currencyPairs...)

	time.Sleep(30 * time.Second) // wait for provider to connect and receive some prices

	// verify ticker price for currency pair is above zero
	tickerPrices, err := pvd.GetTickerPrices(currencyPairs...)
	require.NoError(t, err)

	// verify candle price for currency pair is above zero
	candlePrices, err := pvd.GetCandlePrices(currencyPairs...)
	require.NoError(t, err)

	for _, cp := range currencyPairs {
		currencyPairKey := cp.String()
		if _, ok := tickerPrices[currencyPairKey]; !ok {
			t.Fatalf("ticker prices did not contain required currency pair: %s", currencyPairKey)
		}
		require.True(t, tickerPrices[currencyPairKey].Price.GT(sdk.NewDec(0)))

		if _, ok := candlePrices[currencyPairKey]; !ok {
			t.Fatalf("candle prices did not contain required currency pair: %s", currencyPairKey)
		}
		require.True(t, candlePrices[currencyPairKey][0].Price.GT(sdk.NewDec(0)))
	}

	cancel()
}
