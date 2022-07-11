package query

import (
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	lvtypes "github.com/umee-network/umee/v2/x/leverage/types"
)

// HandleBorrowed handles the get of every borrowed value of an address.
func (q UmeeQuery) HandleBorrowed(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.Borrowed(sdk.WrapSDKContext(ctx), q.Borrowed)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Borrowed", err)}
	}

	return MarshalResponse(resp)
}

// HandleRegisteredTokens handles the get all registered tokens query and response.
func (q UmeeQuery) HandleRegisteredTokens(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.RegisteredTokens(sdk.WrapSDKContext(ctx), &lvtypes.QueryRegisteredTokens{})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Registered Tokens", err)}
	}

	return MarshalResponse(resp)
}

// HandleLeverageParams handles the get the x/leverage module's parameters.
func (q UmeeQuery) HandleLeverageParams(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.Params(sdk.WrapSDKContext(ctx), &lvtypes.QueryParamsRequest{})
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Leverage Params", err)}
	}

	return MarshalResponse(resp)
}

// HandleBorrowedValue handles the borrowed value of an specific denom and address.
func (q UmeeQuery) HandleBorrowedValue(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.BorrowedValue(sdk.WrapSDKContext(ctx), q.BorrowedValue)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Borrowed Value", err)}
	}

	return MarshalResponse(resp)
}

// HandleSupplied handles the Supplied amount of an address.
func (q UmeeQuery) HandleSupplied(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.Supplied(sdk.WrapSDKContext(ctx), q.Supplied)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Supplied", err)}
	}

	return MarshalResponse(resp)
}

// HandleSuppliedValue handles the Supplied amount of an address in USD.
func (q UmeeQuery) HandleSuppliedValue(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.SuppliedValue(sdk.WrapSDKContext(ctx), q.SuppliedValue)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Supplied Value", err)}
	}

	return MarshalResponse(resp)
}

// HandleAvailableBorrow retrieves the available borrow amount of an denom.
func (q UmeeQuery) HandleAvailableBorrow(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.AvailableBorrow(sdk.WrapSDKContext(ctx), q.AvailableBorrow)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Available Borrow", err)}
	}

	return MarshalResponse(resp)
}

// HandleBorrowAPY retrieves the current borrow interest rate on a token denom.
func (q UmeeQuery) HandleBorrowAPY(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.BorrowAPY(sdk.WrapSDKContext(ctx), q.BorrowAPY)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Borrow APY", err)}
	}

	return MarshalResponse(resp)
}

// HandleSupplyAPY derives the current Supply interest rate on a token denom.
func (q UmeeQuery) HandleSupplyAPY(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.SupplyAPY(sdk.WrapSDKContext(ctx), q.SupplyAPY)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Supply APY", err)}
	}

	return MarshalResponse(resp)
}

// HandleMarketSize get the market size in USD of a token denom.
func (q UmeeQuery) HandleMarketSize(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.MarketSize(sdk.WrapSDKContext(ctx), q.MarketSize)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Market Size", err)}
	}

	return MarshalResponse(resp)
}

// HandleTokenMarketSize handles the market size of an token.
func (q UmeeQuery) HandleTokenMarketSize(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.TokenMarketSize(sdk.WrapSDKContext(ctx), q.TokenMarketSize)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Token Market Size", err)}
	}

	return MarshalResponse(resp)
}

// HandleReserveAmount gets the amount reserved of a specified token.
func (q UmeeQuery) HandleReserveAmount(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.ReserveAmount(sdk.WrapSDKContext(ctx), q.ReserveAmount)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Reserve Amount Size", err)}
	}

	return MarshalResponse(resp)
}

// HandleCollateral gets the collateral amount of a user by token denomination.
// If the denomination is not supplied, all of the user's collateral tokens
// are returned.
func (q UmeeQuery) HandleCollateral(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.Collateral(sdk.WrapSDKContext(ctx), q.Collateral)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Collateral", err)}
	}

	return MarshalResponse(resp)
}

// HandleCollateralValue gets the total USD value of a user's collateral, or
// the USD value held as a given base asset's associated uToken denomination.
func (q UmeeQuery) HandleCollateralValue(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.CollateralValue(sdk.WrapSDKContext(ctx), q.CollateralValue)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Collateral Value", err)}
	}

	return MarshalResponse(resp)
}

// HandleExchangeRate gets calculated the token:uToken exchange
// rate of a base token denom.
func (q UmeeQuery) HandleExchangeRate(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.ExchangeRate(sdk.WrapSDKContext(ctx), q.ExchangeRate)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Exchange Rate", err)}
	}

	return MarshalResponse(resp)
}

// HandleBorrowLimit uses the price oracle to determine the borrow limit
// (in USD) provided by collateral sdk.Coins.
func (q UmeeQuery) HandleBorrowLimit(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.BorrowLimit(sdk.WrapSDKContext(ctx), q.BorrowLimit)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Borrow Limit", err)}
	}

	return MarshalResponse(resp)
}

// HandleLiquidationThreshold determines the maximum borrowed value (in USD) that a borrower with given
// collateral could reach before being eligible for liquidation.
func (q UmeeQuery) HandleLiquidationThreshold(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.LiquidationThreshold(sdk.WrapSDKContext(ctx), q.LiquidationThreshold)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{
			Kind: fmt.Sprintf("error %+v to assigned query Liquidation Threshold", err),
		}
	}

	return MarshalResponse(resp)
}

// HandleLiquidationTargets determines an list of borrower addresses eligible for liquidation.
func (q UmeeQuery) HandleLiquidationTargets(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.LiquidationTargets(sdk.WrapSDKContext(ctx), q.LiquidationTargets)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Liquidation Targets", err)}
	}

	return MarshalResponse(resp)
}

// HandleMarketSummary gets the summary data of an denom.
func (q UmeeQuery) HandleMarketSummary(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.MarketSummary(sdk.WrapSDKContext(ctx), q.MarketSummary)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Market Summary", err)}
	}

	return MarshalResponse(resp)
}

// HandleTotalCollateral gets the total collateral system-wide of a given
// uToken denomination.
func (q UmeeQuery) HandleTotalCollateral(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.TotalCollateral(sdk.WrapSDKContext(ctx), q.TotalCollateral)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Total Collateral", err)}
	}

	return MarshalResponse(resp)
}

// HandleTotalBorrowed gets the total borrowed system-wide of a given
// token denomination.
func (q UmeeQuery) HandleTotalBorrowed(
	ctx sdk.Context,
	qs lvtypes.QueryServer,
) ([]byte, error) {
	resp, err := qs.TotalBorrowed(sdk.WrapSDKContext(ctx), q.TotalBorrowed)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Total Borrowed", err)}
	}

	return MarshalResponse(resp)
}
