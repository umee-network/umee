package types

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewExchangeRate creates a ExchangeRate instance
func NewExchangeRate(denom string, exchangeRate sdk.Dec, t time.Time) ExchangeRate {
	return ExchangeRate{
		Denom:     denom,
		Rate:      exchangeRate,
		Timestamp: t,
	}
}
