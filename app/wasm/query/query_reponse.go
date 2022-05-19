package query

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// QueryResponse Calls the keeper and build the response
func (getBorrow *GetBorrow) QueryResponse(ctx sdk.Context, keepers Keepers) (interface{}, error) {
	return GetBorrowResponse{
		BorrowedAmount: keepers.GetBorrow(ctx, getBorrow.BorrowerAddr, getBorrow.Denom),
	}, nil
}

// QueryResponse Calls the keeper and build the response
func (getExchangeRateBase *GetExchangeRateBase) QueryResponse(ctx sdk.Context, keepers Keepers) (interface{}, error) {
	exchangeRateBase, err := keepers.GetExchangeRateBase(ctx, getExchangeRateBase.Denom)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "umee get exchange rate base query")
	}

	return GetExchangeRateBaseResponse{
		ExchangeRateBase: exchangeRateBase,
	}, nil
}
