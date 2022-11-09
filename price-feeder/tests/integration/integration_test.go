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
			runPriceTest(t, tc.provider, tc.currencyPairs...)
		})
	}
}

func runPriceTest(t *testing.T, providerName provider.Name, currencyPairs ...types.CurrencyPair) {
	ctx, cancel := context.WithCancel(context.Background())
	pvd, _ := oracle.NewProvider(ctx, providerName, getLogger(), provider.Endpoint{}, currencyPairs...)

	time.Sleep(30 * time.Second) // wait for provider to connect and receive some prices

	checkForPrices(t, pvd, currencyPairs)

	cancel()
}

func (s *IntegrationTestSuite) TestSubscribeCurrencyPairs() {
	ctx, cancel := context.WithCancel(context.Background())
	currencyPairs := []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}}
	pvd, _ := provider.NewBinanceProvider(ctx, getLogger(), provider.Endpoint{}, currencyPairs...)

	time.Sleep(5 * time.Second)

	newPairs := []types.CurrencyPair{{Base: "ETH", Quote: "USDT"}}
	pvd.SubscribeCurrencyPairs(newPairs...)

	currencyPairs = append(currencyPairs, newPairs...)

	time.Sleep(25 * time.Second) // wait for provider to connect and receive some prices

	checkForPrices(s.T(), pvd, currencyPairs)

	cancel()
}

func checkForPrices(t *testing.T, pvd provider.Provider, currencyPairs []types.CurrencyPair) {
	tickerPrices, err := pvd.GetTickerPrices(currencyPairs...)
	require.NoError(t, err)

	candlePrices, err := pvd.GetCandlePrices(currencyPairs...)
	require.NoError(t, err)

	for _, cp := range currencyPairs {
		currencyPairKey := cp.String()

		// verify ticker price for currency pair is above zero
		require.True(t, tickerPrices[currencyPairKey].Price.GT(sdk.NewDec(0)))

		// verify candle price for currency pair is above zero
		require.True(t, candlePrices[currencyPairKey][0].Price.GT(sdk.NewDec(0)))
	}
}
