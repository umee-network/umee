package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

func TestHuobiProvider_GetTickerPrices(t *testing.T) {
	p := NewHuobiProvider()

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		var count int
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if count == 0 {
				require.Equal(t, "/market/trade?symbol=atomusdt", req.URL.String())
				resp := `{
					"status": "ok",
					"tick": {
						"data": [
							{
								"price": 28.0991
							}
						]
					}
				}
				`
				rw.Write([]byte(resp))
			} else {
				require.Equal(t, "/market/detail?symbol=atomusdt", req.URL.String())
				resp := `{
					"status": "ok",
					"tick": {
						"vol": 16511168.890881622
					}
				}
				`
				rw.Write([]byte(resp))
			}

			count++
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, sdk.MustNewDecFromStr("28.0991"), prices["ATOMUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("16511168.890881622"), prices["ATOMUSDT"].Volume)
	})

	t.Run("valid_request_multi_ticker", func(t *testing.T) {
		var count int
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			switch count {
			case 0:
				require.Equal(t, "/market/trade?symbol=atomusdt", req.URL.String())
				resp := `{
					"status": "ok",
					"tick": {
						"data": [
							{
								"price": 28.0991
							}
						]
					}
				}
				`
				rw.Write([]byte(resp))

			case 1:
				require.Equal(t, "/market/detail?symbol=atomusdt", req.URL.String())
				resp := `{
					"status": "ok",
					"tick": {
						"vol": 16511168.890881622
					}
				}
				`
				rw.Write([]byte(resp))

			case 2:
				require.Equal(t, "/market/trade?symbol=lunausdt", req.URL.String())
				resp := `{
					"status": "ok",
					"tick": {
						"data": [
							{
								"price": 99.5148
							}
						]
					}
				}
				`
				rw.Write([]byte(resp))

			case 3:
				require.Equal(t, "/market/detail?symbol=lunausdt", req.URL.String())
				resp := `{
					"status": "ok",
					"tick": {
						"vol": 162129738.74087432
					}
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
		require.Equal(t, sdk.MustNewDecFromStr("28.0991"), prices["ATOMUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("16511168.890881622"), prices["ATOMUSDT"].Volume)
		require.Equal(t, sdk.MustNewDecFromStr("99.5148"), prices["LUNAUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("162129738.74087432"), prices["LUNAUSDT"].Volume)
	})

	t.Run("invalid_request_bad_response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/market/trade?symbol=atomusdt", req.URL.String())
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
			require.Equal(t, "/market/trade?symbol=foobar", req.URL.String())
			resp := `{
				"status": "error",
				"err-code": "invalid-parameter",
				"err-msg": "invalid symbol"
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
