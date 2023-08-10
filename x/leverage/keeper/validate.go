package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

// validateAcceptedDenom validates an sdk.Coin and ensures it is a registered Token
// with Blacklisted == false
func (k Keeper) validateAcceptedDenom(ctx sdk.Context, denom string) error {
	if coin.HasUTokenPrefix(denom) {
		return types.ErrUToken.Wrap(denom)
	}
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return err
	}
	return token.AssertNotBlacklisted()
}

// validateAcceptedAsset validates an sdk.Coin and ensures it is a registered Token
// with Blacklisted == false
func (k Keeper) validateAcceptedAsset(ctx sdk.Context, coin sdk.Coin) error {
	if err := coin.Validate(); err != nil {
		return err
	}
	return k.validateAcceptedDenom(ctx, coin.Denom)
}

// validateSupply validates an sdk.Coin and ensures its Denom is a Token with EnableMsgSupply
func (k Keeper) validateSupply(ctx sdk.Context, coin sdk.Coin) error {
	if err := validateBaseToken(coin); err != nil {
		return err
	}
	token, err := k.GetTokenSettings(ctx, coin.Denom)
	if err != nil {
		return err
	}
	return token.AssertSupplyEnabled()
}

// validateBorrow validates an sdk.Coin and ensures its Denom is a Token with EnableMsgBorrow
func (k Keeper) validateBorrow(ctx sdk.Context, borrow sdk.Coin) error {
	if err := validateBaseToken(borrow); err != nil {
		return err
	}
	token, err := k.GetTokenSettings(ctx, borrow.Denom)
	if err != nil {
		return err
	}
	return token.AssertBorrowEnabled()
}

// validateCollateralize validates an sdk.Coin and ensures it is a uToken of an accepted
// Token with EnableMsgSupply and CollateralWeight > 0
func (k Keeper) validateCollateralize(ctx sdk.Context, collateral sdk.Coin) error {
	if err := validateUToken(collateral); err != nil {
		return err
	}
	token, err := k.GetTokenSettings(ctx, coin.StripUTokenDenom(collateral.Denom))
	if err != nil {
		return err
	}
	if token.CollateralWeight.IsZero() {
		return types.ErrCollateralWeightZero
	}
	return token.AssertSupplyEnabled()
}

// validateBaseToken validates an sdk.Coin and ensures its Denom is not a uToken.
func validateBaseToken(c sdk.Coin) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if coin.HasUTokenPrefix(c.Denom) {
		return types.ErrUToken.Wrap(c.Denom)
	}
	return nil
}

// validateUToken validates an sdk.Coin and ensures its Denom is a uToken.
func validateUToken(c sdk.Coin) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if !coin.HasUTokenPrefix(c.Denom) {
		return types.ErrNotUToken.Wrap(c.Denom)
	}
	return nil
}
