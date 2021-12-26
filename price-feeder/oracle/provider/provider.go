package provider

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	defaultTimeout = 10 * time.Second
)

// Provider defines an interface an exchange price provider must implement.
type Provider interface {
	GetTickerPrices(...types.CurrencyPair) (map[string]TickerPrice, error)
}

// TickerPrice defines price and volume information for a symbol or ticker
// exchange rate.
type TickerPrice struct {
	Price  sdk.Dec // last trade price
	Volume sdk.Dec // 24h volume
}
