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

// HandleBorrowed handles the get of every borrowed value of an address.
func (umeeQuery UmeeQuery) HandleBorrowed(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
) ([]byte, error) {
	resp, err := querier.Borrowed(sdk.WrapSDKContext(ctx), umeeQuery.Borrowed)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Borrowed", err)}
	}

	return MarshalResponse(resp)
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

// HandleRegisteredTokens handles the get all registered tokens query and response.
func (umeeQuery UmeeQuery) HandleRegisteredTokens(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
) ([]byte, error) {
	resp, err := querier.RegisteredTokens(sdk.WrapSDKContext(ctx), &leveragetypes.QueryRegisteredTokens{})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query RegisteredTokens", err)}
	}

	return MarshalResponse(resp)
}

// HandleLeverageParams handles the get the x/leverage module's parameters.
func (umeeQuery UmeeQuery) HandleLeverageParams(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
) ([]byte, error) {
	resp, err := querier.Params(sdk.WrapSDKContext(ctx), &leveragetypes.QueryParamsRequest{})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query LeverageParams", err)}
	}

	return MarshalResponse(resp)
}

// HandleBorrowedValue handles the borrowed value of an specific denom and address.
func (umeeQuery UmeeQuery) HandleBorrowedValue(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
) ([]byte, error) {
	resp, err := querier.BorrowedValue(sdk.WrapSDKContext(ctx), umeeQuery.BorrowedValue)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query BorrowedValue", err)}
	}

	return MarshalResponse(resp)
}

// HandleLoaned handles the loaned amount of an address.
func (umeeQuery UmeeQuery) HandleLoaned(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
) ([]byte, error) {
	resp, err := querier.Loaned(sdk.WrapSDKContext(ctx), umeeQuery.Loaned)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Loaned", err)}
	}

	return MarshalResponse(resp)
}

// HandleLoanedValue handles the loaned amount of an address in USD.
func (umeeQuery UmeeQuery) HandleLoanedValue(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
) ([]byte, error) {
	resp, err := querier.LoanedValue(sdk.WrapSDKContext(ctx), umeeQuery.LoanedValue)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Loaned Value", err)}
	}

	return MarshalResponse(resp)
}

// HandleAvailableBorrow retrieves the available borrow amoun of an denom.
func (umeeQuery UmeeQuery) HandleAvailableBorrow(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
) ([]byte, error) {
	resp, err := querier.AvailableBorrow(sdk.WrapSDKContext(ctx), umeeQuery.AvailableBorrow)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Available Borrow", err)}
	}

	return MarshalResponse(resp)
}

// HandleBorrowAPY retrieves the current borrow interest rate on a token denom.
func (umeeQuery UmeeQuery) HandleBorrowAPY(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
) ([]byte, error) {
	resp, err := querier.BorrowAPY(sdk.WrapSDKContext(ctx), umeeQuery.BorrowAPY)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Borrow APY", err)}
	}

	return MarshalResponse(resp)
}

// HandleLendAPY derives the current lend interest rate on a token denom.
func (umeeQuery UmeeQuery) HandleLendAPY(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
) ([]byte, error) {
	resp, err := querier.LendAPY(sdk.WrapSDKContext(ctx), umeeQuery.LendAPY)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Lend APY", err)}
	}

	return MarshalResponse(resp)
}

// HandleMarketSize get the market size in USD of a token denom.
func (umeeQuery UmeeQuery) HandleMarketSize(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
) ([]byte, error) {
	resp, err := querier.MarketSize(sdk.WrapSDKContext(ctx), umeeQuery.MarketSize)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Market Size", err)}
	}

	return MarshalResponse(resp)
}

// HandleTokenMarketSize handles the market size of an token.
func (umeeQuery UmeeQuery) HandleTokenMarketSize(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
) ([]byte, error) {
	resp, err := querier.TokenMarketSize(sdk.WrapSDKContext(ctx), umeeQuery.TokenMarketSize)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Token Market Size", err)}
	}

	return MarshalResponse(resp)
}

// HandleReserveAmount gets the amount reserved of a specified token.
func (umeeQuery UmeeQuery) HandleReserveAmount(
	ctx sdk.Context,
	querier leveragekeeper.Querier,
) ([]byte, error) {
	resp, err := querier.ReserveAmount(sdk.WrapSDKContext(ctx), umeeQuery.ReserveAmount)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Reserve Amount Size", err)}
	}

	return MarshalResponse(resp)
}
