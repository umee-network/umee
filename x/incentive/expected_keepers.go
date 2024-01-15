package incentive

import (
	context "context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
)

// BankKeeper defines the expected x/bank keeper interface.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, fromModule string, toAddr sdk.AccAddress, coins sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, fromAddr sdk.AccAddress, toModule string, coins sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, fromModule string, toModule string, coins sdk.Coins) error
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// LeverageKeeper defines the expected x/leverage keeper interface.
type LeverageKeeper interface {
	GetCollateral(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) sdk.Coin
	DonateCollateral(ctx sdk.Context, fromAddr sdk.AccAddress, uToken sdk.Coin) error
	GetTokenSettings(ctx sdk.Context, denom string) (leveragetypes.Token, error)
	// These are used for APY queries only
	TotalTokenValue(ctx sdk.Context, coins sdk.Coins, mode leveragetypes.PriceMode) (sdkmath.LegacyDec, error)
	ToToken(ctx sdk.Context, uToken sdk.Coin) (sdk.Coin, error)
}
