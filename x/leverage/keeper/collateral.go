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
	store := ctx.KVStore(k.storeKey)
	prefix := types.CreateCollateralAmountKeyNoDenom(borrowerAddr)

	totalCollateral := sdk.NewCoins()

	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// key is prefix | denom | 0x00
		key, val := iter.Key(), iter.Value()
		denom := string(key[len(prefix) : len(key)-1]) // remove prefix and null-terminator

		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled collateral amount should never happen
			panic(err)
		}

		// for each collateral coin found, add it to totalCollateral
		totalCollateral = totalCollateral.Add(sdk.NewCoin(denom, amount))
	}

	totalCollateral.Sort()
	return totalCollateral
}

// GetAllCollateral returns all collateral across all borrowers and asset types. Uses the
// CollateralAmount struct found in GenesisState, which stores borrower address as a string.
func (k Keeper) GetAllCollateral(ctx sdk.Context) []types.Collateral {
	prefix := types.KeyPrefixCollateralAmount
	collateral := []types.Collateral{}

	iterator := func(key, val []byte) error {
		addr := types.AddressFromKey(key, prefix)
		denom := types.DenomFromKeyWithAddress(key, prefix)

		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled collateral amount should never happen
			return err
		}

		collateral = append(collateral, types.NewCollateral(addr.String(), sdk.NewCoin(denom, amount)))
		return nil
	}

	err := k.iterate(ctx, prefix, iterator)
	if err != nil {
		panic(err)
	}

	return collateral
}

// GetEligibleLiquidationTargets returns a list of borrower addresses eligible for liquidation.
func (k Keeper) GetEligibleLiquidationTargets(ctx sdk.Context) ([]sdk.AccAddress, error) {
	store := ctx.KVStore(k.storeKey)
	borrowPrefix := types.CreateLoanKeyNoAddress()

	iter := sdk.KVStorePrefixIterator(store, borrowPrefix)
	defer iter.Close()

	liquidationTargets := []sdk.AccAddress{}
	checkedAddrs := map[string]struct{}{}

	// Iterate over all open borrows, adding addresses that are eligible for liquidation to a slice.
	for ; iter.Valid(); iter.Next() {
		// key is borrowPrefix | lengthPrefixed(borrowerAddr) | denom | 0x00
		key, _ := iter.Key(), iter.Value()

		// remove prefix | denom and null-terminator
		borrowerAddr := types.AddressFromKey(key, borrowPrefix)

		// if the address is already checked it can move on
		if _, ok := checkedAddrs[borrowerAddr.String()]; ok {
			continue
		}
		checkedAddrs[borrowerAddr.String()] = struct{}{}

		// get total borrowed by borrower (all denoms)
		borrowed := k.GetBorrowerBorrows(ctx, borrowerAddr)

		// get borrower uToken balances, for all uToken denoms enabled as collateral
		collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)

		// use oracle helper functions to find total borrowed value in USD
		borrowValue, err := k.TotalTokenValue(ctx, borrowed)
		if err != nil {
			return nil, err
		}

		// use collateral weights to compute borrow limit from enabled collateral
		borrowLimit, err := k.CalculateBorrowLimit(ctx, collateral)
		if err != nil {
			return nil, err
		}

		// check if the borrower's limit is bigger than the value
		if borrowLimit.GTE(borrowValue) {
			continue
		}

		// if the borrowLimit is smaller then the borrowValue
		// the address is eligible to liquidation
		liquidationTargets = append(liquidationTargets, borrowerAddr)
	}

	return liquidationTargets, nil
}
