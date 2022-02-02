package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/x/leverage/types"
)

// iterate through all keys with a given prefix using a provided function.
// If the provided function returns an error, iteration stops and the error
// is returned.
func (k Keeper) iterate(ctx sdk.Context, prefix []byte, cb func(key, val []byte) error) error {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, val := iter.Key(), iter.Value()

		if err := cb(key, val); err != nil {
			return err
		}
	}

	return nil
}

// GetAllBadDebts gets bad debt instances across all borrowers.
func (k Keeper) GetAllBadDebts(ctx sdk.Context) []types.BadDebt {
	prefix := types.KeyPrefixBadDebt
	badDebts := []types.BadDebt{}

	iterator := func(key, val []byte) error {
		addr := types.AddressFromKey(key, prefix)
		denom := types.DenomFromKeyWithAddress(key, prefix)

		badDebts = append(badDebts, types.NewBadDebt(addr.String(), denom))

		return nil
	}

	err := k.iterate(ctx, prefix, iterator)
	if err != nil {
		panic(err)
	}

	return badDebts
}

// GetAllRegisteredTokens returns all the registered tokens from the x/leverage
// module's KVStore.
func (k Keeper) GetAllRegisteredTokens(ctx sdk.Context) []types.Token {
	tokens := []types.Token{}

	iterator := func(key, val []byte) error {
		var t types.Token
		if err := t.Unmarshal(val); err != nil {
			// improperly marshaled Token should never happen
			return err
		}

		tokens = append(tokens, t)
		return nil
	}

	if err := k.iterate(ctx, types.KeyPrefixRegisteredToken, iterator); err != nil {
		panic(err)
	}

	return tokens
}

// GetAllReserves returns all reserves.
func (k Keeper) GetAllReserves(ctx sdk.Context) sdk.Coins {
	prefix := types.KeyPrefixReserveAmount
	reserves := sdk.NewCoins()

	iterator := func(key, val []byte) error {
		denom := types.DenomFromKey(key, prefix)

		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled reserve amount should never happen
			return err
		}

		reserves = reserves.Add(sdk.NewCoin(denom, amount))
		return nil
	}

	err := k.iterate(ctx, prefix, iterator)
	if err != nil {
		panic(err)
	}

	return reserves
}

// GetBorrowerBorrows returns an sdk.Coins object containing all open borrows
// associated with an address.
func (k Keeper) GetBorrowerBorrows(ctx sdk.Context, borrowerAddr sdk.AccAddress) sdk.Coins {
	prefix := types.CreateAdjustedBorrowKeyNoDenom(borrowerAddr)
	totalBorrowed := sdk.NewCoins()

	iterator := func(key, val []byte) error {
		// get borrow denom from key
		denom := types.DenomFromKeyWithAddress(key, types.KeyPrefixAdjustedBorrow)

		// get borrowed amount
		var adjustedAmount sdk.Dec
		if err := adjustedAmount.Unmarshal(val); err != nil {
			// improperly marshaled borrow amount should never happen
			panic(err)
		}

		// apply interest scalar
		amount := adjustedAmount.Mul(k.getInterestScalar(ctx, denom)).TruncateInt()

		// add to totalBorrowed
		totalBorrowed = totalBorrowed.Add(sdk.NewCoin(denom, amount))
		return nil
	}

	_ = k.iterate(ctx, prefix, iterator)

	return totalBorrowed
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

	return totalCollateral
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

// SweepBadDebts attempts to repay all bad debts in the system.
func (k Keeper) SweepBadDebts(ctx sdk.Context) error {
	prefix := types.KeyPrefixBadDebt

	iterator := func(key, value []byte) error {
		addr := types.AddressFromKey(key, prefix)
		denom := types.DenomFromKeyWithAddress(key, prefix)

		// attempt to repay a single address's debt in this denom
		fullyRepaid, err := k.RepayBadDebt(ctx, addr, denom)
		if err != nil {
			return err
		}
		if fullyRepaid {
			// If the bad debt of this denom at the given address was fully repaid,
			// clear the address|denom pair from the bad debt address list
			k.SetBadDebtAddress(ctx, addr, denom, false)
		}
		return nil
	}

	return k.iterate(ctx, prefix, iterator)
}
