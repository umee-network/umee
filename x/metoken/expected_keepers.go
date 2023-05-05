package metoken

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankKeeper interface {
	//todo: add used functions
}

type LeverageKeeper interface {
	//todo: add used functions
}

// Oracle interface for price feed.
type Oracle interface {
	Price(ctx sdk.Context, denom string) (sdk.Dec, error)
}
