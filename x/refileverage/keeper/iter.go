package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util"
	"github.com/umee-network/umee/v5/util/store"
	"github.com/umee-network/umee/v5/x/refileverage/types"
)

// iterate through all keys with a given prefix using a provided function.
// If the provided function returns an error, iteration stops and the error
// is returned.
func (k Keeper) iterate(ctx sdk.Context, prefix []byte, cb func(key, val []byte) error) error {
	return store.Iterate(ctx.KVStore(k.storeKey), prefix, cb)
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

// GetAllRegisteredTokens returns all the registered tokens from the x/refileverage
// module's KVStore.
func (k Keeper) GetAllRegisteredTokens(ctx sdk.Context) []types.Token {
	return store.MustLoadAll[*types.Token](ctx.KVStore(k.storeKey), types.KeyPrefixRegisteredToken)
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

// GetBorrowerBorrows returns an amount of $$$ value borrowed
func (k Keeper) GetBorrowerBorrows(ctx sdk.Context, borrowerAddr sdk.AccAddress) sdk.Dec {
	key := types.KeyAdjustedBorrow(borrowerAddr)
	d := store.GetDec(ctx.KVStore(k.storeKey), key, "borrows")
	return d.Mul(k.getInterestScalar(ctx, types.Gho))

}

// GetBorrowed returns an amount of $$$ coins borrowed
func (k Keeper) GetBorrowed(ctx sdk.Context, borrowerAddr sdk.AccAddress) sdk.Int {
	return k.GetBorrowerBorrows(ctx, borrowerAddr).Ceil().TruncateInt()
}

// GetBorrowerCollateral returns an sdk.Coins containing all of a borrower's collateral.
func (k Keeper) GetBorrowerCollateral(ctx sdk.Context, borrowerAddr sdk.AccAddress) sdk.Coins {
	prefix := types.KeyCollateralAmountNoDenom(borrowerAddr)
	totalCollateral := sdk.NewCoins()

	iterator := func(key, val []byte) error {
		borrowDenom := types.DenomFromKeyWithAddress(key, types.KeyPrefixCollateralAmount)
		var amount sdkmath.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled amount should never happen
			return err
		}

		totalCollateral = totalCollateral.Add(sdk.NewCoin(borrowDenom, amount))
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
		borrowerAddr := types.AddressFromKey(key, prefix)

		// if the address is already checked, do not check again
		if _, ok := checkedAddrs[borrowerAddr.String()]; ok {
			return nil
		}
		checkedAddrs[borrowerAddr.String()] = struct{}{}

		borrowed := k.GetBorrowerBorrows(ctx, borrowerAddr)
		collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)

		// compute liquidation threshold from enabled collateral
		// in this case, we can't reasonably skip missing prices but can move on
		// to the next borrower instead of stopping the entire query
		liquidationLimit, err := k.CalculateLiquidationThreshold(ctx, collateral)
		if err == nil && liquidationLimit.LT(borrowed) {
			// If liquidation limit is smaller than borrowed value then the
			// address is eligible for liquidation.
			liquidationTargets = append(liquidationTargets, borrowerAddr)
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

		if denom != types.Gho {
			// TODO: make it more efficient
			return nil
		}

		// clear blacklisted collateral while checking for any remaining (valid) collateral
		done, err := k.clearBlacklistedCollateral(ctx, addr)
		if err != nil {
			return err
		}

		// if remaining collateral is zero, attempt to repay this bad debt
		if !done {
			var err error
			done, err = k.RepayBadDebt(ctx, addr)
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
