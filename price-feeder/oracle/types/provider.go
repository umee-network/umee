package types

const (
	ProviderKraken   ProviderName = "kraken"
	ProviderBinance  ProviderName = "binance"
	ProviderOsmosis  ProviderName = "osmosis"
	ProviderHuobi    ProviderName = "huobi"
	ProviderOkx      ProviderName = "okx"
	ProviderGate     ProviderName = "gate"
	ProviderCoinbase ProviderName = "coinbase"
	ProviderMock     ProviderName = "mock"
)

// ProviderName name of an oracle provider. Usually it is an exchange
// but this can be any provider name that can give token prices
// examples.: "binance", "osmosis", "kraken".
type ProviderName string

// String returns the provider name as string.
func (pn ProviderName) String() string {
	return string(pn)
}
