package oracle

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

var (
	atomPrice  = sdk.MustNewDecFromStr("29.93")
	atomVolume = sdk.MustNewDecFromStr("894123.00")
	usdtPrice  = sdk.MustNewDecFromStr("0.98")
	usdtVolume = sdk.MustNewDecFromStr("894123.00")

	atomPair = types.CurrencyPair{
		Base:  "ATOM",
		Quote: "USDT",
	}
	usdtPair = types.CurrencyPair{
		Base:  "USDT",
		Quote: "USD",
	}
)

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
		config.ProviderBinance: {atomPair},
		config.ProviderKraken:  {usdtPair},
	}

	convertedCandles, err := convertCandlesToUSD(
		zerolog.Nop(),
		providerCandles,
		providerPairs,
		make(map[string]sdk.Dec),
	)
	require.NoError(t, err)

	require.Equal(
		t,
		atomPrice.Mul(usdtPrice),
		convertedCandles["binance"]["ATOM"][0].Price,
	)
}

func TestConvertCandlesToUSDFiltering(t *testing.T) {
	providerCandles := make(provider.AggregatedProviderCandles, 2)

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

	gateCandles := make(map[string][]provider.CandlePrice, 1)
	gateCandles["USDT"] = []provider.CandlePrice{
		{
			Price:     usdtPrice,
			Volume:    usdtVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}
	providerCandles[config.ProviderGate] = gateCandles

	okxCandles := make(map[string][]provider.CandlePrice, 1)
	okxCandles["USDT"] = []provider.CandlePrice{
		{
			Price:     sdk.MustNewDecFromStr("2.0"),
			Volume:    usdtVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}
	providerCandles[config.ProviderOkx] = okxCandles

	providerPairs := map[string][]types.CurrencyPair{
		config.ProviderBinance: {atomPair},
		config.ProviderKraken:  {usdtPair},
		config.ProviderGate:    {usdtPair},
		config.ProviderOkx:     {usdtPair},
	}

	convertedCandles, err := convertCandlesToUSD(
		zerolog.Nop(),
		providerCandles,
		providerPairs,
		make(map[string]sdk.Dec),
	)
	require.NoError(t, err)

	require.Equal(
		t,
		atomPrice.Mul(usdtPrice),
		convertedCandles["binance"]["ATOM"][0].Price,
	)
}

func TestConvertTickersToUSD(t *testing.T) {
	providerPrices := make(provider.AggregatedProviderPrices, 2)

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
		config.ProviderBinance: {atomPair},
		config.ProviderKraken:  {usdtPair},
	}

	convertedTickers, err := convertTickersToUSD(
		zerolog.Nop(),
		providerPrices,
		providerPairs,
		make(map[string]sdk.Dec),
	)
	require.NoError(t, err)

	require.Equal(
		t,
		atomPrice.Mul(usdtPrice),
		convertedTickers["binance"]["ATOM"].Price,
	)
}

func TestConvertTickersToUSDFiltering(t *testing.T) {
	providerPrices := make(provider.AggregatedProviderPrices, 2)

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

	gateTicker := make(map[string]provider.TickerPrice, 1)
	gateTicker["USDT"] = krakenTicker["USDT"]
	providerPrices[config.ProviderGate] = gateTicker

	huobiTicker := make(map[string]provider.TickerPrice, 1)
	huobiTicker["USDT"] = provider.TickerPrice{
		Price:  sdk.MustNewDecFromStr("10000"),
		Volume: usdtVolume,
	}
	providerPrices[config.ProviderHuobi] = huobiTicker

	providerPairs := map[string][]types.CurrencyPair{
		config.ProviderBinance: {atomPair},
		config.ProviderKraken:  {usdtPair},
		config.ProviderGate:    {usdtPair},
		config.ProviderHuobi:   {usdtPair},
	}

	covertedDeviation, err := convertTickersToUSD(
		zerolog.Nop(),
		providerPrices,
		providerPairs,
		make(map[string]sdk.Dec),
	)
	require.NoError(t, err)

	require.Equal(
		t,
		atomPrice.Mul(usdtPrice),
		covertedDeviation["binance"]["ATOM"].Price,
	)
}
