package query

import (
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
)

// Validate checks the GetBorrow fields
func (getBorrow *GetBorrow) Validate() error {
	baseMsg := "query 'getBorrow' with %s"

	if getBorrow.BorrowerAddr.Empty() {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf(baseMsg, "empty borrower address")}
	}

	if len(getBorrow.Denom) == 0 {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf(baseMsg, "empty denom")}
	}

	return nil
}

// Validate checks the GetExchangeRateBase field
func (getExchangeRateBase *GetExchangeRateBase) Validate() error {
	if len(getExchangeRateBase.Denom) == 0 {
		return wasmvmtypes.UnsupportedRequest{Kind: "query 'getExchangeRateBase' with empty denom"}
	}

	return nil
}

// Validate GetAllRegisteredTokens implements the iterface.
func (getAllRegisteredTokens *GetAllRegisteredTokens) Validate() error {
	return nil
}
