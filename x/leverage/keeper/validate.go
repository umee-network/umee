package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v3/x/leverage/types"
)

// validateAcceptedDenom validates an sdk.Coin and ensures it is a registered Token
// with Blacklisted == false
func (k Keeper) validateAcceptedDenom(ctx sdk.Context, denom string) error {
	if types.HasUTokenPrefix(denom) {
		return types.ErrUToken.Wrap(denom)
	}
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return err
	}
	return token.AssertNotBlacklisted()
}

// validateAcceptedUTokenDenom validates an sdk.Coin and ensures it is a uToken
// associated with a registered Token with Blacklisted == false
func (k Keeper) validateAcceptedUTokenDenom(ctx sdk.Context, udenom string) error {
	if !types.HasUTokenPrefix(udenom) {
		return types.ErrNotUToken.Wrap(udenom)
	}
	return k.validateAcceptedDenom(ctx, types.ToTokenDenom(udenom))
}

// validateAcceptedAsset validates an sdk.Coin and ensures it is a registered Token
// with Blacklisted == false
func (k Keeper) validateAcceptedAsset(ctx sdk.Context, coin sdk.Coin) error {
	if err := coin.Validate(); err != nil {
		return err
	}
	return k.validateAcceptedDenom(ctx, coin.Denom)
}

// validateAcceptedUToken validates an sdk.Coin and ensures it is a uToken
// associated with a registered Token with Blacklisted == false
func (k Keeper) validateAcceptedUToken(ctx sdk.Context, coin sdk.Coin) error {
	if err := coin.Validate(); err != nil {
		return err
	}
	return k.validateAcceptedUTokenDenom(ctx, coin.Denom)
}

// validateSupply validates an sdk.Coin and ensures its Denom is a Token with EnableMsgSupply
func (k Keeper) validateSupply(ctx sdk.Context, coin sdk.Coin) error {
	if err := coin.Validate(); err != nil {
		return err
	}
	if types.HasUTokenPrefix(coin.Denom) {
		return types.ErrUToken.Wrap(coin.Denom)
	}
	token, err := k.GetTokenSettings(ctx, coin.Denom)
	if err != nil {
		return err
	}
	return token.AssertSupplyEnabled()
}

// validateUToken validates an sdk.Coin and ensures its Denom is a uToken. Used by Withdraw and Decollateralize.
func (k Keeper) validateUToken(ctx sdk.Context, coin sdk.Coin) error {
	if err := coin.Validate(); err != nil {
		return err
	}
	if !types.HasUTokenPrefix(coin.Denom) {
		return types.ErrNotUToken.Wrap(coin.Denom)
	}
	return nil
}

// validateBorrow validates an sdk.Coin and ensures its Denom is a Token with EnableMsgBorrow
func (k Keeper) validateBorrow(ctx sdk.Context, borrow sdk.Coin) error {
	if err := borrow.Validate(); err != nil {
		return err
	}
	if types.HasUTokenPrefix(borrow.Denom) {
		return types.ErrUToken.Wrap(borrow.Denom)
	}
	token, err := k.GetTokenSettings(ctx, borrow.Denom)
	if err != nil {
		return err
	}
	return token.AssertBorrowEnabled()
}

// validateRepay validates an sdk.Coin and ensures its Denom is not a uToken
func (k Keeper) validateRepay(ctx sdk.Context, coin sdk.Coin) error {
	if err := coin.Validate(); err != nil {
		return err
	}
	if types.HasUTokenPrefix(coin.Denom) {
		return types.ErrUToken.Wrap(coin.Denom)
	}
	return nil
}

// validateCollateralize validates an sdk.Coin and ensures it is a uToken of an accepted
// Token with EnableMsgSupply and CollateralWeight > 0
func (k Keeper) validateCollateralize(ctx sdk.Context, collateral sdk.Coin) error {
	if err := collateral.Validate(); err != nil {
		return err
	}
	if !types.HasUTokenPrefix(collateral.Denom) {
		return types.ErrNotUToken.Wrap(collateral.Denom)
	}
	token, err := k.GetTokenSettings(ctx, types.ToTokenDenom(collateral.Denom))
	if err != nil {
		return err
	}
	if token.CollateralWeight.IsZero() {
		return types.ErrCollateralWeightZero
	}
	return token.AssertSupplyEnabled()
}
