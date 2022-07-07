package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// validateAcceptedDenom validates an sdk.Coin and ensures it is a registered Token
// with Blacklisted == false
func (k Keeper) validateAcceptedDenom(ctx sdk.Context, denom string) error {
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
func (k Keeper) validateSupply(ctx sdk.Context, loan sdk.Coin) error {
	if err := loan.Validate(); err != nil {
		return err
	}
	token, err := k.GetTokenSettings(ctx, loan.Denom)
	if err != nil {
		return err
	}
	return token.AssertSupplyEnabled()
}

// validateBorrowAsset validates an sdk.Coin and ensures its Denom is a Token with EnableMsgBorrow
func (k Keeper) validateBorrowAsset(ctx sdk.Context, borrow sdk.Coin) error {
	if err := borrow.Validate(); err != nil {
		return err
	}
	token, err := k.GetTokenSettings(ctx, borrow.Denom)
	if err != nil {
		return err
	}
	return token.AssertBorrowEnabled()
}

// validateCollateralAsset validates an sdk.Coin and ensures its Denom is a Token with EnableMsgSupply
// and CollateralWeight > 0
func (k Keeper) validateCollateralAsset(ctx sdk.Context, collateral sdk.Coin) error {
	if err := collateral.Validate(); err != nil {
		return err
	}
	tokenDenom := k.FromUTokenToTokenDenom(ctx, collateral.Denom)
	token, err := k.GetTokenSettings(ctx, tokenDenom)
	if err != nil {
		return err
	}
	if token.CollateralWeight.IsZero() {
		return types.ErrCollateralWeightZero
	}
	return token.AssertSupplyEnabled()
}
