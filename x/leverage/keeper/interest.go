package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// GetDynamicBorrowInterest derives the current borrow interest rate on an asset
// type, using utilization and params.
func (k Keeper) GetDynamicBorrowInterest(ctx sdk.Context, denom string, utilization sdk.Dec) (sdk.Dec, error) {
	if utilization.IsNegative() || utilization.GT(sdk.OneDec()) {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidUtilization, utilization.String())
	}

	kinkUtilization, err := k.GetInterestKinkUtilization(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	kinkRate, err := k.GetInterestAtKink(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	if utilization.GTE(kinkUtilization) {
		// utilization is between kink value and 100%
		maxRate, err := k.GetInterestMax(ctx, denom)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		return Interpolate(utilization, kinkUtilization, kinkRate, sdk.OneDec(), maxRate), nil
	}

	// utilization is between 0% and kink value
	baseRate, err := k.GetInterestBase(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return Interpolate(utilization, sdk.ZeroDec(), baseRate, kinkUtilization, kinkRate), nil
}

// AccrueAllInterest is called by EndBlock when BlockHeight % InterestEpoch == 0.
// It should accrue interest on all open borrows, increase reserves, and set
// LastInterestTime to BlockTime.
func (k Keeper) AccrueAllInterest(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)

	// get last time at which interest was accrued
	timeKey := types.CreateLastInterestTimeKey()
	prevInterestTime := sdk.ZeroInt()

	bz := store.Get(timeKey)
	if err := prevInterestTime.Unmarshal(bz); err != nil {
		return err
	}

	// Calculate time elapsed since last interest accrual (measured in years for APR math)
	currentTime := ctx.BlockTime().Unix()
	yearsElapsed := sdk.NewDec(currentTime - prevInterestTime.Int64()).QuoInt64(types.SecondsPerYear)

	// Compute total borrows across all borrowers, which are used when calculating
	// borrow utilization.
	totalBorrowed, err := k.GetTotalBorrows(ctx)
	if err != nil {
		return err
	}

	// Derive interest rate from utilization and parameters, for each denom found
	// in totalBorrowed, then multiply it by YearsElapsed to create the amount of
	// interest (expressed as a multiple of borrow amount) that will be applied to
	// each borrow position. Also collect reserve factors.
	interestToApply := map[string]sdk.Dec{}
	reserveFactors := map[string]sdk.Dec{}

	for _, coin := range totalBorrowed {
		reserveFactor, err := k.GetReserveFactor(ctx, coin.Denom)
		if err != nil {
			return err
		}

		utilization, err := k.GetBorrowUtilization(ctx, coin.Denom, coin.Amount)
		if err != nil {
			return err
		}

		derivedRate, err := k.GetDynamicBorrowInterest(ctx, coin.Denom, utilization)
		if err != nil {
			return err
		}

		reserveFactors[coin.Denom] = reserveFactor
		interestToApply[coin.Denom] = derivedRate.Mul(yearsElapsed)
	}

	newReserves := sdk.NewCoins()
	borrowPrefix := types.CreateLoanKeyNoAddress()

	iter := sdk.KVStorePrefixIterator(store, borrowPrefix)
	defer iter.Close()

	// Iterate over all open borrows, accruing interest on each and collecting new
	// reserves.
	for ; iter.Valid(); iter.Next() {
		// key is borrowPrefix | lengthPrefixed(borrowerAddr) | denom | 0x00
		key, val := iter.Key(), iter.Value()

		// remove prefix | lengthPrefixed(addr) and null-terminator
		denom := string(key[len(borrowPrefix)+int(key[len(borrowPrefix)]+1) : len(key)-1])

		var currentBorrow sdk.Int
		if err := currentBorrow.Unmarshal(val); err != nil {
			return err // improperly marshaled borrow amount should never happen
		}

		// Use previously calculated interestToApply (interest rate * time elapsed)
		// to accrue interest.
		amountToAccrue := interestToApply[denom].MulInt(currentBorrow).TruncateInt()
		bz, err := currentBorrow.Add(amountToAccrue).Marshal()
		if err != nil {
			return err
		}

		store.Set(key, bz)

		// A portion of amountToAccrue defined by the denom's reserve factor will be
		// set aside as reserves.
		newReserves = newReserves.Add(sdk.NewCoin(denom, reserveFactors[denom].MulInt(amountToAccrue).TruncateInt()))
	}

	// apply all reserve increases accumulated when iterating over borrows
	for _, coin := range newReserves {
		err = k.IncreaseReserves(ctx, coin)
		if err != nil {
			return err
		}
	}

	// set LastInterestTime
	bz, err = sdk.NewInt(currentTime).Marshal()
	if err != nil {
		return err
	}

	store.Set(timeKey, bz)
	return nil
}

// InitializeLastInterestTime sets LastInterestTime to present if it does not
// exist (used for genesis).
func (k *Keeper) InitializeLastInterestTime(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	timeKey := types.CreateLastInterestTimeKey()
	currentTime := ctx.BlockTime().Unix()

	if store.Get(timeKey) == nil {
		bz, err := sdk.NewInt(currentTime).Marshal()
		if err != nil {
			panic(err)
		}

		store.Set(timeKey, bz)
	}
}
