package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/umee-network/umee/x/leverage/types"
)

// GetBorrowUtilization derives the current borrow utilization of an asset type from the current total borrowed.
func (k Keeper) GetBorrowUtilization(ctx sdk.Context, denom string, totalBorrowed sdk.Int) (sdk.Dec, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}
	if totalBorrowed.IsZero() {
		return sdk.ZeroDec(), nil // catch this 0% utilization case before reading balances and reserves
	}
	if totalBorrowed.IsNegative() {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrNegativeTotalBorrowed, totalBorrowed.String()+" "+denom)
	}

	// Token Utilization = Total Borrows / (Module Account Balance + Total Borrows - Reserve Requirement)
	moduleBalance := k.bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), denom).Amount
	reserveAmount := k.GetReserveAmount(ctx, denom)
	denominator := sdk.NewDecFromInt(totalBorrowed.Add(moduleBalance).Sub(reserveAmount))
	if sdk.NewDecFromInt(totalBorrowed).GTE(denominator) {
		// These edge cases can be safely interpreted as 100% utilization.
		return sdk.OneDec(), nil // denominator < totalBorrows (denominator may even be zero or negative)
	}
	// After the checks above (on both totalBorrowed and denominator), utilization is always between 0 and 1
	utilization := sdk.NewDecFromInt(totalBorrowed).Quo(denominator)

	return utilization, nil
}

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
	totalBorrowed, err := k.GetTotalLoans(ctx)
	if err != nil {
		return err
	}

	// Derive interest rate from utilization and parameters, for all denoms found in totalBorrowed
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

	// This sdk.Coins will aggregate reserve increases from all borrow positions
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
