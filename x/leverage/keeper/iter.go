package keeper

import (
	"sort"

	sdkmath "cosmossdk.io/math"
	prefixstore "github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util"
	"github.com/umee-network/umee/v5/util/keys"
	"github.com/umee-network/umee/v5/util/store"
	"github.com/umee-network/umee/v5/x/leverage/types"
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

// GetAllRegisteredTokens returns all the registered tokens from the x/leverage
// module's KVStore.
func (k Keeper) GetAllRegisteredTokens(ctx sdk.Context) []types.Token {
	return store.MustLoadAll[*types.Token](ctx.KVStore(k.storeKey), types.KeyPrefixRegisteredToken)
}

// GetAllSpecialAssetPairs returns all the special asset pairs from the x/leverage
// module's KVStore.
func (k Keeper) GetAllSpecialAssetPairs(ctx sdk.Context) []types.SpecialAssetPair {
	pairs := store.MustLoadAll[*types.SpecialAssetPair](ctx.KVStore(k.storeKey), types.KeyPrefixSpecialAssetPair)
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].CollateralWeight.LT(pairs[j].CollateralWeight)
	})
	return pairs
}

// GetSpecialAssetPairs returns all the special asset pairs from the x/leverage
// module's KVStore which match a single asset.
func (k Keeper) GetSpecialAssetPairs(ctx sdk.Context, denom string) []types.SpecialAssetPair {
	prefix := types.KeySpecialAssetPairOneDenom(denom)
	pairs := store.MustLoadAll[*types.SpecialAssetPair](ctx.KVStore(k.storeKey), prefix)
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].CollateralWeight.LT(pairs[j].CollateralWeight)
	})
	return pairs
}

// GetBorrowerBorrows returns an sdk.Coins object containing all open borrows
// associated with an address.
func (k Keeper) GetBorrowerBorrows(ctx sdk.Context, borrowerAddr sdk.AccAddress) sdk.Coins {
	prefix := types.KeyAdjustedBorrowNoDenom(borrowerAddr)
	totalBorrowed := sdk.NewCoins()

	iterator := func(key, val []byte) error {
		borrowDenom := types.DenomFromKeyWithAddress(key, types.KeyPrefixAdjustedBorrow)
		var adjustedAmount sdk.Dec
		if err := adjustedAmount.Unmarshal(val); err != nil {
			// improperly marshaled borrow amount should never happen
			return err
		}

		// apply interest scalar
		amount := adjustedAmount.Mul(k.getInterestScalar(ctx, borrowDenom)).Ceil().TruncateInt()
		totalBorrowed = totalBorrowed.Add(sdk.NewCoin(borrowDenom, amount))
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

		position, err := k.getAccountPosition(ctx, borrowerAddr, true)
		if err == nil {
			borrowValue := position.BorrowedValue()
			liquidationThreshold := position.Limit()
			if liquidationThreshold.LT(borrowValue) {
				// If liquidation threshold is smaller than borrowed value then the
				// address is eligible for liquidation.
				liquidationTargets = append(liquidationTargets, borrowerAddr)
			}
		}
		// Non-price errors will cause the query itself to fail, whereas oracle errors
		// simply cause the address to be skipped
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
	return store.SumCoins(k.prefixStore(ctx, types.KeyPrefixUtokenSupply), keys.NoLastByte)
}

// GetAllReserves returns all reserves.
func (k Keeper) GetAllReserves(ctx sdk.Context) sdk.Coins {
	return store.SumCoins(k.prefixStore(ctx, types.KeyPrefixReserveAmount), keys.NoLastByte)
}

func (k Keeper) prefixStore(ctx sdk.Context, p []byte) storetypes.KVStore {
	return prefixstore.NewStore(ctx.KVStore(k.storeKey), p)
}
