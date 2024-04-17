package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected x/bank keeper interface.
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
