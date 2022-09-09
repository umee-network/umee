package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v3/x/leverage/types"
)

// checkBadDebt detects if a borrower has zero non-blacklisted collateral,
// and marks any remaining borrowed tokens as bad debt.
func (k Keeper) checkBadDebt(ctx sdk.Context, borrowerAddr sdk.AccAddress) error {
	// get remaining collateral uTokens, ignoring blacklisted
	remainingCollateral := k.filterAcceptedUTokens(ctx, k.GetBorrowerCollateral(ctx, borrowerAddr))

	// detect bad debt if collateral is completely exhausted
	if remainingCollateral.IsZero() {
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
