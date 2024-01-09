package metoken

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	otypes "github.com/umee-network/umee/v6/x/oracle/types"
)

// BankKeeper defines the expected x/bank keeper interface.
type BankKeeper interface {
	MintCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	SendCoinsFromModuleToAccount(
		ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
	) error
	SendCoinsFromAccountToModule(
		ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
	) error
}

// LeverageKeeper interface for interacting with x/leverage
type LeverageKeeper interface {
	GetTokenSettings(ctx sdk.Context, denom string) (ltypes.Token, error)
	ToUToken(ctx sdk.Context, token sdk.Coin) (sdk.Coin, error)
	ToToken(ctx sdk.Context, uToken sdk.Coin) (sdk.Coin, error)
	SupplyFromModule(ctx sdk.Context, fromModule string, coin sdk.Coin) (sdk.Coin, bool, error)
	WithdrawToModule(ctx sdk.Context, toModule string, uToken sdk.Coin) (sdk.Coin, bool, error)
	ModuleMaxWithdraw(ctx sdk.Context, spendableUTokens sdk.Coin, withdrawalAddr sdk.AccAddress) (sdkmath.Int, error)
	GetTotalSupply(ctx sdk.Context, denom string) (sdk.Coin, error)
	GetAllSupplied(ctx sdk.Context, supplierAddr sdk.AccAddress) (sdk.Coins, error)
}

// OracleKeeper interface for price feed.
type OracleKeeper interface {
	AllMedianPrices(ctx sdk.Context) otypes.Prices
	SetExchangeRate(ctx sdk.Context, denom string, rate sdk.Dec)
}
