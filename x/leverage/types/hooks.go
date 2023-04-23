package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TokenHooks defines hooks other modules can execute when the leverage module
// adds or removes a token.
type TokenHooks interface {
	// AfterTokenRegistered defines a hook any keeper can execute after the
	// x/leverage registers a token.
	AfterTokenRegistered(ctx sdk.Context, token Token)

	// AfterRegisteredTokenRemoved defines a hook any keeper can execute after
	// the x/leverage module deletes a registered token.
	AfterRegisteredTokenRemoved(ctx sdk.Context, token Token)
}

// BondHooks defines hooks leverage module can call on other modules to determine how much
// of a user's uToken collateral is bonded (i.e. not allowed to be withrdawn) or to force
// this amount to be reduced in the event of a liquidation.
type BondHooks interface {
	// Used to ensure bonded or unbonding collateral cannot be decollateralized or withdrawn.
	GetBonded(ctx sdk.Context, addr sdk.AccAddress, uDenom string) sdkmath.Int

	// Used when liquidating an account, and collateral must be unbonded instantly until bonded amount
	// is no greater than the account's remaining collateral uTokens.
	ForceUnondTo(ctx sdk.Context, addr sdk.AccAddress, uToken sdk.Coin) error
}
