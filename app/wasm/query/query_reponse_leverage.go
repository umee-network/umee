package query

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueryResponse Calls the keeper and build the response.
func (getBorrow *GetBorrow) QueryResponse(ctx sdk.Context, keepers Keepers) (interface{}, error) {
	return GetBorrowResponse{
		BorrowedAmount: keepers.GetBorrow(ctx, getBorrow.BorrowerAddr, getBorrow.Denom),
	}, nil
}

// QueryResponse Calls the keeper and build the response.
func (getAllRegisteredTokens *GetAllRegisteredTokens) QueryResponse(
	ctx sdk.Context,
	keepers Keepers,
) (interface{}, error) {
	return GetAllRegisteredTokensResponse{
		Registry: keepers.GetAllRegisteredTokens(ctx),
	}, nil
}
