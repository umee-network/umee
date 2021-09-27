package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// BankKeeper defines the expected x/bank keeper interface.
type BankKeeper interface {
	GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
	IterateAllDenomMetaData(ctx sdk.Context, cb func(banktypes.Metadata) bool)
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	// Note: The following will panic until we have a module account
	// See auth/types/ModuleAccount
	MintCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	SendCoinsFromModuleToAccount(
		ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
	) error
	SendCoinsFromAccountToModule(
		ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
	) error
	// Question: Do we need the following (commented out)?
	// SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	// BlockedAddr(addr sdk.AccAddress) bool
	HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool
	// Question: When processing a deposit message, is it sufficient to use hasBalance,
	//	or must we iterate through spendableCoins?
	// 	That would be the case, for example, if HasBalance counted ibc-locked coins in its total
	// SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
