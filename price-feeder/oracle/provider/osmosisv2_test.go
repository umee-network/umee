package provider

import (
	"context"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

func TestOsmosisV2Provider_GetTickerPrices(t *testing.T) {
	p, err := NewOsmosisV2Provider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "OSMO", Quote: "ATOM"},
	)
	require.NoError(t, err)

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		lastPrice := sdk.MustNewDecFromStr("34.69000000")
		volume := sdk.MustNewDecFromStr("2396974.02000000")

		tickerMap := map[string]types.TickerPrice{}
		tickerMap["OSMO/ATOM"] = types.TickerPrice{
			Price:  lastPrice,
			Volume: volume,
		}

		p.tickers = tickerMap

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "OSMO", Quote: "ATOM"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, lastPrice, prices["OSMOATOM"].Price)
		require.Equal(t, volume, prices["OSMOATOM"].Volume)
	})

	t.Run("valid_request_multi_ticker", func(t *testing.T) {
		lastPriceAtom := sdk.MustNewDecFromStr("34.69000000")
		lastPriceLuna := sdk.MustNewDecFromStr("41.35000000")
		volume := sdk.MustNewDecFromStr("2396974.02000000")

		tickerMap := map[string]types.TickerPrice{}
		tickerMap["ATOM/USDT"] = types.TickerPrice{
			Price:  lastPriceAtom,
			Volume: volume,
		}

		tickerMap["LUNA/USDT"] = types.TickerPrice{
			Price:  lastPriceLuna,
			Volume: volume,
		}

		p.tickers = tickerMap
		prices, err := p.GetTickerPrices(
			types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
			types.CurrencyPair{Base: "LUNA", Quote: "USDT"},
		)
		require.NoError(t, err)
		require.Len(t, prices, 2)
		require.Equal(t, lastPriceAtom, prices["ATOMUSDT"].Price)
		require.Equal(t, volume, prices["ATOMUSDT"].Volume)
		require.Equal(t, lastPriceLuna, prices["LUNAUSDT"].Price)
		require.Equal(t, volume, prices["LUNAUSDT"].Volume)
	})

	t.Run("invalid_request_invalid_ticker", func(t *testing.T) {
		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "FOO", Quote: "BAR"})
		require.Error(t, err)
		require.Equal(t, "osmosisv2 failed to get ticker price for FOO/BAR", err.Error())
		require.Nil(t, prices)
	})
}

func TestOsmosisV2Provider_GetCandlePrices(t *testing.T) {
	p, err := NewOsmosisV2Provider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "OSMO", Quote: "ATOM"},
	)
	require.NoError(t, err)

	t.Run("valid_request_single_candle", func(t *testing.T) {
		price := "34.689998626708984000"
		volume := "2396974.000000000000000000"
		time := int64(1000000)

		candle := OsmosisV2Candle{
			Volume:  volume,
			Close:   price,
			EndTime: time,
		}

		p.setCandlePair("OSMO/ATOM", candle)

		prices, err := p.GetCandlePrices(types.CurrencyPair{Base: "OSMO", Quote: "ATOM"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, sdk.MustNewDecFromStr(price), prices["OSMOATOM"][0].Price)
		require.Equal(t, sdk.MustNewDecFromStr(volume), prices["OSMOATOM"][0].Volume)
		require.Equal(t, time, prices["OSMOATOM"][0].TimeStamp)
	})

	t.Run("invalid_request_invalid_candle", func(t *testing.T) {
		prices, err := p.GetCandlePrices(types.CurrencyPair{Base: "FOO", Quote: "BAR"})
		require.EqualError(t, err, "osmosisv2 failed to get candle price for FOO/BAR")
		require.Nil(t, prices)
	})
}

func TestOsmosisV2Provider_SubscribeCurrencyPairs(t *testing.T) {
	p, err := NewOsmosisV2Provider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "OSMO", Quote: "ATOM"},
	)
	require.NoError(t, err)

	// wait for response from osmosis-api
	time.Sleep(time.Second * 10)

	t.Run("ticker_prices_set", func(t *testing.T) {
		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "OSMO", Quote: "ATOM"})
		require.NoError(t, err)
		require.NotEmpty(t, prices["OSMOATOM"])
	})

	t.Run("candle_prices_set", func(t *testing.T) {
		prices, err := p.GetCandlePrices(types.CurrencyPair{Base: "OSMO", Quote: "ATOM"})
		require.NoError(t, err)
		require.NotEmpty(t, prices["OSMOATOM"])
	})
}

func TestOsmosisV2CurrencyPairToOsmosisV2Pair(t *testing.T) {
	cp := types.CurrencyPair{Base: "ATOM", Quote: "USDT"}
	osmosisv2Symbol := currencyPairToOsmosisV2Pair(cp)
	require.Equal(t, osmosisv2Symbol, "ATOM/USDT")
}
