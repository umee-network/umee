package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

func TestKrakenProvider_GetTickerPrices(t *testing.T) {
	p := NewKrakenProvider()

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/0/public/Ticker?pair=ATOMUSD", req.URL.String())
			resp := `{
				"error": [],
				"result": {
					"ATOMUSD": {
						"c": ["35.0872000", "0.32546988"],
						"v": ["1920.83610601", "7954.00219674"]
					}
				}
			}
			`
			rw.Write([]byte(resp))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USD"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, sdk.MustNewDecFromStr("35.0872"), prices["ATOMUSD"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("7954.00219674"), prices["ATOMUSD"].Volume)
	})

	t.Run("valid_request_multi_ticker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/0/public/Ticker?pair=ATOMUSD,XXBTZUSD", req.URL.String())
			resp := `{
				"error": [],
				"result": {
					"ATOMUSD": {
						"c": ["35.0872000", "0.32546988"],
						"v": ["1920.83610601", "7954.00219674"]
					},
					"XXBTZUSD": {
						"c": ["63339.40000", "0.00010000"],
						"v": ["1920.83610601", "7954.00219674"]
					}
				}
			}
			`
			rw.Write([]byte(resp))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices(
			types.CurrencyPair{Base: "ATOM", Quote: "USD"},
			types.CurrencyPair{Base: "XXBTZ", Quote: "USD"},
		)
		require.NoError(t, err)
		require.Len(t, prices, 2)
		require.Equal(t, sdk.MustNewDecFromStr("35.0872"), prices["ATOMUSD"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("7954.00219674"), prices["ATOMUSD"].Volume)
		require.Equal(t, sdk.MustNewDecFromStr("63339.40000"), prices["XXBTZUSD"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("7954.00219674"), prices["XXBTZUSD"].Volume)
	})

	t.Run("invalid_request_bad_response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/0/public/Ticker?pair=ATOMUSD", req.URL.String())
			rw.Write([]byte(`FOO`))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USD"})
		require.Error(t, err)
		require.Nil(t, prices)
	})

	t.Run("invalid_request_invalid_ticker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/0/public/Ticker?pair=FOOBAR", req.URL.String())
			resp := `{
				"error": ["EQuery:Unknown asset pair"]
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
