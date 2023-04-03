package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account keeper used for leverage simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

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
	SendCoinsFromModuleToModule(
		ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins,
	) error
	SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// OracleKeeper defines the expected x/oracle keeper interface.
type OracleKeeper interface {
	GetExchangeRate(ctx sdk.Context, denom string) (sdk.Dec, error)
	MedianOfHistoricMedians(ctx sdk.Context, denom string, numStamps uint64) (sdk.Dec, uint32, error)
}

// IncentiveKeeper defines the expected x/incentive keeper interface.
type IncentiveKeeper interface {
	// Used to ensure bonded or unbonding collateral cannot be decollateralized or withdrawn.
	RestrictedCollateral(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	// Used when liquidating an account, and collateral must be reduced by unbonding instantly if necessary
	ForceSetCollateral(ctx sdk.Context, addr sdk.AccAddress, uToken sdk.Coin) error
}
