package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// getAdjustedBorrow gets the adjusted amount borrowed by an address in a given denom. Returns zero
// if no value is found.
func (k Keeper) getAdjustedBorrow(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Dec {
	key := types.CreateAdjustedBorrowKey(addr, denom)
	adjustedAmount := sdk.ZeroDec()

	if bz := ctx.KVStore(k.storeKey).Get(key); bz != nil {
		if err := adjustedAmount.Unmarshal(bz); err == nil {
			return adjustedAmount
		}
	}

	return sdk.ZeroDec()
}

// getAdjustedTotalBorrowed gets the total borrowed amount for a given denom. Returns zero if
// no value is stored.
func (k Keeper) getAdjustedTotalBorrowed(ctx sdk.Context, denom string) sdk.Dec {
	key := types.CreateAdjustedTotalBorrowKey(denom)
	adjustedAmount := sdk.ZeroDec()

	if bz := ctx.KVStore(k.storeKey).Get(key); bz != nil {
		if err := adjustedAmount.Unmarshal(bz); err == nil {
			return adjustedAmount
		}
	}

	return sdk.ZeroDec()
}

// setAdjustedBorrow sets the adjusted amount borrowed by an address in a given denom directly instead
// of computing it from real borrowed amount. Should only be used by genesis and SetBorrow, as other
// functions deal in non-adjusted amounts using SetBorrow. Also updates AdjustedTotalBorrowed by the
// resulting changes in borrowed amount. If either amount to store is zero, any stored value is cleared.
func (k Keeper) setAdjustedBorrow(ctx sdk.Context, addr sdk.AccAddress, borrow sdk.DecCoin) error {
	if err := borrow.Validate(); err != nil {
		return err
	}

	// Update total adjusted borrow based on the change in this borrow's adjusted amount
	delta := borrow.Amount.Sub(k.getAdjustedBorrow(ctx, addr, borrow.Denom))
	total := sdk.NewDecCoinFromDec(borrow.Denom, k.getAdjustedTotalBorrowed(ctx, borrow.Denom).Add(delta))

	// Set new borrow value, or clear if zero
	store := ctx.KVStore(k.storeKey)
	key := types.CreateAdjustedBorrowKey(addr, borrow.Denom)

	if borrow.Amount.IsZero() {
		store.Delete(key)
	} else {
		bz, err := borrow.Amount.Marshal()
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}

	// Also set total borrowed value, or clear if zero
	key = types.CreateAdjustedTotalBorrowKey(total.Denom)
	if !total.Amount.IsPositive() {
		store.Delete(key)
	} else {
		bz, err := total.Amount.Marshal()
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}
	return nil
}

// setCollateralSetting enables or disables a borrower's collateral setting for a single denom
func (k Keeper) setCollateralSetting(ctx sdk.Context, addr sdk.AccAddress, denom string, enable bool) error {
	if err := sdk.ValidateDenom(denom); err != nil {
		return err
	}

	// Enable sets to true; disable removes from KVstore rather than setting false
	store := ctx.KVStore(k.storeKey)
	key := types.CreateCollateralSettingKey(addr, denom)
	if enable {
		store.Set(key, []byte{0x01})
	} else {
		store.Delete(key)
	}
	return nil
}

// getInterestScalar gets the interest scalar for a given base token
// denom. Returns 1.0 if no value is stored.
func (k Keeper) getInterestScalar(ctx sdk.Context, denom string) sdk.Dec {
	key := types.CreateInterestScalarKey(denom)
	scalar := sdk.OneDec()

	if bz := ctx.KVStore(k.storeKey).Get(key); bz != nil {
		if err := scalar.Unmarshal(bz); err != nil {
			panic(err)
		}
	}

	return scalar
}

// setInterestScalar sets the interest scalar for a given base token denom.
func (k Keeper) setInterestScalar(ctx sdk.Context, denom string, scalar sdk.Dec) error {
	if err := sdk.ValidateDenom(denom); err != nil {
		return err
	}
	if scalar.LT(sdk.OneDec()) {
		return sdkerrors.Wrap(types.ErrInvalidInteresrScalar, scalar.String()+denom)
	}

	bz, err := scalar.Marshal()
	if err != nil {
		return err
	}

	key := types.CreateInterestScalarKey(denom)
	ctx.KVStore(k.storeKey).Set(key, bz)
	return nil
}
