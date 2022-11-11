package integration

import (
	"github.com/umee-network/umee/price-feeder/oracle/provider"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

var ProviderAndCurrencyPairsFixture = []struct {
	provider      provider.Name
	currencyPairs []types.CurrencyPair
}{
	{
		provider:      provider.ProviderBinance,
		currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
	},
	{
		provider:      provider.ProviderMexc,
		currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
	},
	{
		provider:      provider.ProviderKraken,
		currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
	},
	{
		provider: provider.ProviderOsmosisV2,
		currencyPairs: []types.CurrencyPair{
			{Base: "OSMO", Quote: "ATOM"},
			{Base: "ATOM", Quote: "JUNO"},
			{Base: "ATOM", Quote: "STARGAZE"},
			{Base: "OSMO", Quote: "WBTC"},
			{Base: "OSMO", Quote: "WETH"},
			{Base: "OSMO", Quote: "CRO"},
		},
	},
	{
		provider:      provider.ProviderCoinbase,
		currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
	},
	{
		provider:      provider.ProviderBitget,
		currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
	},
	{
		provider:      provider.ProviderGate,
		currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
	},
	{
		provider:      provider.ProviderOkx,
		currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
	},
	{
		provider:      provider.ProviderHuobi,
		currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USDT"}},
	},
	{
		provider:      provider.ProviderCrypto,
		currencyPairs: []types.CurrencyPair{{Base: "ATOM", Quote: "USD"}},
	},
}
