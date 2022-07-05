package query

import (
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	lvtypes "github.com/umee-network/umee/v2/x/leverage/types"
)

// HandleBorrowed handles the get of every borrowed value of an address.
func (umeeQuery UmeeQuery) HandleBorrowed(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.Borrowed(sdk.WrapSDKContext(ctx), umeeQuery.Borrowed)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Borrowed", err)}
	}

	return MarshalResponse(resp)
}

// HandleRegisteredTokens handles the get all registered tokens query and response.
func (umeeQuery UmeeQuery) HandleRegisteredTokens(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.RegisteredTokens(sdk.WrapSDKContext(ctx), &lvtypes.QueryRegisteredTokens{})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query RegisteredTokens", err)}
	}

	return MarshalResponse(resp)
}

// HandleLeverageParams handles the get the x/leverage module's parameters.
func (umeeQuery UmeeQuery) HandleLeverageParams(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.Params(sdk.WrapSDKContext(ctx), &lvtypes.QueryParamsRequest{})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query LeverageParams", err)}
	}

	return MarshalResponse(resp)
}

// HandleBorrowedValue handles the borrowed value of an specific denom and address.
func (umeeQuery UmeeQuery) HandleBorrowedValue(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.BorrowedValue(sdk.WrapSDKContext(ctx), umeeQuery.BorrowedValue)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query BorrowedValue", err)}
	}

	return MarshalResponse(resp)
}

// HandleLoaned handles the loaned amount of an address.
func (umeeQuery UmeeQuery) HandleLoaned(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.Loaned(sdk.WrapSDKContext(ctx), umeeQuery.Loaned)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Loaned", err)}
	}

	return MarshalResponse(resp)
}

// HandleLoanedValue handles the loaned amount of an address in USD.
func (umeeQuery UmeeQuery) HandleLoanedValue(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.LoanedValue(sdk.WrapSDKContext(ctx), umeeQuery.LoanedValue)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Loaned Value", err)}
	}

	return MarshalResponse(resp)
}

// HandleAvailableBorrow retrieves the available borrow amoun of an denom.
func (umeeQuery UmeeQuery) HandleAvailableBorrow(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.AvailableBorrow(sdk.WrapSDKContext(ctx), umeeQuery.AvailableBorrow)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Available Borrow", err)}
	}

	return MarshalResponse(resp)
}

// HandleBorrowAPY retrieves the current borrow interest rate on a token denom.
func (umeeQuery UmeeQuery) HandleBorrowAPY(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.BorrowAPY(sdk.WrapSDKContext(ctx), umeeQuery.BorrowAPY)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Borrow APY", err)}
	}

	return MarshalResponse(resp)
}

// HandleLendAPY derives the current lend interest rate on a token denom.
func (umeeQuery UmeeQuery) HandleLendAPY(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.LendAPY(sdk.WrapSDKContext(ctx), umeeQuery.LendAPY)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Lend APY", err)}
	}

	return MarshalResponse(resp)
}

// HandleMarketSize get the market size in USD of a token denom.
func (umeeQuery UmeeQuery) HandleMarketSize(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.MarketSize(sdk.WrapSDKContext(ctx), umeeQuery.MarketSize)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Market Size", err)}
	}

	return MarshalResponse(resp)
}

// HandleTokenMarketSize handles the market size of an token.
func (umeeQuery UmeeQuery) HandleTokenMarketSize(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.TokenMarketSize(sdk.WrapSDKContext(ctx), umeeQuery.TokenMarketSize)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Token Market Size", err)}
	}

	return MarshalResponse(resp)
}

// HandleReserveAmount gets the amount reserved of a specified token.
func (umeeQuery UmeeQuery) HandleReserveAmount(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.ReserveAmount(sdk.WrapSDKContext(ctx), umeeQuery.ReserveAmount)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Reserve Amount Size", err)}
	}

	return MarshalResponse(resp)
}

// HandleExchangeRate gets calculated the token:uToken exchange
// rate of a base token denom.
func (umeeQuery UmeeQuery) HandleExchangeRate(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.ExchangeRate(sdk.WrapSDKContext(ctx), umeeQuery.ExchangeRate)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Exchange Rate", err)}
	}

	return MarshalResponse(resp)
}

// HandleBorrowLimit uses the price oracle to determine the borrow limit
// (in USD) provided by collateral sdk.Coins.
func (umeeQuery UmeeQuery) HandleBorrowLimit(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.BorrowLimit(sdk.WrapSDKContext(ctx), umeeQuery.BorrowLimit)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Borrow Limit", err)}
	}

	return MarshalResponse(resp)
}

// HandleLiquidationThreshold determines the maximum borrowed value (in USD) that a borrower with given
// collateral could reach before being eligible for liquidation.
func (umeeQuery UmeeQuery) HandleLiquidationThreshold(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.LiquidationThreshold(sdk.WrapSDKContext(ctx), umeeQuery.LiquidationThreshold)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{
			Kind: fmt.Sprintf("error %+v to assigned query Liquidation Threshold", err),
		}
	}

	return MarshalResponse(resp)
}

// HandleLiquidationTargets determines an list of borrower addresses eligible for liquidation.
func (umeeQuery UmeeQuery) HandleLiquidationTargets(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.LiquidationTargets(sdk.WrapSDKContext(ctx), umeeQuery.LiquidationTargets)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Liquidation Targets", err)}
	}

	return MarshalResponse(resp)
}

// HandleMarketSummary gets the summary data of an denom.
func (umeeQuery UmeeQuery) HandleMarketSummary(
	ctx sdk.Context,
	queryServer lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.MarketSummary(sdk.WrapSDKContext(ctx), umeeQuery.MarketSummary)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Market Summary", err)}
	}

	return MarshalResponse(resp)
}
