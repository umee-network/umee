package query

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetExchangeRateBase wraps the oracle GetExchangeRateBase query
type GetExchangeRateBase struct {
	Denom string `json:"denom"`
}

// GetExchangeRateBaseResponse wraps the response of GetExchangeRateBase query
type GetExchangeRateBaseResponse struct {
	ExchangeRateBase sdk.Dec `json:"exchange_rate_base"`
}
