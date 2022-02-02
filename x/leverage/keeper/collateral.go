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

// SetCollateralAmount sets the amount of a given denom the x/leverage module account
// currently holds as collateral for a given borrower. If the amount is zero, any
// stored value is cleared. A negative amount or invalid coin causes an error.
// This function does not move coins to or from the module account.
func (k Keeper) SetCollateralAmount(ctx sdk.Context, borrowerAddr sdk.AccAddress, collateral sdk.Coin) error {
	if !collateral.IsValid() {
		return sdkerrors.Wrap(types.ErrInvalidAsset, collateral.String())
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

// GetBorrowerCollateral returns an sdk.Coins containing all of a borrower's collateral.
func (k Keeper) GetBorrowerCollateral(ctx sdk.Context, borrowerAddr sdk.AccAddress) sdk.Coins {
	prefix := types.CreateCollateralAmountKeyNoDenom(borrowerAddr)
	totalCollateral := sdk.NewCoins()

	iterator := func(key, val []byte) error {
		// get borrow denom from key
		denom := types.DenomFromKeyWithAddress(key, types.KeyPrefixCollateralAmount)

		// get collateral amount
		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled amount should never happen
			panic(err)
		}

		// add to totalBorrowed
		totalCollateral = totalCollateral.Add(sdk.NewCoin(denom, amount))
		return nil
	}

	_ = k.iterate(ctx, prefix, iterator)

	return totalCollateral.Sort()
}

// GetEligibleLiquidationTargets returns a list of borrower addresses eligible for liquidation.
func (k Keeper) GetEligibleLiquidationTargets(ctx sdk.Context) ([]sdk.AccAddress, error) {
	prefix := types.KeyPrefixAdjustedBorrow
	liquidationTargets := []sdk.AccAddress{}
	checkedAddrs := map[string]struct{}{}

	iterator := func(key, val []byte) error {
		// get borrower address from key
		addr := types.AddressFromKey(key, prefix)

		// if the address is already checked, do not check again
		if _, ok := checkedAddrs[addr.String()]; ok {
			return nil
		}
		checkedAddrs[addr.String()] = struct{}{}

		// get borrower's total borrowed
		borrowed := k.GetBorrowerBorrows(ctx, addr)

		// get borrower's total collateral
		collateral := k.GetBorrowerCollateral(ctx, addr)

		// use oracle helper functions to find total borrowed value in USD
		borrowValue, err := k.TotalTokenValue(ctx, borrowed)
		if err != nil {
			return err
		}

		// use collateral weights to compute borrow limit from enabled collateral
		borrowLimit, err := k.CalculateBorrowLimit(ctx, collateral)
		if err != nil {
			return err
		}

		// if the borrowLimit is smaller then the borrowValue
		// the address is eligible to liquidation
		if borrowLimit.LT(borrowValue) {
			liquidationTargets = append(liquidationTargets, addr)
		}

		return nil
	}

	if err := k.iterate(ctx, prefix, iterator); err != nil {
		return nil, err
	}

	return liquidationTargets, nil
}
