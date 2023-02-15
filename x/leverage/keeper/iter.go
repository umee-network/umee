package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util"
	"github.com/umee-network/umee/v4/x/leverage/types"
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

// getAllBadDebts gets bad debt instances across all borrowers.
func (k Keeper) getAllBadDebts(ctx sdk.Context) []types.BadDebt {
	prefix := types.KeyPrefixBadDebt
	badDebts := []types.BadDebt{}

	iterator := func(key, _ []byte) error {
		addr := types.AddressFromKey(key, prefix)
		denom := types.DenomFromKeyWithAddress(key, prefix)
		badDebts = append(badDebts, types.NewBadDebt(addr.String(), denom))

		return nil
	}

	util.Panic(k.iterate(ctx, prefix, iterator))

	return badDebts
}

// GetAllRegisteredTokens returns all the registered tokens from the x/leverage
// module's KVStore.
func (k Keeper) GetAllRegisteredTokens(ctx sdk.Context) []types.Token {
	tokens := []types.Token{}

	iterator := func(_, val []byte) error {
		var t types.Token
		if err := t.Unmarshal(val); err != nil {
			// improperly marshaled Token should never happen
			return err
		}

		tokens = append(tokens, t)
		return nil
	}

	util.Panic(k.iterate(ctx, types.KeyPrefixRegisteredToken, iterator))

	return tokens
}

// GetAllReserves returns all reserves.
func (k Keeper) GetAllReserves(ctx sdk.Context) sdk.Coins {
	prefix := types.KeyPrefixReserveAmount
	reserves := sdk.NewCoins()

	iterator := func(key, val []byte) error {
		denom := types.DenomFromKey(key, prefix)

		var amount sdkmath.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled reserve amount should never happen
			return err
		}

		reserves = reserves.Add(sdk.NewCoin(denom, amount))
		return nil
	}

	util.Panic(k.iterate(ctx, prefix, iterator))

	return reserves
}

// GetBorrowerBorrows returns an sdk.Coins object containing all open borrows
// associated with an address.
func (k Keeper) GetBorrowerBorrows(ctx sdk.Context, borrowerAddr sdk.AccAddress) sdk.Coins {
	prefix := types.KeyAdjustedBorrowNoDenom(borrowerAddr)
	totalBorrowed := sdk.NewCoins()

	iterator := func(key, val []byte) error {
		// get borrow denom from key
		denom := types.DenomFromKeyWithAddress(key, types.KeyPrefixAdjustedBorrow)

		// get borrowed amount
		var adjustedAmount sdk.Dec
		if err := adjustedAmount.Unmarshal(val); err != nil {
			// improperly marshaled borrow amount should never happen
			return err
		}

		// apply interest scalar
		amount := adjustedAmount.Mul(k.getInterestScalar(ctx, denom)).Ceil().TruncateInt()

		// add to totalBorrowed
		totalBorrowed = totalBorrowed.Add(sdk.NewCoin(denom, amount))
		return nil
	}

	util.Panic(k.iterate(ctx, prefix, iterator))

	return totalBorrowed
}

// GetBorrowerCollateral returns an sdk.Coins containing all of a borrower's collateral.
func (k Keeper) GetBorrowerCollateral(ctx sdk.Context, borrowerAddr sdk.AccAddress) sdk.Coins {
	prefix := types.KeyCollateralAmountNoDenom(borrowerAddr)
	totalCollateral := sdk.NewCoins()

	iterator := func(key, val []byte) error {
		// get borrow denom from key
		denom := types.DenomFromKeyWithAddress(key, types.KeyPrefixCollateralAmount)

		// get collateral amount
		var amount sdkmath.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled amount should never happen
			return err
		}

		totalCollateral = totalCollateral.Add(sdk.NewCoin(denom, amount))
		return nil
	}

	util.Panic(k.iterate(ctx, prefix, iterator))

	return totalCollateral
}

// GetEligibleLiquidationTargets returns a list of borrower addresses eligible for liquidation.
func (k Keeper) GetEligibleLiquidationTargets(ctx sdk.Context) ([]sdk.AccAddress, error) {
	prefix := types.KeyPrefixAdjustedBorrow
	liquidationTargets := []sdk.AccAddress{}
	checkedAddrs := map[string]struct{}{}

	iterator := func(key, _ []byte) error {
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
		// skips denoms without prices
		borrowValue, err := k.VisibleTokenValue(ctx, borrowed, types.PriceModeSpot)
		if err != nil {
			return err
		}

		// compute liquidation threshold from enabled collateral
		// in this case, we can't reasonably skip missing prices but can move on
		// to the next borrower instead of stopping the entire query
		liquidationLimit, err := k.CalculateLiquidationThreshold(ctx, collateral)
		if err == nil && liquidationLimit.LT(borrowValue) {
			// If liquidation limit is smaller than borrowed value then the
			// address is eligible for liquidation.
			liquidationTargets = append(liquidationTargets, addr)
		}
		// Non-price errors will cause the query itself to fail
		if nonOracleError(err) {
			return err
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

	iterator := func(key, _ []byte) error {
		addr := types.AddressFromKey(key, prefix)
		denom := types.DenomFromKeyWithAddress(key, prefix)

		// clear blacklisted collateral while checking for any remaining (valid) collateral
		done, err := k.clearBlacklistedCollateral(ctx, addr)
		if err != nil {
			return err
		}

		// if remaining collateral is zero, attempt to repay this bad debt
		if !done {
			var err error
			done, err = k.RepayBadDebt(ctx, addr, denom)
			if err != nil {
				return err
			}
		}

		// if collateral found or debt fully repaid, clear the bad debt entry for this address|denom
		if done {
			if err := k.setBadDebtAddress(ctx, addr, denom, false); err != nil {
				return err
			}
		}
		return nil
	}

	return k.iterate(ctx, prefix, iterator)
}

// GetAllUTokenSupply returns total supply of all uToken denoms.
func (k Keeper) GetAllUTokenSupply(ctx sdk.Context) sdk.Coins {
	prefix := types.KeyPrefixUtokenSupply
	supplies := sdk.NewCoins()

	iterator := func(key, val []byte) error {
		denom := types.DenomFromKey(key, prefix)

		var amount sdkmath.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled utoken supply should never happen
			return err
		}

		supplies = supplies.Add(sdk.NewCoin(denom, amount))
		return nil
	}

	util.Panic(k.iterate(ctx, prefix, iterator))

	return supplies
}
