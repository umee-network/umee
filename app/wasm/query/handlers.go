package query

import (
	"encoding/json"
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	leveragekeeper "github.com/umee-network/umee/v2/x/leverage/keeper"
	leveragetypes "github.com/umee-network/umee/v2/x/leverage/types"
)

// MarshalResponse marshals any response.
func MarshalResponse(resp interface{}) ([]byte, error) {
	bz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v umee query response error on marshal", err)}
	}
	return bz, err
}

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
		return nil, sdkerrors.Wrap(err, "umee query response error on marshal")
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

// HandleGetAllRegisteredTokens handles the get all registered tokens query and response.
func (umeeQuery UmeeQuery) HandleGetAllRegisteredTokens(ctx sdk.Context, keepers Keepers) ([]byte, error) {
	getAllRegisteredTokens := umeeQuery.GetAllRegisteredTokens
	return umeeQuery.Handler(ctx, keepers, getAllRegisteredTokens)
}

// HandleRegisteredTokens handles the get all registered tokens query and response.
func (umeeQuery UmeeQuery) HandleRegisteredTokens(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
	keepers Keepers,
) ([]byte, error) {
	resp, err := querier.RegisteredTokens(sdk.WrapSDKContext(ctx), &leveragetypes.QueryRegisteredTokens{})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query RegisteredTokens", err)}
	}

	return MarshalResponse(resp)
}
