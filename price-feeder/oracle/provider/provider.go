package provider

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	defaultTimeout = 10 * time.Second
)

// Provider defines an interface an exchange price provider must implement.
type Provider interface {
	GetTickerPrices(ticker ...string) (map[string]TickerPrice, error)
}

// TickerPrice defines price and volume information for a symbol or ticker
// exchange rate.
type TickerPrice struct {
	Price  sdk.Dec
	Volume sdk.Int
}
