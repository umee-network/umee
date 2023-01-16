package uibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"

	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// BankKeeper defines the expected x/bank keeper interface.
type BankKeeper interface {
	GetDenomMetaData(ctx sdk.Context, denom string) (types.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData types.Metadata)
	IterateAllDenomMetaData(ctx sdk.Context, cb func(types.Metadata) bool)
}

type LeverageKeeper interface {
	GetTokenSettings(ctx sdk.Context, baseDenom string) (leveragetypes.Token, error)
	TokenValue(ctx sdk.Context, coin sdk.Coin, historic bool) (sdk.Dec, error)
	ExchangeUToken(ctx sdk.Context, uToken sdk.Coin) (sdk.Coin, error)
}
