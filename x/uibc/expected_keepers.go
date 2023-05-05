package uibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"

	ltypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// BankKeeper defines the expected x/bank keeper interface.
type BankKeeper interface {
	GetDenomMetaData(ctx sdk.Context, denom string) (types.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData types.Metadata)
	IterateAllDenomMetaData(ctx sdk.Context, cb func(types.Metadata) bool)
}

type Leverage interface {
	GetTokenSettings(ctx sdk.Context, baseDenom string) (ltypes.Token, error)
	ExchangeUToken(ctx sdk.Context, uToken sdk.Coin) (sdk.Coin, error)
	DeriveExchangeRate(ctx sdk.Context, denom string) sdk.Dec
}

// Oracle interface for price feed.
// The uibc design doesn't depend on any particular price metric (spot price, avg ...), so it's
// up to the integration which price should be used.
type Oracle interface {
	Price(ctx sdk.Context, denom string) (sdk.Dec, error)
}
