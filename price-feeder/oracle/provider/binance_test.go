package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestBinanceProvider_GetTickerPrices(t *testing.T) {
	p := NewBinanceProvider()

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/api/v3/ticker/price?symbol=ATOMUSDT", req.URL.String())
			resp := `{
				"symbol": "ATOMUSDT",
				"price": "34.69000000"
			}			
			`
			rw.Write([]byte(resp))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices("ATOMUSDT")
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, sdk.MustNewDecFromStr("34.69000000"), prices["ATOMUSDT"])
	})

	t.Run("valid_request_multi_ticker", func(t *testing.T) {
		var count int
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if count == 0 {
				require.Equal(t, "/api/v3/ticker/price?symbol=ATOMUSDT", req.URL.String())
				resp := `{
					"symbol": "ATOMUSDT",
					"price": "34.69000000"
				}			
				`
				rw.Write([]byte(resp))
			} else {
				require.Equal(t, "/api/v3/ticker/price?symbol=LUNAUSDT", req.URL.String())
				resp := `{
					"symbol": "LUNAUSDT",
					"price": "41.35000000"
				}			
				`
				rw.Write([]byte(resp))
			}

			count++
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices("ATOMUSDT", "LUNAUSDT")
		require.NoError(t, err)
		require.Len(t, prices, 2)
		require.Equal(t, sdk.MustNewDecFromStr("34.69000000"), prices["ATOMUSDT"])
		require.Equal(t, sdk.MustNewDecFromStr("41.35000000"), prices["LUNAUSDT"])
	})

	t.Run("invalid_request_bad_response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/api/v3/ticker/price?symbol=ATOMUSDT", req.URL.String())
			rw.Write([]byte(`FOO`))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices("ATOMUSDT")
		require.Error(t, err)
		require.Nil(t, prices)
	})

	t.Run("invalid_request_invalid_ticker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/api/v3/ticker/price?symbol=FOOBAR", req.URL.String())
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

		prices, err := p.GetTickerPrices("FOOBAR")
		require.Error(t, err)
		require.Nil(t, prices)
	})
}
