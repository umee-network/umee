package incentive

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// BankKeeper defines the expected x/bank keeper interface.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, fromModule string, toAddr sdk.AccAddress, coins sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, fromAddr sdk.AccAddress, toModule string, coins sdk.Coins) error
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// LeverageKeeper defines the expected x/leverage keeper interface.
type LeverageKeeper interface {
	GetCollateral(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) sdk.Coin
	DonateCollateral(ctx sdk.Context, fromAddr sdk.AccAddress, uToken sdk.Coin) error
	GetTokenSettings(ctx sdk.Context, denom string) (leveragetypes.Token, error)
}
