package oracle

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/oracle/client"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
	"github.com/umee-network/umee/price-feeder/oracle/types"

	oracletypes "github.com/umee-network/umee/v2/x/oracle/types"
)

type mockProvider struct {
	prices map[string]provider.TickerPrice
}

func (m mockProvider) GetTickerPrices(_ ...types.CurrencyPair) (map[string]provider.TickerPrice, error) {
	return m.prices, nil
}

func (m mockProvider) GetCandlePrices(_ ...types.CurrencyPair) (map[string][]provider.CandlePrice, error) {
	candles := make(map[string][]provider.CandlePrice)
	for pair, price := range m.prices {
		candles[pair] = []provider.CandlePrice{
			{
				Price:     price.Price,
				TimeStamp: provider.PastUnixTime(1 * time.Minute),
				Volume:    price.Volume,
			},
		}
	}
	return candles, nil
}

func (m mockProvider) SubscribeCurrencyPairs(_ ...types.CurrencyPair) error {
	return nil
}

func (m mockProvider) GetAvailablePairs() (map[string]struct{}, error) {
	return map[string]struct{}{}, nil
}

type failingProvider struct {
	prices map[string]provider.TickerPrice
}

func (m failingProvider) GetTickerPrices(_ ...types.CurrencyPair) (map[string]provider.TickerPrice, error) {
	return nil, fmt.Errorf("unable to get ticker prices")
}

func (m failingProvider) GetCandlePrices(_ ...types.CurrencyPair) (map[string][]provider.CandlePrice, error) {
	return nil, fmt.Errorf("unable to get candle prices")
}

func (m failingProvider) SubscribeCurrencyPairs(_ ...types.CurrencyPair) error {
	return nil
}

func (m failingProvider) GetAvailablePairs() (map[string]struct{}, error) {
	return map[string]struct{}{}, nil
}

type OracleTestSuite struct {
	suite.Suite

	oracle *Oracle
}

// SetupSuite executes once before the suite's tests are executed.
func (ots *OracleTestSuite) SetupSuite() {
	ots.oracle = New(
		zerolog.Nop(),
		client.OracleClient{},
		[]config.CurrencyPair{
			{
				Base:      "UMEE",
				Quote:     "USDT",
				Providers: []string{config.ProviderBinance},
			},
			{
				Base:      "UMEE",
				Quote:     "USDC",
				Providers: []string{config.ProviderKraken},
			},
			{
				Base:      "XBT",
				Quote:     "USDT",
				Providers: []string{config.ProviderOsmosis},
			},
			{
				Base:      "USDC",
				Quote:     "USD",
				Providers: []string{config.ProviderHuobi},
			},
			{
				Base:      "USDT",
				Quote:     "USD",
				Providers: []string{config.ProviderCoinbase},
			},
		},
		time.Millisecond*100,
	)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OracleTestSuite))
}

func (ots *OracleTestSuite) TestStop() {
	ots.Eventually(
		func() bool {
			ots.oracle.Stop()
			return true
		},
		5*time.Second,
		time.Second,
	)
}

func (ots *OracleTestSuite) TestGetLastPriceSyncTimestamp() {
	// when no tick() has been invoked, assume zero value
	ots.Require().Equal(time.Time{}, ots.oracle.GetLastPriceSyncTimestamp())
}

func (ots *OracleTestSuite) TestPrices() {
	// initial prices should be empty (not set)
	ots.Require().Empty(ots.oracle.GetPrices())

	// Use a mock provider with exchange rates that are not specified in
	// configuration.
	ots.oracle.priceProviders = map[string]provider.Provider{
		config.ProviderBinance: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDX": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		config.ProviderKraken: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDX": {
					Price:  sdk.MustNewDecFromStr("3.70"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}
	acceptList := oracletypes.DenomList{
		oracletypes.Denom{
			BaseDenom:   "UUMEE",
			SymbolDenom: "UMEE",
			Exponent:    6,
		},
	}

	ots.Require().Error(ots.oracle.SetPrices(context.TODO(), acceptList))
	ots.Require().Empty(ots.oracle.GetPrices())

	// use a mock provider without a conversion rate for these stablecoins
	ots.oracle.priceProviders = map[string]provider.Provider{
		config.ProviderBinance: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDT": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		config.ProviderKraken: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDC": {
					Price:  sdk.MustNewDecFromStr("3.70"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().Error(ots.oracle.SetPrices(context.TODO(), acceptList))

	prices := ots.oracle.GetPrices()
	ots.Require().Len(prices, 0)

	// use a mock provider to provide prices for the configured exchange pairs
	ots.oracle.priceProviders = map[string]provider.Provider{
		config.ProviderBinance: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDT": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		config.ProviderKraken: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDC": {
					Price:  sdk.MustNewDecFromStr("3.70"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		config.ProviderHuobi: mockProvider{
			prices: map[string]provider.TickerPrice{
				"USDCUSD": {
					Price:  sdk.MustNewDecFromStr("1"),
					Volume: sdk.MustNewDecFromStr("2396974.34000000"),
				},
			},
		},
		config.ProviderCoinbase: mockProvider{
			prices: map[string]provider.TickerPrice{
				"USDTUSD": {
					Price:  sdk.MustNewDecFromStr("1"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		config.ProviderOsmosis: mockProvider{
			prices: map[string]provider.TickerPrice{
				"XBTUSDT": {
					Price:  sdk.MustNewDecFromStr("3.717"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().NoError(ots.oracle.SetPrices(context.TODO(), acceptList))

	prices = ots.oracle.GetPrices()
	ots.Require().Len(prices, 4)
	ots.Require().Equal(sdk.MustNewDecFromStr("3.710916056220858266"), prices["UMEE"])
	ots.Require().Equal(sdk.MustNewDecFromStr("3.717"), prices["XBT"])
	ots.Require().Equal(sdk.MustNewDecFromStr("1"), prices["USDC"])
	ots.Require().Equal(sdk.MustNewDecFromStr("1"), prices["USDT"])

	// use one working provider and one provider with an incorrect exchange rate
	ots.oracle.priceProviders = map[string]provider.Provider{
		config.ProviderBinance: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDX": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		config.ProviderKraken: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDC": {
					Price:  sdk.MustNewDecFromStr("3.70"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		config.ProviderHuobi: mockProvider{
			prices: map[string]provider.TickerPrice{
				"USDCUSD": {
					Price:  sdk.MustNewDecFromStr("1"),
					Volume: sdk.MustNewDecFromStr("2396974.34000000"),
				},
			},
		},
		config.ProviderCoinbase: mockProvider{
			prices: map[string]provider.TickerPrice{
				"USDTUSD": {
					Price:  sdk.MustNewDecFromStr("1"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		config.ProviderOsmosis: mockProvider{
			prices: map[string]provider.TickerPrice{
				"XBTUSDT": {
					Price:  sdk.MustNewDecFromStr("3.717"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().NoError(ots.oracle.SetPrices(context.TODO(), acceptList))
	prices = ots.oracle.GetPrices()
	ots.Require().Len(prices, 4)
	ots.Require().Equal(sdk.MustNewDecFromStr("3.70"), prices["UMEE"])
	ots.Require().Equal(sdk.MustNewDecFromStr("3.717"), prices["XBT"])
	ots.Require().Equal(sdk.MustNewDecFromStr("1"), prices["USDC"])
	ots.Require().Equal(sdk.MustNewDecFromStr("1"), prices["USDT"])

	// use one working provider and one provider that fails
	ots.oracle.priceProviders = map[string]provider.Provider{
		config.ProviderBinance: failingProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDC": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		config.ProviderKraken: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDC": {
					Price:  sdk.MustNewDecFromStr("3.71"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		config.ProviderHuobi: mockProvider{
			prices: map[string]provider.TickerPrice{
				"USDCUSD": {
					Price:  sdk.MustNewDecFromStr("1"),
					Volume: sdk.MustNewDecFromStr("2396974.34000000"),
				},
			},
		},
		config.ProviderCoinbase: mockProvider{
			prices: map[string]provider.TickerPrice{
				"USDTUSD": {
					Price:  sdk.MustNewDecFromStr("1"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		config.ProviderOsmosis: mockProvider{
			prices: map[string]provider.TickerPrice{
				"XBTUSDT": {
					Price:  sdk.MustNewDecFromStr("3.717"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().NoError(ots.oracle.SetPrices(context.TODO(), acceptList))
	prices = ots.oracle.GetPrices()
	ots.Require().Len(prices, 4)
	ots.Require().Equal(sdk.MustNewDecFromStr("3.71"), prices["UMEE"])
	ots.Require().Equal(sdk.MustNewDecFromStr("3.717"), prices["XBT"])
	ots.Require().Equal(sdk.MustNewDecFromStr("1"), prices["USDC"])
	ots.Require().Equal(sdk.MustNewDecFromStr("1"), prices["USDT"])
}

func TestGenerateSalt(t *testing.T) {
	salt, err := GenerateSalt(0)
	require.Error(t, err)
	require.Empty(t, salt)

	salt, err = GenerateSalt(32)
	require.NoError(t, err)
	require.NotEmpty(t, salt)
}

func TestGenerateExchangeRatesString(t *testing.T) {
	testCases := map[string]struct {
		input    map[string]sdk.Dec
		expected string
	}{
		"empty input": {
			input:    make(map[string]sdk.Dec),
			expected: "",
		},
		"single denom": {
			input: map[string]sdk.Dec{
				"UMEE": sdk.MustNewDecFromStr("3.72"),
			},
			expected: "UMEE:3.720000000000000000",
		},
		"multi denom": {
			input: map[string]sdk.Dec{
				"UMEE": sdk.MustNewDecFromStr("3.72"),
				"ATOM": sdk.MustNewDecFromStr("40.13"),
				"OSMO": sdk.MustNewDecFromStr("8.69"),
			},
			expected: "ATOM:40.130000000000000000,OSMO:8.690000000000000000,UMEE:3.720000000000000000",
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			out := GenerateExchangeRatesString(tc.input)
			require.Equal(t, tc.expected, out)
		})
	}
}

func TestSuccessSetProviderTickerPricesAndCandles(t *testing.T) {
	providerPrices := make(provider.AggregatedProviderPrices, 1)
	providerCandles := make(provider.AggregatedProviderCandles, 1)
	pair := types.CurrencyPair{
		Base:  "ATOM",
		Quote: "USDT",
	}

	atomPrice := sdk.MustNewDecFromStr("29.93")
	atomVolume := sdk.MustNewDecFromStr("894123.00")

	prices := make(map[string]provider.TickerPrice, 1)
	prices[pair.String()] = provider.TickerPrice{
		Price:  atomPrice,
		Volume: atomVolume,
	}

	candles := make(map[string][]provider.CandlePrice, 1)
	candles[pair.String()] = []provider.CandlePrice{
		{
			Price:     atomPrice,
			Volume:    atomVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}

	success := SetProviderTickerPricesAndCandles(
		config.ProviderGate,
		providerPrices,
		providerCandles,
		prices,
		candles,
		pair,
	)

	require.True(t, success, "It should successfully set the prices")
	require.Equal(t, providerPrices[config.ProviderGate][pair.Base].Price, atomPrice)
	require.Equal(t, providerCandles[config.ProviderGate][pair.Base][0].Price, atomPrice)
}

func TestFailedSetProviderTickerPricesAndCandles(t *testing.T) {
	success := SetProviderTickerPricesAndCandles(
		config.ProviderCoinbase,
		make(provider.AggregatedProviderPrices, 1),
		make(provider.AggregatedProviderCandles, 1),
		make(map[string]provider.TickerPrice, 1),
		make(map[string][]provider.CandlePrice, 1),
		types.CurrencyPair{
			Base:  "ATOM",
			Quote: "USDT",
		},
	)

	require.False(t, success, "It should failed to set the prices, prices and candle are empty")
}

func TestSuccessGetComputedPricesCandles(t *testing.T) {
	providerCandles := make(provider.AggregatedProviderCandles, 1)
	pair := types.CurrencyPair{
		Base:  "ATOM",
		Quote: "USD",
	}

	atomPrice := sdk.MustNewDecFromStr("29.93")
	atomVolume := sdk.MustNewDecFromStr("894123.00")

	candles := make(map[string][]provider.CandlePrice, 1)
	candles[pair.String()] = []provider.CandlePrice{
		{
			Price:     atomPrice,
			Volume:    atomVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}
	providerCandles[config.ProviderBinance] = candles

	providerPair := map[string][]types.CurrencyPair{
		"binance": {pair},
	}

	prices, err := GetComputedPrices(
		zerolog.Nop(),
		providerCandles,
		make(provider.AggregatedProviderPrices, 1),
		providerPair,
	)

	require.NoError(t, err, "It should successfully get computed candle prices")
	require.Equal(t, prices[pair.String()], atomPrice)
}

func TestSuccessGetComputedPricesTickers(t *testing.T) {
	providerPrices := make(provider.AggregatedProviderPrices, 1)
	pair := types.CurrencyPair{
		Base:  "ATOM",
		Quote: "USD",
	}

	atomPrice := sdk.MustNewDecFromStr("29.93")
	atomVolume := sdk.MustNewDecFromStr("894123.00")

	tickerPrices := make(map[string]provider.TickerPrice, 1)
	tickerPrices[pair.String()] = provider.TickerPrice{
		Price:  atomPrice,
		Volume: atomVolume,
	}
	providerPrices[config.ProviderBinance] = tickerPrices

	providerPair := map[string][]types.CurrencyPair{
		"binance": {pair},
	}

	prices, err := GetComputedPrices(
		zerolog.Nop(),
		make(provider.AggregatedProviderCandles, 1),
		providerPrices,
		providerPair,
	)

	require.NoError(t, err, "It should successfully get computed ticker prices")
	require.Equal(t, prices[pair.String()], atomPrice)
}

func TestSuccessFilterCandleDeviations(t *testing.T) {
	providerCandles := make(provider.AggregatedProviderCandles, 4)
	pair := types.CurrencyPair{
		Base:  "ATOM",
		Quote: "USDT",
	}

	atomPrice := sdk.MustNewDecFromStr("29.93")
	atomVolume := sdk.MustNewDecFromStr("1994674.34000000")

	atomCandlePrice := []provider.CandlePrice{
		{
			Price:     atomPrice,
			Volume:    atomVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}

	providerCandles[config.ProviderBinance] = map[string][]provider.CandlePrice{
		pair.String(): atomCandlePrice,
	}
	providerCandles[config.ProviderHuobi] = map[string][]provider.CandlePrice{
		pair.String(): atomCandlePrice,
	}
	providerCandles[config.ProviderKraken] = map[string][]provider.CandlePrice{
		pair.String(): atomCandlePrice,
	}
	providerCandles[config.ProviderCoinbase] = map[string][]provider.CandlePrice{
		pair.String(): {
			{
				Price:     sdk.MustNewDecFromStr("27.1"),
				Volume:    atomVolume,
				TimeStamp: provider.PastUnixTime(1 * time.Minute),
			},
		},
	}

	pricesFiltered, err := FilterCandleDeviations(
		zerolog.Nop(),
		providerCandles,
	)

	_, ok := pricesFiltered[config.ProviderCoinbase]
	require.NoError(t, err, "It should successfully filter out the provider using candles")
	require.False(t, ok, "The filtered candle deviation price at coinbase should be empty")
}

func TestSuccessFilterTickerDeviations(t *testing.T) {
	providerTickers := make(provider.AggregatedProviderPrices, 4)
	pair := types.CurrencyPair{
		Base:  "ATOM",
		Quote: "USDT",
	}

	atomPrice := sdk.MustNewDecFromStr("29.93")
	atomVolume := sdk.MustNewDecFromStr("1994674.34000000")

	atomTickerPrice := provider.TickerPrice{
		Price:  atomPrice,
		Volume: atomVolume,
	}

	providerTickers[config.ProviderBinance] = map[string]provider.TickerPrice{
		pair.String(): atomTickerPrice,
	}
	providerTickers[config.ProviderHuobi] = map[string]provider.TickerPrice{
		pair.String(): atomTickerPrice,
	}
	providerTickers[config.ProviderKraken] = map[string]provider.TickerPrice{
		pair.String(): atomTickerPrice,
	}
	providerTickers[config.ProviderCoinbase] = map[string]provider.TickerPrice{
		pair.String(): {
			Price:  sdk.MustNewDecFromStr("27.1"),
			Volume: atomVolume,
		},
	}

	pricesFiltered, err := FilterTickerDeviations(
		zerolog.Nop(),
		providerTickers,
	)

	_, ok := pricesFiltered[config.ProviderCoinbase]
	require.NoError(t, err, "It should successfully filter out the provider using tickers")
	require.False(t, ok, "The filtered ticker deviation price at coinbase should be empty")
}

func TestGetUSDBasedProviders(t *testing.T) {
	providerPairs := make(map[string][]types.CurrencyPair, 3)
	providerPairs["coinbase"] = []types.CurrencyPair{
		{
			Base:  "FOO",
			Quote: "USD",
		},
	}
	providerPairs["huobi"] = []types.CurrencyPair{
		{
			Base:  "FOO",
			Quote: "USD",
		},
	}
	providerPairs["kraken"] = []types.CurrencyPair{
		{
			Base:  "FOO",
			Quote: "USDT",
		},
	}
	providerPairs["binance"] = []types.CurrencyPair{
		{
			Base:  "USDT",
			Quote: "USD",
		},
	}

	pairs, err := getUSDBasedProviders("FOO", providerPairs)
	require.NoError(t, err)
	expectedPairs := map[string]struct{}{
		"coinbase": {},
		"huobi":    {},
	}
	require.Equal(t, pairs, expectedPairs)

	pairs, err = getUSDBasedProviders("USDT", providerPairs)
	require.NoError(t, err)
	expectedPairs = map[string]struct{}{
		"binance": {},
	}
	require.Equal(t, pairs, expectedPairs)

	pairs, err = getUSDBasedProviders("BAR", providerPairs)
	require.Error(t, err)
}

func TestConvertCandlesToUSD(t *testing.T) {
	providerCandles := make(provider.AggregatedProviderCandles, 2)
	pairs := []types.CurrencyPair{
		{
			Base:  "ATOM",
			Quote: "USDT",
		},
		{
			Base:  "USDT",
			Quote: "USD",
		},
	}

	atomPrice := sdk.MustNewDecFromStr("29.93")
	atomVolume := sdk.MustNewDecFromStr("894123.00")

	usdtPrice := sdk.MustNewDecFromStr("0.98")
	usdtVolume := sdk.MustNewDecFromStr("894123.00")

	binanceCandles := make(map[string][]provider.CandlePrice, 1)
	binanceCandles["ATOM"] = []provider.CandlePrice{
		{
			Price:     atomPrice,
			Volume:    atomVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}
	providerCandles[config.ProviderBinance] = binanceCandles

	krakenCandles := make(map[string][]provider.CandlePrice, 1)
	krakenCandles["USDT"] = []provider.CandlePrice{
		{
			Price:     usdtPrice,
			Volume:    usdtVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}
	providerCandles[config.ProviderKraken] = krakenCandles

	providerPairs := map[string][]types.CurrencyPair{
		config.ProviderBinance: {pairs[0]},
		config.ProviderKraken:  {pairs[1]},
	}

	convertedCandles, err := convertCandlesToUSD(providerCandles, providerPairs)
	require.NoError(t, err)

	require.Equal(
		t,
		atomPrice.Mul(usdtPrice),
		convertedCandles["binance"]["ATOM"][0].Price,
	)
}

func TestConvertTickersToUSD(t *testing.T) {
	providerPrices := make(provider.AggregatedProviderPrices, 2)
	pairs := []types.CurrencyPair{
		{
			Base:  "ATOM",
			Quote: "USDT",
		},
		{
			Base:  "USDT",
			Quote: "USD",
		},
	}

	atomPrice := sdk.MustNewDecFromStr("29.93")
	atomVolume := sdk.MustNewDecFromStr("894123.00")

	usdtPrice := sdk.MustNewDecFromStr("0.98")
	usdtVolume := sdk.MustNewDecFromStr("894123.00")

	binanceTickers := make(map[string]provider.TickerPrice, 1)
	binanceTickers["ATOM"] = provider.TickerPrice{
		Price:  atomPrice,
		Volume: atomVolume,
	}
	providerPrices[config.ProviderBinance] = binanceTickers

	krakenTicker := make(map[string]provider.TickerPrice, 1)
	krakenTicker["USDT"] = provider.TickerPrice{
		Price:  usdtPrice,
		Volume: usdtVolume,
	}
	providerPrices[config.ProviderKraken] = krakenTicker

	providerPairs := map[string][]types.CurrencyPair{
		config.ProviderBinance: {pairs[0]},
		config.ProviderKraken:  {pairs[1]},
	}

	convertedTickers, err := convertTickersToUSD(providerPrices, providerPairs)
	require.NoError(t, err)

	require.Equal(
		t,
		atomPrice.Mul(usdtPrice),
		convertedTickers["binance"]["ATOM"].Price,
	)
}
