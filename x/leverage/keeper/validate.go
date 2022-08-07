package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// ValidateAcceptedDenom validates an sdk.Coin and ensures it is a registered Token
// with Blacklisted == false
func (k Keeper) ValidateAcceptedDenom(ctx sdk.Context, denom string) error {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return err
	}
	return token.AssertNotBlacklisted()
}

// ValidateAcceptedUTokenDenom validates an sdk.Coin and ensures it is a uToken
// associated with a registered Token with Blacklisted == false
func (k Keeper) ValidateAcceptedUTokenDenom(ctx sdk.Context, udenom string) error {
	if !types.HasUTokenPrefix(udenom) {
		return types.ErrNotUToken.Wrap(udenom)
	}
	return k.ValidateAcceptedDenom(ctx, types.ToTokenDenom(udenom))
}

// ValidateAcceptedAsset validates an sdk.Coin and ensures it is a registered Token
// with Blacklisted == false
func (k Keeper) ValidateAcceptedAsset(ctx sdk.Context, coin sdk.Coin) error {
	if err := coin.Validate(); err != nil {
		return err
	}
	return k.ValidateAcceptedDenom(ctx, coin.Denom)
}

// ValidateAcceptedUToken validates an sdk.Coin and ensures it is a uToken
// associated with a registered Token with Blacklisted == false
func (k Keeper) ValidateAcceptedUToken(ctx sdk.Context, coin sdk.Coin) error {
	if err := coin.Validate(); err != nil {
		return err
	}
	return k.ValidateAcceptedUTokenDenom(ctx, coin.Denom)
}

// ValidateSupply validates an sdk.Coin and ensures its Denom is a Token with EnableMsgSupply
func (k Keeper) ValidateSupply(ctx sdk.Context, loan sdk.Coin) error {
	if err := loan.Validate(); err != nil {
		return err
	}
	token, err := k.GetTokenSettings(ctx, loan.Denom)
	if err != nil {
		return err
	}
	return token.AssertSupplyEnabled()
}

// ValidateBorrow validates an sdk.Coin and ensures its Denom is a Token with EnableMsgBorrow
func (k Keeper) ValidateBorrow(ctx sdk.Context, borrow sdk.Coin) error {
	if err := borrow.Validate(); err != nil {
		return err
	}
	token, err := k.GetTokenSettings(ctx, borrow.Denom)
	if err != nil {
		return err
	}
	return token.AssertBorrowEnabled()
}

// ValidateCollateralize validates an sdk.Coin and ensures it is a uToken of an accepted
// Token with EnableMsgSupply and CollateralWeight > 0
func (k Keeper) ValidateCollateralize(ctx sdk.Context, collateral sdk.Coin) error {
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
