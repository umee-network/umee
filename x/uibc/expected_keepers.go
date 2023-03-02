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

type LeverageKeeper interface {
	GetTokenSettings(ctx sdk.Context, baseDenom string) (ltypes.Token, error)
	ExchangeUToken(ctx sdk.Context, uToken sdk.Coin) (sdk.Coin, error)
	DeriveExchangeRate(ctx sdk.Context, denom string) sdk.Dec
}

type OracleKeeper interface {
	HistoricAvgPrice(ctx sdk.Context, denom string) (sdk.Dec, error)
}
