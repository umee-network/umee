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
		make(map[string]sdk.Dec),
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
		make(map[string]sdk.Dec),
	)

	_, ok := pricesFiltered[config.ProviderCoinbase]
	require.NoError(t, err, "It should successfully filter out the provider using tickers")
	require.False(t, ok, "The filtered ticker deviation price at coinbase should be empty")
}
