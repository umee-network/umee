package provider

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

func TestOkxProvider_GetTickerPrices(t *testing.T) {
	p, err := NewOkxProvider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "BTC", Quote: "USDT"},
	)
	require.NoError(t, err)

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		lastPrice := "34.69000000"
		volume := "2396974.02000000"

		syncMap := map[string]OkxTickerPair{}
		syncMap["ATOM-USDT"] = OkxTickerPair{
			OkxInstID: OkxInstID{
				InstID: "ATOM-USDT",
			},
			Last:   lastPrice,
			Vol24h: volume,
		}

		p.tickers = syncMap

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, sdk.MustNewDecFromStr(lastPrice), prices["ATOMUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr(volume), prices["ATOMUSDT"].Volume)
	})

	t.Run("valid_request_multi_ticker", func(t *testing.T) {
		lastPriceAtom := "34.69000000"
		lastPriceLuna := "41.35000000"
		volume := "2396974.02000000"

		syncMap := map[string]OkxTickerPair{}
		syncMap["ATOM-USDT"] = OkxTickerPair{
			OkxInstID: OkxInstID{
				InstID: "ATOM-USDT",
			},
			Last:   lastPriceAtom,
			Vol24h: volume,
		}

		syncMap["LUNA-USDT"] = OkxTickerPair{
			OkxInstID: OkxInstID{
				InstID: "LUNA-USDT",
			},
			Last:   lastPriceLuna,
			Vol24h: volume,
		}

		p.tickers = syncMap
		prices, err := p.GetTickerPrices(
			types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
			types.CurrencyPair{Base: "LUNA", Quote: "USDT"},
		)
		require.NoError(t, err)
		require.Len(t, prices, 2)
		require.Equal(t, sdk.MustNewDecFromStr(lastPriceAtom), prices["ATOMUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr(volume), prices["ATOMUSDT"].Volume)
		require.Equal(t, sdk.MustNewDecFromStr(lastPriceLuna), prices["LUNAUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr(volume), prices["LUNAUSDT"].Volume)
	})

	t.Run("invalid_request_invalid_ticker", func(t *testing.T) {
		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "FOO", Quote: "BAR"})
		require.EqualError(t, err, "okx failed to get ticker price for FOO-BAR")
		require.Nil(t, prices)
	})
}

func TestOkxProvider_SubscribeCurrencyPairs(t *testing.T) {
	p, err := NewOkxProvider(
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

func TestOkxCurrencyPairToOkxPair(t *testing.T) {
	cp := types.CurrencyPair{Base: "ATOM", Quote: "USDT"}
	okxSymbol := currencyPairToOkxPair(cp)
	require.Equal(t, okxSymbol, "ATOM-USDT")
}
