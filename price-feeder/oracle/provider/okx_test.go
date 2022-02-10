package provider

import (
	"testing"
)

func TestOkxProvider_GetTickerPrices(t *testing.T) {
	// ctx := context.TODO()
	// NewOkxProvider(ctx)
	// p := NewOkxProvider()

	// t.Run("valid_request_single_ticker", func(t *testing.T) {
	// 	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	// 		require.Equal(t, "/api/v5/market/ticker?instId=ATOM-USDT", req.URL.String())
	// 		resp := `{
	// 			"code": "0",
	// 			"msg": "",
	// 			"data": [
	// 				{
	// 					"instId": "ATOM-USDT",
	// 					"last": "34.69000000",
	// 					"vol24h": "2396974.02000000"
	// 				}
	// 			]
	// 		}`
	// 		rw.Write([]byte(resp))
	// 	}))
	// 	defer server.Close()

	// 	p.client = server.Client()
	// 	p.baseURL = server.URL

	// 	prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
	// 	require.NoError(t, err)
	// 	require.Len(t, prices, 1)
	// 	require.Equal(t, sdk.MustNewDecFromStr("34.69000000"), prices["ATOMUSDT"].Price)
	// 	require.Equal(t, sdk.MustNewDecFromStr("2396974.02000000"), prices["ATOMUSDT"].Volume)
	// })

	// t.Run("valid_request_multi_ticker", func(t *testing.T) {
	// 	var count int
	// 	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	// 		if count == 0 {
	// 			require.Equal(t, "/api/v5/market/ticker?instId=ATOM-USDT", req.URL.String())
	// 			resp := `{
	// 				"code": "0",
	// 				"msg": "",
	// 				"data": [
	// 					{
	// 						"instId": "ATOM-USDT",
	// 						"last": "34.69000000",
	// 						"vol24h": "2396974.02000000"
	// 					}
	// 				]
	// 			}`
	// 			rw.Write([]byte(resp))
	// 		} else {
	// 			require.Equal(t, "/api/v5/market/ticker?instId=LUNA-USDT", req.URL.String())
	// 			resp := `{
	// 				"code": "0",
	// 				"msg": "",
	// 				"data": [
	// 					{
	// 						"instId": "LUNA-USDT",
	// 						"last": "41.35000000",
	// 						"vol24h": "2396974.02000000"
	// 					}
	// 				]
	// 			}`
	// 			rw.Write([]byte(resp))
	// 		}

	// 		count++
	// 	}))
	// 	defer server.Close()

	// 	p.client = server.Client()
	// 	p.baseURL = server.URL

	// 	prices, err := p.GetTickerPrices(
	// 		types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
	// 		types.CurrencyPair{Base: "LUNA", Quote: "USDT"},
	// 	)
	// 	require.NoError(t, err)
	// 	require.Len(t, prices, 2)
	// 	require.Equal(t, sdk.MustNewDecFromStr("34.69000000"), prices["ATOMUSDT"].Price)
	// 	require.Equal(t, sdk.MustNewDecFromStr("2396974.02000000"), prices["ATOMUSDT"].Volume)
	// 	require.Equal(t, sdk.MustNewDecFromStr("41.35000000"), prices["LUNAUSDT"].Price)
	// 	require.Equal(t, sdk.MustNewDecFromStr("2396974.02000000"), prices["LUNAUSDT"].Volume)
	// })

	// t.Run("invalid_request_bad_response", func(t *testing.T) {
	// 	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	// 		require.Equal(t, "/api/v5/market/ticker?instId=ATOM-USDT", req.URL.String())
	// 		rw.Write([]byte(`FOO`))
	// 	}))
	// 	defer server.Close()

	// 	p.client = server.Client()
	// 	p.baseURL = server.URL

	// 	prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
	// 	require.Error(t, err)
	// 	require.Nil(t, prices)
	// })

	// t.Run("invalid_request_invalid_ticker", func(t *testing.T) {
	// 	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	// 		require.Equal(t, "/api/v5/market/ticker?instId=FOO-BAR", req.URL.String())
	// 		resp := `{
	// 			"code": "51012",
	// 			"msg": "Token does not exist.",
	// 			"data":[]
	// 		}`
	// 		rw.Write([]byte(resp))
	// 	}))
	// 	defer server.Close()

	// 	p.client = server.Client()
	// 	p.baseURL = server.URL

	// 	prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "FOO", Quote: "BAR"})
	// 	require.Error(t, err)
	// 	require.Nil(t, prices)
	// })

	// t.Run("check_redirect", func(t *testing.T) {
	// 	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
	// 		http.Redirect(rw, r, p.baseURL, http.StatusTemporaryRedirect)
	// 	}))
	// 	defer server.Close()

	// 	server.Client().CheckRedirect = preventRedirect
	// 	p.client = server.Client()
	// 	p.baseURL = server.URL

	// 	prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
	// 	require.Error(t, err)
	// 	require.Nil(t, prices)
	// })
}
