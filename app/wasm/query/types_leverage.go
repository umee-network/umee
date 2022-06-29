package query

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetBorrow wraps the leverage GetBorrow query.
type GetBorrow struct {
	BorrowerAddr sdk.AccAddress `json:"borrower_addr"`
	Denom        string         `json:"denom"`
}

// GetBorrowResponse wraps the response of GetBorrow query.
type GetBorrowResponse struct {
	BorrowedAmount sdk.Coin `json:"borrowed_amount"`
}
