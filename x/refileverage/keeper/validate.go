package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/refileverage/types"
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

// validateCollateralize validates an sdk.Coin and ensures it is a uToken of an accepted
// Token with EnableMsgSupply and CollateralWeight > 0
func (k Keeper) validateCollateralize(ctx sdk.Context, collateral sdk.Coin) error {
	if err := validateUToken(collateral); err != nil {
		return err
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

// validateBaseToken validates an sdk.Coin and ensures its Denom is not a uToken.
func validateBaseToken(coin sdk.Coin) error {
	if err := coin.Validate(); err != nil {
		return err
	}
	if types.HasUTokenPrefix(coin.Denom) {
		return types.ErrUToken.Wrap(coin.Denom)
	}
	return nil
}

// validateUToken validates an sdk.Coin and ensures its Denom is a uToken.
func validateUToken(coin sdk.Coin) error {
	if err := coin.Validate(); err != nil {
		return err
	}
	if !types.HasUTokenPrefix(coin.Denom) {
		return types.ErrNotUToken.Wrap(coin.Denom)
	}
	return nil
}
