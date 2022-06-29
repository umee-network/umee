package query

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	lvtypes "github.com/umee-network/umee/v2/x/leverage/types"
)

// AssignedQuery defines the query to be called.
type AssignedQuery uint16

const (
	// AssignedQueryBorrowed represents the call to query the Borrowed coins of an address.
	AssignedQueryBorrowed AssignedQuery = iota + 1
	// AssignedQueryGetExchangeRateBase represents the call of oracle get exchange rate.
	AssignedQueryGetExchangeRateBase
	// AssignedQueryRegisteredTokens represents the call of leverage get all registered tokens.
	AssignedQueryRegisteredTokens
	// AssignedQueryLeverageParams represents the call of the x/leverage module's parameters.
	AssignedQueryLeverageParams
	// AssignedQueryBorrowedValue represents the call to query the Borrowed amount of an
	// specific coin of an address.
	AssignedQueryBorrowedValue
	// AssignedQueryLoaned represents the call to query the Loaned amoun of an address.
	AssignedQueryLoaned
	// AssignedQueryLoaned represents the call to query the Loaned amount of an
	// address in USD.
	AssignedQueryLoanedValue
	// AssignedQueryAvailableBorrow represents the call to query the Available
	// amount of an denom.
	AssignedQueryAvailableBorrow
	// AssignedQueryBorrowAPY represents the call to query the current borrow interest
	// rate on a token denom.
	AssignedQueryBorrowAPY
	// AssignedQueryLendAPY represents the call to query and derives the current lend
	// interest rate on a token denom.
	AssignedQueryLendAPY
	// AssignedQueryMarketSize represents the call to query the market size of
	// an token denom in USD.
	AssignedQueryMarketSize
	// AssignedQueryTokenMarketSize represents the call to query the market size of
	// an token denom.
	AssignedQueryTokenMarketSize
	// AssignedQueryReserveAmount represents the call to query the gets the amount
	// reserved of a specified token.
	AssignedQueryReserveAmount
	// AssignedQueryExchangeRate represents the call to query and calculate the
	// token:uToken exchange rate of a base token denom.
	AssignedQueryExchangeRate
	// AssignedQueryBorrowLimit represents the call to query and calculate the
	// borrow limit (in USD).
	AssignedQueryBorrowLimit
	// AssignedQueryLiquidationThreshold represents the call to query and calculate
	// the maximum borrowed value (in USD) that a borrower with given
	// collateral could reach before being eligible for liquidation.
	AssignedQueryLiquidationThreshold
)

// Handler query handler that an object must implement.
type Handler interface {
	Validate() error
	QueryResponse(ctx sdk.Context, keepers Keepers) (interface{}, error)
}

// Keepers wraps the interface to encapsulate keepers.
type Keepers interface {
	// GetExchangeRateBase executes the GetExchangeRateBase from oracle keeper.
	GetExchangeRateBase(ctx sdk.Context, denom string) (sdk.Dec, error)
}

// UmeeQuery wraps all the queries availables for cosmwasm smartcontracts.
type UmeeQuery struct {
	// Mandatory field to determine which query to call.
	AssignedQuery AssignedQuery `json:"assigned_query"`
	// Used to query the Borrowed coins of an address.
	Borrowed *lvtypes.QueryBorrowedRequest `json:"borrowed,omitempty"`
	// Used to query an exchange rate of a denom.
	GetExchangeRateBase *GetExchangeRateBase `json:"get_exchange_rate_base,omitempty"`
	// Used to query all the registered tokens.
	RegisteredTokens *lvtypes.QueryRegisteredTokens `json:"registered_tokens,omitempty"`
	// Used to query the x/leverage module's parameters.
	LeverageParams *lvtypes.QueryParamsRequest `json:"leverage_params,omitempty"`
	// Used to query an specific borrow address value in usd.
	BorrowedValue *lvtypes.QueryBorrowedValueRequest `json:"borrowed_value,omitempty"`
	// Used to query an the amount loaned of an address.
	Loaned *lvtypes.QueryLoanedRequest `json:"loaned,omitempty"`
	// Used to query an the amount loaned of an address in USD.
	LoanedValue *lvtypes.QueryLoanedValueRequest `json:"loaned_value,omitempty"`
	// Used to query an the amount available to borrow.
	AvailableBorrow *lvtypes.QueryAvailableBorrowRequest `json:"available_borrow,omitempty"`
	// Used to query an current borrow interest rate on a token denom.
	BorrowAPY *lvtypes.QueryBorrowAPYRequest `json:"borrow_apy,omitempty"`
	// Used to derives the current lend interest rate on a token denom.
	LendAPY *lvtypes.QueryLendAPYRequest `json:"lend_apy,omitempty"`
	// Used to get the market size in USD of a token denom.
	MarketSize *lvtypes.QueryMarketSizeRequest `json:"market_size,omitempty"`
	// Used to get the market size of a token denom.
	TokenMarketSize *lvtypes.QueryTokenMarketSizeRequest `json:"token_market_size,omitempty"`
	// Used to gets the amount reserved of a specified token.
	ReserveAmount *lvtypes.QueryReserveAmountRequest `json:"reserve_amount,omitempty"`
	// Used to calculate the token:uToken exchange rate of a base token denom.
	ExchangeRate *lvtypes.QueryExchangeRateRequest `json:"exchange_rate,omitempty"`
	// Uses the price oracle to determine the borrow limit (in USD).
	BorrowLimit *lvtypes.QueryBorrowLimitRequest `json:"borrow_limit,omitempty"`
	// determines the maximum borrowed value (in USD) that a borrower with given
	// collateral could reach before being eligible for liquidation.
	LiquidationThreshold *lvtypes.QueryLiquidationThresholdRequest `json:"liquidation_threshold,omitempty"`
}
