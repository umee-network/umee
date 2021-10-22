package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
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
						"c": ["35.0872000", "0.32546988"]
					}
				}
			}
			`
			rw.Write([]byte(resp))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices("ATOMUSD")
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, sdk.MustNewDecFromStr("35.0872"), prices["ATOMUSD"])
	})

	t.Run("valid_request_multi_ticker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/0/public/Ticker?pair=ATOMUSD,XXBTZUSD", req.URL.String())
			resp := `{
				"error": [],
				"result": {
					"ATOMUSD": {
						"c": ["35.0872000", "0.32546988"]
					},
					"XXBTZUSD": {
						"c": ["63339.40000", "0.00010000"]
					}
				}
			}
			`
			rw.Write([]byte(resp))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices("ATOMUSD", "XXBTZUSD")
		require.NoError(t, err)
		require.Len(t, prices, 2)
		require.Equal(t, sdk.MustNewDecFromStr("35.0872"), prices["ATOMUSD"])
		require.Equal(t, sdk.MustNewDecFromStr("63339.40000"), prices["XXBTZUSD"])
	})

	t.Run("invalid_request_bad_response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/0/public/Ticker?pair=ATOMUSD", req.URL.String())
			rw.Write([]byte(`FOO`))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices("ATOMUSD")
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

		prices, err := p.GetTickerPrices("FOOBAR")
		require.Error(t, err)
		require.Nil(t, prices)
	})
}
