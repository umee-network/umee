package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected x/bank keeper interface.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, fromModule string, toAddr sdk.AccAddress, coins sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, fromAddr sdk.AccAddress, toModule string, coins sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, fromModule string, toModule string, coins sdk.Coins) error
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
