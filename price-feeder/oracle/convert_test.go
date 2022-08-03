package oracle

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
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
	providerPairs := make(map[types.ProviderName][]types.CurrencyPair, 3)
	providerPairs[types.ProviderCoinbase] = []types.CurrencyPair{
		{
			Base:  "FOO",
			Quote: "USD",
		},
	}
	providerPairs[types.ProviderHuobi] = []types.CurrencyPair{
		{
			Base:  "FOO",
			Quote: "USD",
		},
	}
	providerPairs[types.ProviderKraken] = []types.CurrencyPair{
		{
			Base:  "FOO",
			Quote: "USDT",
		},
	}
	providerPairs[types.ProviderBinance] = []types.CurrencyPair{
		{
			Base:  "USDT",
			Quote: "USD",
		},
	}

	pairs, err := getUSDBasedProviders("FOO", providerPairs)
	require.NoError(t, err)
	expectedPairs := map[types.ProviderName]struct{}{
		types.ProviderCoinbase: {},
		types.ProviderHuobi:    {},
	}
	require.Equal(t, pairs, expectedPairs)

	pairs, err = getUSDBasedProviders("USDT", providerPairs)
	require.NoError(t, err)
	expectedPairs = map[types.ProviderName]struct{}{
		types.ProviderBinance: {},
	}
	require.Equal(t, pairs, expectedPairs)

	_, err = getUSDBasedProviders("BAR", providerPairs)
	require.Error(t, err)
}

func TestConvertCandlesToUSD(t *testing.T) {
	providerCandles := make(provider.AggregatedProviderCandles, 2)

	binanceCandles := map[string][]provider.CandlePrice{
		"ATOM": {{
			Price:     atomPrice,
			Volume:    atomVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		}},
	}
	providerCandles[types.ProviderBinance] = binanceCandles

	krakenCandles := map[string][]provider.CandlePrice{
		"USDT": {{
			Price:     usdtPrice,
			Volume:    usdtVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		}},
	}
	providerCandles[types.ProviderKraken] = krakenCandles

	providerPairs := map[types.ProviderName][]types.CurrencyPair{
		types.ProviderBinance: {atomPair},
		types.ProviderKraken:  {usdtPair},
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
		convertedCandles[types.ProviderBinance]["ATOM"][0].Price,
	)
}

func TestConvertCandlesToUSDFiltering(t *testing.T) {
	providerCandles := make(provider.AggregatedProviderCandles, 2)

	binanceCandles := map[string][]provider.CandlePrice{
		"ATOM": {{
			Price:     atomPrice,
			Volume:    atomVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		}},
	}
	providerCandles[types.ProviderBinance] = binanceCandles

	krakenCandles := map[string][]provider.CandlePrice{
		"USDT": {{
			Price:     usdtPrice,
			Volume:    usdtVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		}},
	}
	providerCandles[types.ProviderKraken] = krakenCandles

	gateCandles := map[string][]provider.CandlePrice{
		"USDT": {{
			Price:     usdtPrice,
			Volume:    usdtVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		}},
	}
	providerCandles[types.ProviderGate] = gateCandles

	okxCandles := map[string][]provider.CandlePrice{
		"USDT": {{
			Price:     sdk.MustNewDecFromStr("100.0"),
			Volume:    usdtVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		}},
	}
	providerCandles[types.ProviderOkx] = okxCandles

	providerPairs := map[types.ProviderName][]types.CurrencyPair{
		types.ProviderBinance: {atomPair},
		types.ProviderKraken:  {usdtPair},
		types.ProviderGate:    {usdtPair},
		types.ProviderOkx:     {usdtPair},
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
		convertedCandles[types.ProviderBinance]["ATOM"][0].Price,
	)
}

func TestConvertTickersToUSD(t *testing.T) {
	providerPrices := make(provider.AggregatedProviderPrices, 2)

	binanceTickers := map[string]provider.TickerPrice{
		"ATOM": {
			Price:  atomPrice,
			Volume: atomVolume,
		},
	}
	providerPrices[types.ProviderBinance] = binanceTickers

	krakenTicker := map[string]provider.TickerPrice{
		"USDT": {
			Price:  usdtPrice,
			Volume: usdtVolume,
		},
	}
	providerPrices[types.ProviderKraken] = krakenTicker

	providerPairs := map[types.ProviderName][]types.CurrencyPair{
		types.ProviderBinance: {atomPair},
		types.ProviderKraken:  {usdtPair},
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
		convertedTickers[types.ProviderBinance]["ATOM"].Price,
	)
}

func TestConvertTickersToUSDFiltering(t *testing.T) {
	providerPrices := make(provider.AggregatedProviderPrices, 2)

	binanceTickers := map[string]provider.TickerPrice{
		"ATOM": {
			Price:  atomPrice,
			Volume: atomVolume,
		},
	}
	providerPrices[types.ProviderBinance] = binanceTickers

	krakenTicker := map[string]provider.TickerPrice{
		"USDT": {
			Price:  usdtPrice,
			Volume: usdtVolume,
		},
	}
	providerPrices[types.ProviderKraken] = krakenTicker

	gateTicker := map[string]provider.TickerPrice{
		"USDT": krakenTicker["USDT"],
	}
	providerPrices[types.ProviderGate] = gateTicker

	huobiTicker := map[string]provider.TickerPrice{
		"USDT": {
			Price:  sdk.MustNewDecFromStr("10000"),
			Volume: usdtVolume,
		},
	}
	providerPrices[types.ProviderHuobi] = huobiTicker

	providerPairs := map[types.ProviderName][]types.CurrencyPair{
		types.ProviderBinance: {atomPair},
		types.ProviderKraken:  {usdtPair},
		types.ProviderGate:    {usdtPair},
		types.ProviderHuobi:   {usdtPair},
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
		covertedDeviation[types.ProviderBinance]["ATOM"].Price,
	)
}
