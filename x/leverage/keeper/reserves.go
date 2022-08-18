package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// GetReserveAmount gets the amount reserved of a specified token. On invalid
// asset, the reserved amount is zero.
func (k Keeper) GetReserveAmount(ctx sdk.Context, denom string) sdkmath.Int {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateReserveAmountKey(denom)
	amount := sdk.ZeroInt()

	if bz := store.Get(key); bz != nil {
		err := amount.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
	}

	if amount.IsNegative() {
		panic("negative reserve amount detected")
	}

	return amount
}

// setReserveAmount sets the amount reserved of a specified token.
func (k Keeper) setReserveAmount(ctx sdk.Context, coin sdk.Coin) error {
	if err := coin.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	reserveKey := types.CreateReserveAmountKey(coin.Denom)

	// save the new reserve amount
	bz, err := coin.Amount.Marshal()
	if err != nil {
		return err
	}

	store.Set(reserveKey, bz)
	return nil
}

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
	reserved := k.GetReserveAmount(ctx, denom)

	amountToRepay := sdk.MinInt(borrowed.Amount, reserved)
	amountToRepay = sdk.MinInt(amountToRepay, k.ModuleBalance(ctx, denom))

	newBorrowed := borrowed.SubAmount(amountToRepay)
	newReserved := sdk.NewCoin(denom, reserved.Sub(amountToRepay))

	if amountToRepay.IsPositive() {
		if err := k.setBorrow(ctx, borrowerAddr, newBorrowed); err != nil {
			return false, err
		}

		if err := k.setReserveAmount(ctx, newReserved); err != nil {
			return false, err
		}

		// Because this action is not caused by a message, logging and
		// events are here instead of msg_server.go
		k.Logger(ctx).Debug(
			"bad debt repaid",
			"borrower", borrowerAddr.String(),
			"denom", denom,
			"amount", amountToRepay.String(),
		)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRepayBadDebt,
				sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
				sdk.NewAttribute(types.EventAttrDenom, denom),
				sdk.NewAttribute(sdk.AttributeKeyAmount, amountToRepay.String()),
			),
		)
	}

	// Reserve exhaustion logs track any bad debts that were not repaid
	if newBorrowed.IsPositive() {
		k.Logger(ctx).Debug(
			"reserves exhausted",
			"borrower", borrowerAddr.String(),
			"denom", denom,
			"amount", newBorrowed.Amount.String(),
		)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeReservesExhausted,
				sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
				sdk.NewAttribute(types.EventAttrDenom, denom),
				sdk.NewAttribute(sdk.AttributeKeyAmount, newBorrowed.Amount.String()),
			),
		)
	}

	// True is returned on full repayment
	return newBorrowed.IsZero(), nil
}
