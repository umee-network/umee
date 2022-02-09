package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// GetCollateralAmount returns an sdk.Coin representing how much of a given denom the
// x/leverage module account currently holds as collateral for a given borrower.
func (k Keeper) GetCollateralAmount(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	collateral := sdk.NewCoin(denom, sdk.ZeroInt())
	key := types.CreateCollateralAmountKey(borrowerAddr, denom)

	if bz := store.Get(key); bz != nil {
		err := collateral.Amount.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
	}

	return collateral
}

// setCollateralAmount sets the amount of a given denom the x/leverage module account
// currently holds as collateral for a given borrower. If the amount is zero, any
// stored value is cleared. A negative amount or invalid coin causes an error.
// This function does not move coins to or from the module account.
func (k Keeper) setCollateralAmount(ctx sdk.Context, borrowerAddr sdk.AccAddress, collateral sdk.Coin) error {
	if !collateral.IsValid() {
		return sdkerrors.Wrap(types.ErrInvalidAsset, collateral.String())
	}

	if borrowerAddr.Empty() {
		return types.ErrEmptyAddress
	}

	bz, err := collateral.Amount.Marshal()
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.CreateCollateralAmountKey(borrowerAddr, collateral.Denom)

	if collateral.Amount.IsZero() {
		store.Delete(key)
	} else {
		store.Set(key, bz)
	}
	return nil
}

// detectCollateralDust returns an error if the input collateral is worth less than 0.1 of a
// token's display denom, if it is nonzero. For example, if 10^6 uatom = 1 atom, this
// function would return an error for u/uatom amounts 0 < (amount * uTokenExchangeRate) < 10^5.
func (k Keeper) detectCollateralDust(ctx sdk.Context, collateral sdk.Coin) error {
	if collateral.Amount.IsPositive() {
		token, err := k.ExchangeUToken(ctx, collateral)
		if err != nil {
			return err
		}
		baseToken, err := k.GetRegisteredToken(ctx, token.Denom)
		if err != nil {
			return err
		}
		if token.Amount.LT(oneTenthFromExponent(baseToken.Exponent)) {
			return sdkerrors.Wrap(types.ErrCollateralDust, collateral.String())
		}
	}
	return nil
}
