package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/umee-network/umee/v3/x/leverage/types"
)

// getStoredDec retrieves an sdk.Dec from the KVStore, or zero if no value is stored.
// It panics if a stored value fails to unmarshal into a value higher than a specified minimum.
// Accepts an additional string which should describe the field being retrieved in panic messages.
func (k Keeper) getStoredDec(ctx sdk.Context, key []byte, minimum sdk.Dec, desc string) sdk.Dec {
	if bz := ctx.KVStore(k.storeKey).Get(key); bz != nil {
		val := sdk.ZeroDec()
		if err := val.Unmarshal(bz); err != nil {
			panic(err)
		}
		if val.LTE(minimum) {
			panic(types.ErrGetAmount.Wrapf("%s is not above the minimum %s of %s", val, desc, minimum))
		}
		return val
	}
	// No stored bytes at key
	return minimum
}

// setStoredDec stores an sdk.Dec in the KVStore, or clears if setting to zero.
// Returns an error on attempting to store value lower than a specified minimum or on failure to encode.
// Accepts an additional string which should describe the field being set in custom error messages.
func (k Keeper) setStoredDec(ctx sdk.Context, key []byte, val, minimum sdk.Dec, desc string) error {
	store := ctx.KVStore(k.storeKey)
	if val.LT(minimum) {
		return types.ErrSetAmount.Wrapf("%s is below the minimum %s of %s", val, desc, minimum)
	}
	if val.Equal(minimum) {
		store.Delete(key)
	} else {
		bz, err := val.Marshal()
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}
	return nil
}

// getStoredInt retrieves an sdkmath.Int from the KVStore, or zero if no value is stored.
// It panics if a stored value fails to unmarshal or is not positive.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func (k Keeper) getStoredInt(ctx sdk.Context, key []byte, desc string) sdkmath.Int {
	if bz := ctx.KVStore(k.storeKey).Get(key); bz != nil {
		val := sdk.ZeroInt()
		if err := val.Unmarshal(bz); err != nil {
			panic(err)
		}
		if val.LTE(sdk.ZeroInt()) {
			panic(types.ErrGetAmount.Wrapf("%s is not above the minimum %s of zero", val, desc))
		}
		return val
	}
	// No stored bytes at key
	return sdk.ZeroInt()
}

// setStoredInt stores an sdkmath.Int in the KVStore, or clears if setting to zero.
// Returns an error on attempting to store negative value or on failure to encode.
// Accepts an additional string which should describe the field being set in custom error messages.
func (k Keeper) setStoredInt(ctx sdk.Context, key []byte, val sdkmath.Int, desc string) error {
	store := ctx.KVStore(k.storeKey)
	if val.IsNegative() {
		return types.ErrSetAmount.Wrapf("%s is below the minimum %s of zero", val, desc)
	}
	if val.IsZero() {
		store.Delete(key)
	} else {
		bz, err := val.Marshal()
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}
	return nil
}

// getAdjustedBorrow gets the adjusted amount borrowed by an address in a given denom.
// Returned value is non-negative.
func (k Keeper) getAdjustedBorrow(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Dec {
	key := types.CreateAdjustedBorrowKey(addr, denom)
	return k.getStoredDec(ctx, key, sdk.ZeroDec(), "adjusted borrow")
}

// getAdjustedTotalBorrowed gets the total amount borrowed across all borrowers for a given denom.
// Returned value is non-negative.
func (k Keeper) getAdjustedTotalBorrowed(ctx sdk.Context, denom string) sdk.Dec {
	key := types.CreateAdjustedTotalBorrowKey(denom)
	return k.getStoredDec(ctx, key, sdk.ZeroDec(), "adjusted total borrow")
}

// setAdjustedBorrow sets the adjusted amount borrowed by an address in a given denom directly instead
// of computing it from real borrowed amount. Should only be used by genesis and setBorrow, as other
// functions set actual borrowed amount using setBorrow. Also updates AdjustedTotalBorrowed by the
// resulting change in borrowed amount. Value must always be non-negative.
func (k Keeper) setAdjustedBorrow(ctx sdk.Context, addr sdk.AccAddress, adjustedBorrow sdk.DecCoin) error {
	if err := validateBaseDenom(adjustedBorrow.Denom); err != nil {
		return err
	}
	if addr.Empty() {
		return types.ErrEmptyAddress
	}

	// Determine the increase or decrease in total borrowed. A decrease is negative.
	delta := adjustedBorrow.Amount.Sub(k.getAdjustedBorrow(ctx, addr, adjustedBorrow.Denom))

	// Update total adjusted borrow
	key := types.CreateAdjustedTotalBorrowKey(adjustedBorrow.Denom)
	newTotal := k.getAdjustedTotalBorrowed(ctx, adjustedBorrow.Denom).Add(delta)
	err := k.setStoredDec(ctx, key, newTotal, sdk.ZeroDec(), "total adjusted borrow")
	if err != nil {
		return err
	}

	// Set new adjusted borrow
	key = types.CreateAdjustedBorrowKey(addr, adjustedBorrow.Denom)
	return k.setStoredDec(ctx, key, adjustedBorrow.Amount, sdk.ZeroDec(), "adjusted borrow")
}

// GetCollateral returns an sdk.Coin representing how much of a given denom the
// x/leverage module account currently holds as collateral for a given borrower.
func (k Keeper) GetCollateral(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) sdk.Coin {
	key := types.CreateCollateralAmountKey(borrowerAddr, denom)
	amount := k.getStoredInt(ctx, key, "collateral")
	return sdk.NewCoin(denom, amount)
}

// setCollateral sets the amount of a given denom the x/leverage module account
// currently holds as collateral for a given borrower. Collateral must be a uToken.
// This function does not move coins to or from the module account.
func (k Keeper) setCollateral(ctx sdk.Context, borrowerAddr sdk.AccAddress, collateral sdk.Coin) error {
	if err := validateUToken(collateral); err != nil {
		return err
	}
	if borrowerAddr.Empty() {
		return types.ErrEmptyAddress
	}
	key := types.CreateCollateralAmountKey(borrowerAddr, collateral.Denom)
	return k.setStoredInt(ctx, key, collateral.Amount, "collateral")
}

// GetReserves gets the reserved amount of a specified token.
// On invalid asset, the reserved amount is zero.
func (k Keeper) GetReserves(ctx sdk.Context, denom string) sdk.Coin {
	key := types.CreateReserveAmountKey(denom)
	amount := k.getStoredInt(ctx, key, "reserves")
	return sdk.NewCoin(denom, amount)
}

// setReserves sets the reserved amount of a specified token.
func (k Keeper) setReserves(ctx sdk.Context, reserves sdk.Coin) error {
	if err := validateBaseToken(reserves); err != nil {
		return err
	}

	key := types.CreateReserveAmountKey(reserves.Denom)
	return k.setStoredInt(ctx, key, reserves.Amount, "reserves")
}

// getLastInterestTime returns unix timestamp (in seconds) when the last interest was accrued.
// Returns 0 if the value if the value is absent.
func (k Keeper) getLastInterestTime(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	timeKey := types.CreateLastInterestTimeKey()
	bz := store.Get(timeKey)
	if bz == nil {
		return 0
	}

	val := gogotypes.Int64Value{}
	if err := k.cdc.Unmarshal(bz, &val); err != nil {
		panic(err)
	}
	if val.Value < 0 {
		panic(types.ErrGetAmount.Wrapf("%d is below the minimum LastInterestTime of zero", val.Value))
	}

	return val.Value
}

// setLastInterestTime sets LastInterestTime to a given value
func (k *Keeper) setLastInterestTime(ctx sdk.Context, interestTime int64) error {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateLastInterestTimeKey()

	prevTime := k.getLastInterestTime(ctx)
	if interestTime < prevTime {
		// prevent time from moving backwards
		return types.ErrNegativeTimeElapsed.Wrapf("cannot set LastInterestTime from %d to %d",
			prevTime, interestTime)
	}

	bz, err := k.cdc.Marshal(&gogotypes.Int64Value{Value: interestTime})
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// setBadDebtAddress sets or deletes an address in a denom's list of addresses with unpaid bad debt.
func (k Keeper) setBadDebtAddress(ctx sdk.Context, addr sdk.AccAddress, denom string, hasDebt bool) error {
	if err := validateBaseDenom(denom); err != nil {
		return err
	}
	if addr.Empty() {
		return types.ErrEmptyAddress
	}

	store := ctx.KVStore(k.storeKey)
	key := types.CreateBadDebtKey(denom, addr)

	if hasDebt {
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
	return k.getStoredDec(ctx, key, sdk.OneDec(), "interest scalar")
}

// setInterestScalar sets the interest scalar for a given base token denom.
func (k Keeper) setInterestScalar(ctx sdk.Context, denom string, scalar sdk.Dec) error {
	if err := validateBaseDenom(denom); err != nil {
		return err
	}
	key := types.CreateInterestScalarKey(denom)
	return k.setStoredDec(ctx, key, scalar, sdk.OneDec(), "interest scalar")
}

// GetUTokenSupply gets the total supply of a specified utoken, as tracked by
// module state. On invalid asset or non-uToken, the supply is zero.
func (k Keeper) GetUTokenSupply(ctx sdk.Context, denom string) sdk.Coin {
	key := types.CreateUTokenSupplyKey(denom)
	amount := k.getStoredInt(ctx, key, "uToken supply")
	return sdk.NewCoin(denom, amount)
}

// setUTokenSupply sets the total supply of a uToken.
func (k Keeper) setUTokenSupply(ctx sdk.Context, uToken sdk.Coin) error {
	if err := validateUToken(uToken); err != nil {
		return err
	}
	key := types.CreateUTokenSupplyKey(uToken.Denom)
	return k.setStoredInt(ctx, key, uToken.Amount, "uToken supply")
}
