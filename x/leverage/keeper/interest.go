package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/umee-network/umee/x/leverage/types"
)

// GetDynamicBorrowInterest derives the current borrow interest rate on an asset type, using utilization and params.
func (k Keeper) GetDynamicBorrowInterest(ctx sdk.Context, denom string, utilization sdk.Dec) (sdk.Dec, error) {
	if utilization.IsNegative() || utilization.GTE(sdk.OneDec()) {
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
		// Utilization is between kink value and 100%
		maxRate, err := k.GetInterestMax(ctx, denom)
		if err != nil {
			return sdk.ZeroDec(), err
		}
		return k.Interpolate(ctx, utilization, kinkUtilization, kinkRate, sdk.OneDec(), maxRate), nil
	}
	// Utilization is between 0% and kink value
	baseRate, err := k.GetInterestBase(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return k.Interpolate(ctx, utilization, sdk.ZeroDec(), baseRate, kinkUtilization, kinkRate), nil
}

// AccrueAllInterest is called by EndBlock when BlockHeight % InterestEpoch == 0.
// It should accrue interest on all open borrows, increase reserves, and set LastInterestTime to BlockTime.
func (k Keeper) AccrueAllInterest(ctx sdk.Context) error {
	// Get last time at which interest was accrued
	store := ctx.KVStore(k.storeKey)
	timeKey := types.CreateLastInterestTimeKey()
	prevInterestTime := sdk.ZeroInt()
	bz := store.Get(timeKey)
	err := prevInterestTime.Unmarshal(bz)
	if err != nil {
		panic(err)
	}

	// Calculate time elapsed since last interest accrual (measured in years for APR math)
	currentTime := ctx.BlockTime().Unix()
	// yearsElapsed := sdk.NewDec(currentTime - prevInterestTime.Int64()).QuoInt64(31536000)

	// Compute total borrows across all borrowers, required for utilization and interest rate derivation
	totalBorrowed, err := k.GetTotalBorrows(ctx)
	if err != nil {
		return err
	}

	// Derive interest rate from utilization and parameters, for each denom found in totalBorrowed
	interestRates := map[string]sdk.Dec{}
	for _, coin := range totalBorrowed {
		utilization, err := k.GetBorrowUtilization(ctx, coin.Denom, coin.Amount)
		if err != nil {
			return err
		}
		derivedRate, err := k.GetDynamicBorrowInterest(ctx, coin.Denom, utilization)
		if err != nil {
			return err
		}
		interestRates[coin.Denom] = derivedRate
	}

	// This sdk.Coins will collect reserve increases from all borrow positions
	newReserves := sdk.NewCoins()

	// - - - TODO - - -
	// Iterate over all loans
	//   + Calculate interest and increase amount owed
	//   + Only derive interest rate once per token, then save and reuse
	// - - - TODO - - -

	// Apply all reserve increases accumulated when iterating over borrows
	for _, coin := range newReserves {
		err = k.IncreaseReserves(ctx, coin)
		if err != nil {
			return err
		}
	}

	// Set LastInterestTime
	bz, err = sdk.NewInt(currentTime).Marshal()
	if err != nil {
		panic(err)
	}
	store.Set(timeKey, bz)
	return nil
}
