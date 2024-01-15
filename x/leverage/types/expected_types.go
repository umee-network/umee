package types

import (
	context "context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	oracle "github.com/umee-network/umee/v6/x/oracle/types"
)

// AccountKeeper defines the expected account keeper used for leverage simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines the expected x/bank keeper interface.
type BankKeeper interface {
	MintCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
	SendCoinsFromModuleToAccount(
		ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
	) error
	SendCoinsFromAccountToModule(
		ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
	) error
	SendCoinsFromModuleToModule(
		ctx context.Context, senderModule, recipientModule string, amt sdk.Coins,
	) error
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// OracleKeeper defines the expected x/oracle keeper interface.
type OracleKeeper interface {
	GetExchangeRate(ctx sdk.Context, denom string) (oracle.ExchangeRate, error)
	MedianOfHistoricMedians(ctx sdk.Context, denom string, numStamps uint64) (sdkmath.LegacyDec, uint32, error)
}
