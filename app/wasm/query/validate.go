package query

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
)

// Validate checks the GetExchangeRateBase field
func (getExchangeRateBase *GetExchangeRateBase) Validate() error {
	if len(getExchangeRateBase.Denom) == 0 {
		return wasmvmtypes.UnsupportedRequest{Kind: "query 'getExchangeRateBase' with empty denom"}
	}

	return nil
}
