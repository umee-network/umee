package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Provider defines an interface an exchange price provider must implement.
type Provider interface {
	GetTickerPrices(ticker ...string) (map[string]sdk.Dec, error)
}
