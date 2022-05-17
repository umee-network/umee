package query

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AssignedQuery defines the query to be called
type AssignedQuery uint16

const (
	// AssignedQueryGetBorrow represents the call of leverage get borrow
	AssignedQueryGetBorrow = iota + 1
	// AssignedQueryGetExchangeRateBase represents the call of oracle get exchange rate
	AssignedQueryGetExchangeRateBase
)

// UmeeQuery wraps all the queries availables for cosmwasm smartcontracts
type UmeeQuery struct {
	// Mandatory field to determine which query to call
	AssignedQuery AssignedQuery `json:"assignedQuery"`
	// Used to query an open borrow position of an address for specific denom
	GetBorrow *GetBorrow `json:"getBorrow,omitempty"`
	// Used to query an exchange rate of a denom
	GetExchangeRateBase *GetExchangeRateBase `json:"getExchangeRateBase,omitempty"`
}

// GetBorrow wraps the leverage GetBorrow query
type GetBorrow struct {
	BorrowerAddr sdk.AccAddress `json:"borrowerAddr"`
	Denom        string         `json:"denom"`
}

// GetBorrowResponse wraps the response of GetBorrow query
type GetBorrowResponse struct {
	BorrowedAmount sdk.Coin `json:"borrowedAmount"`
}

// GetExchangeRateBase wraps the oracle GetExchangeRateBase query
type GetExchangeRateBase struct {
	Denom string `json:"denom"`
}

// GetExchangeRateBaseResponse wraps the response of GetExchangeRateBase query
type GetExchangeRateBaseResponse struct {
	ExchangeRateBase sdk.Dec `json:"exchangeRateBase"`
}

// Keepers wraps the interface to encapsulate keepers
type Keepers interface {
	// GetBorrow executes the GetBorrow from leverage keeper
	GetBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) sdk.Coin
	// GetExchangeRateBase executes the GetExchangeRateBase from oracle keeper
	GetExchangeRateBase(ctx sdk.Context, denom string) (sdk.Dec, error)
}
