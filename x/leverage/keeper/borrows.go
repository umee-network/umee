package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/umee-network/umee/x/leverage/types"
)

// GetBorrow returns an sdk.Coin representing how much of a given denom a borrower currently owes.
func (k Keeper) GetBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	owed := sdk.NewCoin(denom, sdk.ZeroInt())
	key := types.CreateLoanKey(borrowerAddr, denom)
	bz := store.Get(key)
	if bz != nil {
		err := owed.Amount.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
	}
	return owed
}

// SetBorrow sets the amount borrowed by an address in a given denom to an amount
func (k Keeper) SetBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string, amount sdk.Int) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := amount.Marshal()
	if err != nil {
		return err
	}
	store.Set(types.CreateLoanKey(borrowerAddr, denom), bz)
	return nil
}

// GetTotalBorrows returns total borrows across all borrowers and asset types as an sdk.Coins.
// It is done for all asset types at once, rather than one denom at a time, because either case would require
// iterating through all open borrows the way borrows are currently stored ( prefix | address | denom ).
func (k Keeper) GetTotalBorrows(ctx sdk.Context) (sdk.Coins, error) {
	store := ctx.KVStore(k.storeKey)
	prefix := types.KeyPrefixLoanToken
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	totalBorrowed := sdk.NewCoins()

	for ; iter.Valid(); iter.Next() {
		// key is prefix | lengthPrefixed(borrowerAddr) | denom | 0x00
		key, val := iter.Key(), iter.Value()
		// remove prefix | lengthPrefixed(addr) and null-terminator
		denom := string(key[len(prefix)+int(key[len(prefix)]+1) : len(key)-1])
		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			return sdk.NewCoins(), err // improperly marshaled borrow amount should never happen
		}
		// For each loan found, add it to totalBorrowed
		totalBorrowed = totalBorrowed.Add(sdk.NewCoin(denom, amount))
	}

	totalBorrowed.Sort()
	return totalBorrowed, nil
}

// GetBorrowerBorrows returns an sdk.Coins object containing all open borrows associated with an address
func (k Keeper) GetBorrowerBorrows(ctx sdk.Context, borrowerAddr sdk.AccAddress) (sdk.Coins, error) {
	totalBorrowed := sdk.NewCoins()
	store := ctx.KVStore(k.storeKey)
	prefix := types.CreateLoanKeyNoDenom(borrowerAddr)
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// key is prefix | denom | 0x00
		key, val := iter.Key(), iter.Value()
		denom := string(key[len(prefix) : len(key)-1]) // remove prefix and null-terminator
		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			return sdk.NewCoins(), err // improperly marshaled loan amount should never happen
		}
		// For each loan found, add it to totalBorrowed
		totalBorrowed = totalBorrowed.Add(sdk.NewCoin(denom, amount))
	}
	totalBorrowed.Sort()
	return totalBorrowed, nil
}

// GetBorrowUtilization derives the current borrow utilization of an asset type from the current total borrowed.
func (k Keeper) GetBorrowUtilization(ctx sdk.Context, denom string, totalBorrowed sdk.Int) (sdk.Dec, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}
	if totalBorrowed.IsNegative() {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrNegativeTotalBorrowed, totalBorrowed.String()+" "+denom)
	}

	// Token Utilization = Total Borrows / (Module Account Balance + Total Borrows - Reserve Requirement)
	moduleBalance := k.bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), denom).Amount
	reserveAmount := k.GetReserveAmount(ctx, denom)
	denominator := totalBorrowed.Add(moduleBalance).Sub(reserveAmount)
	if totalBorrowed.GTE(denominator) {
		// These edge cases can be safely interpreted as 100% utilization.
		return sdk.OneDec(), nil // denominator < totalBorrows (denominator may even be zero or negative)
	}
	// After the checks above (on both totalBorrowed and denominator), utilization is always between 0 and 1
	utilization := sdk.NewDecFromInt(totalBorrowed).Quo(sdk.NewDecFromInt(denominator))

	return utilization, nil
}
