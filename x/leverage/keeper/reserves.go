package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v3/x/leverage/types"
)

// clearBlacklistedCollateral decollateralizes any blacklisted uTokens
// from a borrower's collateral. It is used during liquidations and before
// repaying bad debts with reserves to make any subsequent checks for
// remaining collateral on the borrower's address more efficient.
// Also returns a boolean indicating whether valid collateral remains.
func (k Keeper) clearBlacklistedCollateral(ctx sdk.Context, borrowerAddr sdk.AccAddress) (bool, error) {
	collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)
	hasCollateral := false
	for _, coin := range collateral {
		denom := types.ToTokenDenom(coin.Denom)
		token, err := k.GetTokenSettings(ctx, denom)
		if err != nil {
			return false, err
		}
		if token.Blacklist {
			// Decollateralize any blacklisted uTokens encountered
			err := k.decollateralize(ctx, borrowerAddr, borrowerAddr, coin)
			if err != nil {
				return false, err
			}
		} else {
			// At least one non-blacklisted uToken was found
			hasCollateral = true
		}
	}
	// Any remaining collateral is non-blacklisted
	return hasCollateral, nil
}

// checkBadDebt detects if a borrower has zero non-blacklisted collateral,
// and marks any remaining borrowed tokens as bad debt.
func (k Keeper) checkBadDebt(ctx sdk.Context, borrowerAddr sdk.AccAddress) error {
	// clear blacklisted collateral while checking for any remaining (valid) collateral
	hasCollateral, err := k.clearBlacklistedCollateral(ctx, borrowerAddr)
	if err != nil {
		return err
	}

	// mark bad debt if collateral is completely exhausted
	if !hasCollateral {
		for _, coin := range k.GetBorrowerBorrows(ctx, borrowerAddr) {
			// set a bad debt flag for each borrowed denom
			if err := k.setBadDebtAddress(ctx, borrowerAddr, coin.Denom, true); err != nil {
				return err
			}
		}
	}

	return nil
}

// RepayBadDebt uses reserves to repay borrower's debts of a given denom.
// It returns a boolean representing whether full repayment was achieved.
// This function assumes the borrower has already been verified to have
// no collateral remaining.
func (k Keeper) RepayBadDebt(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) (bool, error) {
	borrowed := k.GetBorrow(ctx, borrowerAddr, denom)
	borrower := borrowerAddr.String()
	reserved := k.GetReserves(ctx, denom).Amount

	amountToRepay := sdk.MinInt(borrowed.Amount, reserved)
	amountToRepay = sdk.MinInt(amountToRepay, k.ModuleBalance(ctx, denom).Amount)

	newBorrowed := borrowed.SubAmount(amountToRepay)
	newReserved := sdk.NewCoin(denom, reserved.Sub(amountToRepay))

	if amountToRepay.IsPositive() {
		if err := k.setBorrow(ctx, borrowerAddr, newBorrowed); err != nil {
			return false, err
		}

		if err := k.setReserves(ctx, newReserved); err != nil {
			return false, err
		}

		// This action is not caused by a message so we need to make an event here
		asset := sdk.NewCoin(denom, amountToRepay)
		k.Logger(ctx).Debug(
			"bad debt repaid",
			"borrower", borrower,
			"asset", asset,
		)
		err := ctx.EventManager().EmitTypedEvent(&types.EventRepayBadDebt{
			Borrower: borrower, Asset: asset,
		})
		if err != nil {
			return false, err
		}
	}

	newModuleBalance := k.ModuleBalance(ctx, denom)

	// Reserve exhaustion logs track any bad debts that were not repaid
	if newBorrowed.IsPositive() {
		k.Logger(ctx).Debug(
			"reserves exhausted",
			"borrower", borrower,
			"asset", newBorrowed,
			"module balance", newModuleBalance,
			"reserves", newReserved,
		)
		err := ctx.EventManager().EmitTypedEvent(&types.EventReservesExhausted{
			Borrower: borrower, OutstandingDebt: newBorrowed,
			ModuleBalance: newModuleBalance, Reserves: newReserved,
		})
		if err != nil {
			return false, err
		}
	}

	// True is returned on full repayment
	return newBorrowed.IsZero(), nil
}
