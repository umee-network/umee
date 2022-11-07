package provider

import (
	"context"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

func TestBinanceProvider_GetTickerPrices(t *testing.T) {
	p, err := NewBinanceProvider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
	)
	require.NoError(t, err)

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		lastPrice := "34.69000000"
		volume := "2396974.02000000"

		tickerMap := map[string]BinanceTicker{}
		tickerMap["ATOMUSDT"] = BinanceTicker{
			Symbol:    "ATOMUSDT",
			LastPrice: lastPrice,
			Volume:    volume,
		}

		p.tickers = tickerMap

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

		tickerMap := map[string]BinanceTicker{}
		tickerMap["ATOMUSDT"] = BinanceTicker{
			Symbol:    "ATOMUSDT",
			LastPrice: lastPriceAtom,
			Volume:    volume,
		}

		tickerMap["LUNAUSDT"] = BinanceTicker{
			Symbol:    "LUNAUSDT",
			LastPrice: lastPriceLuna,
			Volume:    volume,
		}

		p.tickers = tickerMap
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
		require.EqualError(t, err, "binance failed to get ticker price for FOOBAR")
		require.Nil(t, prices)
	})
}

func TestBinanceCurrencyPairToBinancePair(t *testing.T) {
	cp := types.CurrencyPair{Base: "ATOM", Quote: "USDT"}
	binanceSymbol := currencyPairToBinanceTickerPair(cp)
	require.Equal(t, binanceSymbol, "atomusdt@ticker")
}

func TestBinanceProvider_getSubscriptionMsgs(t *testing.T) {
	provider := &BinanceProvider{
		subscribedPairs: map[string]types.CurrencyPair{},
	}
	cps := []types.CurrencyPair{
		{Base: "ATOM", Quote: "USDT"},
	}
	provider.setSubscribedPairs(cps...)
	subMsgs := provider.getSubscriptionMsgs()

	msg, _ := json.Marshal(subMsgs[0])
	require.Equal(t, "{\"method\":\"SUBSCRIBE\",\"params\":[\"atomusdt@ticker\"],\"id\":1}", string(msg))

	msg, _ = json.Marshal(subMsgs[1])
	require.Equal(t, "{\"method\":\"SUBSCRIBE\",\"params\":[\"atomusdt@kline_1m\"],\"id\":1}", string(msg))
}
