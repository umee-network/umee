package query

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// HandleGetBorrow handles the get borrow query and response
func (umeeQuery UmeeQuery) HandleGetBorrow(ctx sdk.Context, queryPlugin Keepers) ([]byte, error) {
	getBorrow := umeeQuery.GetBorrow
	if getBorrow == nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "assigned query GetBorrow with 'getBorrow' empty"}
	}

	if err := getBorrow.Validate(); err != nil {
		return nil, err
	}

	getBorrowResponse := GetBorrowResponse{
		BorrowedAmount: queryPlugin.GetBorrow(ctx, getBorrow.BorrowerAddr, getBorrow.Denom),
	}

	bz, err := json.Marshal(getBorrowResponse)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "umee get borrow response marshal")
	}

	return bz, nil
}

// HandleGetExchangeRateBase handles the get exchange rate base query and response
func (umeeQuery UmeeQuery) HandleGetExchangeRateBase(ctx sdk.Context, keepers Keepers) ([]byte, error) {
	getExchangeRateBase := umeeQuery.GetExchangeRateBase
	if getExchangeRateBase == nil {
		return nil, wasmvmtypes.UnsupportedRequest{
			Kind: "assigned query GetExchangeRateBase with 'getExchangeRateBase' empty",
		}
	}

	if err := getExchangeRateBase.Validate(); err != nil {
		return nil, err
	}

	exchangeRateBase, err := keepers.GetExchangeRateBase(ctx, getExchangeRateBase.Denom)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "umee get exchange rate base query")
	}

	getExchangeRateBaseResponse := GetExchangeRateBaseResponse{
		ExchangeRateBase: exchangeRateBase,
	}

	bz, err := json.Marshal(getExchangeRateBaseResponse)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "umee get exchange rate base response marshal")
	}

	return bz, nil
}
