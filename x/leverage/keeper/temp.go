package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

/*
  TODO: Remove or rename this file after writing the proper functions to access governance params.

  Update: Use GetRegisteredToken once it exists, then extract the relevant field.

  Except: GetBorrowInterestEpoch should use x/params, not the token registry.
*/

// GetBorrowInterestEpoch gets the borrow interest epoch.
func (k Keeper) GetBorrowInterestEpoch(ctx sdk.Context) (sdk.Int, error) {
	// TODO: Implement this by retrieving the appropriate governance parameter.
	defaultEpoch := sdk.NewInt(100)
	return defaultEpoch, nil
}

// GetReserveFactor gets the reserve factor for a given token.
func (k Keeper) GetReserveFactor(ctx sdk.Context, denom string) (sdk.Dec, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}
	// TODO: Implement this by retrieving the appropriate token-specific governance parameter.
	defaultFactor, _ := sdk.NewDecFromStr("0.25")
	return defaultFactor, nil
}

// GetInterestBase gets the base interest rate for a given token.
func (k Keeper) GetInterestBase(ctx sdk.Context, denom string) (sdk.Dec, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}
	// TODO: Implement this by retrieving the appropriate token-specific governance parameter.
	defaultRate, _ := sdk.NewDecFromStr("0.02")
	return defaultRate, nil
}

// GetInterestMax gets the maximum interest rate for a given token.
func (k Keeper) GetInterestMax(ctx sdk.Context, denom string) (sdk.Dec, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}
	// TODO: Implement this by retrieving the appropriate token-specific governance parameter.
	defaultRate, _ := sdk.NewDecFromStr("1.0")
	return defaultRate, nil
}

// GetInterestAtKink gets the interest rate at the "kink" in the utilization:interest graph for a given token.
func (k Keeper) GetInterestAtKink(ctx sdk.Context, denom string) (sdk.Dec, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}
	// TODO: Implement this by retrieving the appropriate token-specific governance parameter.
	defaultRate, _ := sdk.NewDecFromStr("0.2")
	return defaultRate, nil
}

// GetInterestKinkUtilization gets the utilization at the "kink" in the utilization:interest graph for a given token.
func (k Keeper) GetInterestKinkUtilization(ctx sdk.Context, denom string) (sdk.Dec, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}
	// TODO: Implement this by retrieving the appropriate token-specific governance parameter.
	defaultKink, _ := sdk.NewDecFromStr("0.2")
	return defaultKink, nil
}
