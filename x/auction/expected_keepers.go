package auction

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected x/bank keeper interface.
type BankKeeper interface {
	BurnCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string,
		amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress,
		amt sdk.Coins) error
}
