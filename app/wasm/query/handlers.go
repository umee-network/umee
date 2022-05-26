package query

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Handler handles any query that implement the QueryHandler interface
func (umeeQuery UmeeQuery) Handler(ctx sdk.Context, keepers Keepers, query Handler) ([]byte, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	resp, err := query.QueryResponse(ctx, keepers)
	if err != nil {
		return nil, err
	}

	bz, err := json.Marshal(resp)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "umee get exchange rate base response marshal")
	}

	return bz, nil
}

// HandleGetBorrow handles the get borrow query and response
func (umeeQuery UmeeQuery) HandleGetBorrow(ctx sdk.Context, keepers Keepers) ([]byte, error) {
	getBorrow := umeeQuery.GetBorrow
	if getBorrow == nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "assigned query GetBorrow with 'getBorrow' empty"}
	}

	return umeeQuery.Handler(ctx, keepers, getBorrow)
}

// HandleGetExchangeRateBase handles the get exchange rate base query and response
func (umeeQuery UmeeQuery) HandleGetExchangeRateBase(ctx sdk.Context, keepers Keepers) ([]byte, error) {
	getExchangeRateBase := umeeQuery.GetExchangeRateBase
	if getExchangeRateBase == nil {
		return nil, wasmvmtypes.UnsupportedRequest{
			Kind: "assigned query GetExchangeRateBase with 'getExchangeRateBase' empty",
		}
	}

	return umeeQuery.Handler(ctx, keepers, getExchangeRateBase)
}
