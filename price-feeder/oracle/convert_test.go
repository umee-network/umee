package oracle

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
	"github.com/umee-network/umee/price-feeder/oracle/types"
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
