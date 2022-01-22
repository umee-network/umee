package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// GetBorrow returns an sdk.Coin representing how much of a given denom a
// borrower currently owes.
func (k Keeper) GetBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	owed := sdk.NewCoin(denom, sdk.ZeroInt())
	key := types.CreateLoanKey(borrowerAddr, denom)

	if bz := store.Get(key); bz != nil {
		err := owed.Amount.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
	}

	return owed
}

// SetBorrow sets the amount borrowed by an address in a given denom.
// If the amount is zero, any stored value is cleared.
func (k Keeper) SetBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, borrow sdk.Coin) error {
	bz, err := borrow.Amount.Marshal()
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.CreateLoanKey(borrowerAddr, borrow.Denom)
	if borrow.Amount.IsZero() {
		store.Delete(key) // Completely repaid
	} else {
		store.Set(key, bz)
	}
	return nil
}

// GetAllBorrows returns all borrows across all borrowers and asset types. Uses the
// Borrow struct found in GenesisState, which stores borrower address as a string.
func (k Keeper) GetAllBorrows(ctx sdk.Context) []types.Borrow {
	prefix := types.KeyPrefixLoanToken
	borrows := []types.Borrow{}

	iterator := func(key, val []byte) error {
		addr := types.AddressFromKey(key, prefix)
		denom := types.DenomFromKeyWithAddress(key, prefix)

		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled borrow amount should never happen
			return err
		}

		borrows = append(borrows, types.NewBorrow(addr.String(), sdk.NewCoin(denom, amount)))
		return nil
	}

	err := k.iterate(ctx, prefix, iterator)
	if err != nil {
		panic(err)
	}

	return borrows
}

// GetTotalBorrows returns total borrows across all borrowers and asset types as
// an sdk.Coins. It is done for all asset types at once, rather than one denom
// at a time, because either case would require iterating through all open
// borrows the way borrows are currently stored ( prefix | address | denom ).
func (k Keeper) GetTotalBorrows(ctx sdk.Context) sdk.Coins {
	store := ctx.KVStore(k.storeKey)
	prefix := types.KeyPrefixLoanToken
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	totalBorrowed := sdk.NewCoins()

	for ; iter.Valid(); iter.Next() {
		// key is prefix | lengthPrefixed(borrowerAddr) | denom | 0x00
		key, val := iter.Key(), iter.Value()

		// remove prefix | lengthPrefixed(addr) and null-terminator
		denom := types.DenomFromKeyWithAddress(key, prefix)

		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled borrow amount should never happen
			panic(err)
		}

		// For each loan found, add it to totalBorrowed
		totalBorrowed = totalBorrowed.Add(sdk.NewCoin(denom, amount))
	}

	totalBorrowed.Sort()
	return totalBorrowed
}

// GetBorrowerBorrows returns an sdk.Coins object containing all open borrows
// associated with an address.
func (k Keeper) GetBorrowerBorrows(ctx sdk.Context, borrowerAddr sdk.AccAddress) sdk.Coins {
	store := ctx.KVStore(k.storeKey)
	prefix := types.CreateLoanKeyNoDenom(borrowerAddr)

	totalBorrowed := sdk.NewCoins()

	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// key is prefix | denom | 0x00
		key, val := iter.Key(), iter.Value()
		denom := string(key[len(prefix) : len(key)-1]) // remove prefix and null-terminator

		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled loan amount should never happen
			panic(err)
		}

		// for each loan found, add it to totalBorrowed
		totalBorrowed = totalBorrowed.Add(sdk.NewCoin(denom, amount))
	}

	totalBorrowed.Sort()
	return totalBorrowed
}

// GetBorrowUtilization derives the current borrow utilization of an asset type
// from the current total borrowed.
func (k Keeper) GetBorrowUtilization(ctx sdk.Context, denom string, totalBorrowed sdk.Int) (sdk.Dec, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}
	if totalBorrowed.IsNegative() {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrNegativeTotalBorrowed, totalBorrowed.String()+" "+denom)
	}

	// Token Utilization = Total Borrows / (Module Account Balance + Total Borrows - Reserve Requirement)
	moduleBalance := k.ModuleBalance(ctx, denom)
	reserveAmount := k.GetReserveAmount(ctx, denom)
	denominator := totalBorrowed.Add(moduleBalance).Sub(reserveAmount)
	if totalBorrowed.GTE(denominator) {
		// These edge cases can be safely interpreted as 100% utilization.
		// denominator < totalBorrows (denominator may even be zero or negative)
		return sdk.OneDec(), nil
	}

	// After the checks above (on both totalBorrowed and denominator), utilization
	// is always between 0 and 1.
	utilization := sdk.NewDecFromInt(totalBorrowed).Quo(sdk.NewDecFromInt(denominator))

	return utilization, nil
}

// CalculateBorrowLimit uses the price oracle to determine the borrow limit (in USD) provided by
// collateral sdk.Coins, using each token's uToken exchange rate and collateral weight.
// An error is returned if any input coins are not uTokens or if value calculation fails.
func (k Keeper) CalculateBorrowLimit(ctx sdk.Context, collateral sdk.Coins) (sdk.Dec, error) {
	limit := sdk.ZeroDec()

	for _, coin := range collateral {
		// convert uToken collateral to base assets
		baseAsset, err := k.ExchangeUToken(ctx, coin)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		// get USD value of base assets
		value, err := k.TokenValue(ctx, baseAsset)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		weight, err := k.GetCollateralWeight(ctx, baseAsset.Denom)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		// add each collateral coin's weighted value to borrow limit
		limit = limit.Add(value.Mul(weight))
	}

	return limit, nil
}

// SetBadDebtAddress sets or deletes an address in a denom's list of addresses with unpaid bad debt.
func (k Keeper) SetBadDebtAddress(ctx sdk.Context, denom string, borrowerAddr sdk.AccAddress, hasDebt bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateBadDebtKey(denom, borrowerAddr)

	if hasDebt {
		store.Set(key, []byte{0x01})
	} else {
		store.Delete(key)
	}
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

// GetBorrowAPY returns an sdk.Dec of an borrow APY
// returns sdk.ZeroDec if not found
func (k Keeper) GetBorrowAPY(ctx sdk.Context, denom string) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateBorrowAPYKey(denom)

	bz := store.Get(key)
	if bz == nil {
		return sdk.ZeroDec()
	}

	var borrowAPY sdk.Dec
	if err := borrowAPY.Unmarshal(bz); err != nil {
		panic(err)
	}

	return borrowAPY
}

// GetAllBorrowAPY returns all borrow APYs, arranged as a sorted sdk.DecCoins.
func (k Keeper) GetAllBorrowAPY(ctx sdk.Context) sdk.DecCoins {
	prefix := types.KeyPrefixBorrowAPY
	rates := sdk.NewDecCoins()

	iterator := func(key, val []byte) error {
		denom := types.DenomFromKey(key, prefix)

		var rate sdk.Dec
		if err := rate.Unmarshal(val); err != nil {
			// improperly marshaled APY should never happen
			return err
		}

		rates = rates.Add(sdk.NewDecCoinFromDec(denom, rate))
		return nil
	}

	err := k.iterate(ctx, prefix, iterator)
	if err != nil {
		panic(err)
	}

	return rates
}

// SetBorrowAPY sets the borrow APY of an specific denom
func (k Keeper) SetBorrowAPY(ctx sdk.Context, denom string, borrowAPY sdk.Dec) error {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	if borrowAPY.IsNegative() {
		return sdkerrors.Wrap(types.ErrNegativeAPY, denom+borrowAPY.String())
	}

	bz, err := borrowAPY.Marshal()
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.CreateBorrowAPYKey(denom)
	store.Set(key, bz)
	return nil
}

// GetAvailableToBorrow gets the amount available to borrow of a given token.
func (k Keeper) GetAvailableToBorrow(ctx sdk.Context, denom string) sdk.Int {
	// Available for borrow = Module Balance - Reserve Amount
	moduleBalance := k.ModuleBalance(ctx, denom)
	reserveAmount := k.GetReserveAmount(ctx, denom)

	return sdk.MaxInt(moduleBalance.Sub(reserveAmount), sdk.ZeroInt())
}
