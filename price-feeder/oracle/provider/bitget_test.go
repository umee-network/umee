package provider

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

func TestBitgetProvider_GetTickerPrices(t *testing.T) {
	p, err := NewBitgetProvider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "BTC", Quote: "USDT"},
	)
	require.NoError(t, err)

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		lastPrice := "34.69000000"
		volume := "2396974.02000000"
		instId := "ATOMUSDT"

		tickerMap := map[string]BitgetTicker{}
		tickerMap[instId] = BitgetTicker{
			Arg: BitgetSubscriptionArg{
				Channel: "tickers",
				InstID:  instId,
			},
			Data: []BitgetTickerData{
				{
					InstID: instId,
					Price:  lastPrice,
					Volume: volume,
				},
			},
		}

		p.tickers = tickerMap

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, sdk.MustNewDecFromStr(lastPrice), prices["ATOMUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr(volume), prices["ATOMUSDT"].Volume)
	})

	t.Run("valid_request_multi_ticker", func(t *testing.T) {
		atomInstID := "ATOMUSDT"
		atomLastPrice := "34.69000000"
		lunaInstID := "LUNAUSDT"
		lunaLastPrice := "41.35000000"
		volume := "2396974.02000000"

		tickerMap := map[string]BitgetTicker{}
		tickerMap[atomInstID] = BitgetTicker{
			Arg: BitgetSubscriptionArg{
				Channel: "tickers",
				InstID:  atomInstID,
			},
			Data: []BitgetTickerData{
				{
					InstID: atomInstID,
					Price:  atomLastPrice,
					Volume: volume,
				},
			},
		}
		tickerMap[lunaInstID] = BitgetTicker{
			Arg: BitgetSubscriptionArg{
				Channel: "tickers",
				InstID:  lunaInstID,
			},
			Data: []BitgetTickerData{
				{
					InstID: lunaInstID,
					Price:  lunaLastPrice,
					Volume: volume,
				},
			},
		}
		p.tickers = tickerMap
		prices, err := p.GetTickerPrices(
			types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
			types.CurrencyPair{Base: "LUNA", Quote: "USDT"},
		)

		require.NoError(t, err)
		require.Len(t, prices, 2)
		require.Equal(t, sdk.MustNewDecFromStr(atomLastPrice), prices["ATOMUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr(volume), prices["ATOMUSDT"].Volume)
		require.Equal(t, sdk.MustNewDecFromStr(lunaLastPrice), prices["LUNAUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr(volume), prices["LUNAUSDT"].Volume)
	})

	t.Run("invalid_request_invalid_ticker", func(t *testing.T) {
		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "FOO", Quote: "BAR"})
		require.EqualError(t, err, "bitget failed to get ticker price for FOOBAR")
		require.Nil(t, prices)
	})
}

func TestBitgetProvider_SubscribeCurrencyPairs(t *testing.T) {
	p, err := NewBitgetProvider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
	)
	require.NoError(t, err)

	t.Run("invalid_subscribe_channels_empty", func(t *testing.T) {
		err = p.SubscribeCurrencyPairs([]types.CurrencyPair{}...)
		require.ErrorContains(t, err, "currency pairs is empty")
	})
}

func TestBitgetCurrencyPairToBitgetPair(t *testing.T) {
	cp := types.CurrencyPair{Base: "ATOM", Quote: "USDT"}
	bitgetSymbol := cp.String()
	require.Equal(t, bitgetSymbol, "ATOM/USDT")
}
