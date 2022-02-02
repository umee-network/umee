package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

func TestBinanceProvider_GetTickerPrices(t *testing.T) {
	p := NewBinanceProvider()

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/api/v3/ticker/24hr?symbol=ATOMUSDT", req.URL.String())
			resp := `{
				"symbol": "ATOMUSDT",
				"lastPrice": "34.69000000",
				"volume": "2396974.02000000"
			}
			`
			rw.Write([]byte(resp))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL
		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, sdk.MustNewDecFromStr("34.69000000"), prices["ATOMUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("2396974.02000000"), prices["ATOMUSDT"].Volume)
	})

	t.Run("valid_request_multi_ticker", func(t *testing.T) {
		var count int
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if count == 0 {
				require.Equal(t, "/api/v3/ticker/24hr?symbol=ATOMUSDT", req.URL.String())
				resp := `{
					"symbol": "ATOMUSDT",
					"lastPrice": "34.69000000",
					"volume": "2396974.02000000"
				}
				`
				rw.Write([]byte(resp))
			} else {
				require.Equal(t, "/api/v3/ticker/24hr?symbol=LUNAUSDT", req.URL.String())
				resp := `{
					"symbol": "LUNAUSDT",
					"lastPrice": "41.35000000",
					"volume": "2396974.02000000"
				}
				`
				rw.Write([]byte(resp))
			}

			count++
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices(
			types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
			types.CurrencyPair{Base: "LUNA", Quote: "USDT"},
		)
		require.NoError(t, err)
		require.Len(t, prices, 2)
		require.Equal(t, sdk.MustNewDecFromStr("34.69000000"), prices["ATOMUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("2396974.02000000"), prices["ATOMUSDT"].Volume)
		require.Equal(t, sdk.MustNewDecFromStr("41.35000000"), prices["LUNAUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("2396974.02000000"), prices["LUNAUSDT"].Volume)
	})

	t.Run("invalid_request_bad_response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/api/v3/ticker/24hr?symbol=ATOMUSDT", req.URL.String())
			rw.Write([]byte(`FOO`))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
		require.Error(t, err)
		require.Nil(t, prices)
	})

	t.Run("invalid_request_invalid_ticker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/api/v3/ticker/24hr?symbol=FOOBAR", req.URL.String())
			resp := `{
				"code": -1121,
				"msg": "Invalid symbol."
			}
			`
			rw.Write([]byte(resp))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "FOO", Quote: "BAR"})
		require.Error(t, err)
		require.Nil(t, prices)
	})

	t.Run("check_redirect", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			http.Redirect(rw, r, p.baseURL, http.StatusTemporaryRedirect)
		}))
		defer server.Close()

		server.Client().CheckRedirect = preventRedirect
		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
		require.Error(t, err)
		require.Nil(t, prices)
	})
}
