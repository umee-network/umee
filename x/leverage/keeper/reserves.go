package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// GetReserveAmount gets the amount reserved of a specified token. On invalid
// asset, the reserved amount is zero.
func (k Keeper) GetReserveAmount(ctx sdk.Context, denom string) sdk.Int {
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

// SetReserveAmount sets the amount reserved of a specified token.
func (k Keeper) SetReserveAmount(ctx sdk.Context, coin sdk.Coin) error {
	if !k.IsAcceptedToken(ctx, coin.Denom) || !coin.IsValid() {
		return sdkerrors.Wrap(types.ErrInvalidAsset, coin.String())
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

// SweepBadDebts attempts to repay all bad debts in the system, efficiently
// stopping for any denomination where reserves become inqdequate. It should be
// called every Interest Epoch.
func (k Keeper) SweepBadDebts(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)

	// Outer iterator will iterate over all token denoms with bad debt
	denomPrefix := types.CreateBadDebtDenomKeyNoDenom()
	denomIter := sdk.KVStorePrefixIterator(store, denomPrefix)
	defer denomIter.Close()

	for ; denomIter.Valid(); denomIter.Next() {
		// denomKey is prefix | denom
		denomKey := denomIter.Key()
		denom := string(denomKey[len(denomPrefix):]) // remove prefix

		// Inner iterator will iterator over a single denom's bad debt addresses
		addrPrefix := types.CreateBadDebtAddressKeyNoAddress(denom)
		addrIter := sdk.KVStorePrefixIterator(store, addrPrefix)
		defer addrIter.Close()

		// Track whether this denom's reserves have failed to repay a bad debt
		reservesExhausted := false

		// Iterate over addresses with bad debt until denom's reserves exhausted
		for ; !reservesExhausted && addrIter.Valid(); addrIter.Next() {
			// addrKey is prefix | lengthPrefixed(denom) | addr
			addrKey := addrIter.Key()
			addr := sdk.AccAddress(addrKey[len(addrPrefix):]) // remove prefix and denom

			// attempt to repay a single address's debt in this denom
			fullyRepaid, err := k.RepayBadDebt(ctx, addr, denom)
			if err != nil {
				return err
			}
			if fullyRepaid {
				// If the bad debt of this denom at the given address was fully repaid,
				// clear the denom|address pair from this denom's bad debt address list
				k.SetBadDebtAddress(ctx, denom, addr, false)
			} else {
				// If the debt was not fully repaid, reserves are exhausted and iteration
				// over this denom should stop
				reservesExhausted = true
			}
		}

		if !reservesExhausted {
			// If all bad debt for this denom was successfully repaid,
			// clear the denom from the bad debt denoms list
			k.SetBadDebtDenom(ctx, denom, false)
		}
	}
	return nil
}

// RepayBadDebt uses reserves to repay borrower's debts of a given denom.
// It returns a boolean representing whether full repayment was achieved.
func (k Keeper) RepayBadDebt(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) (bool, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return false, sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	borrowed := k.GetBorrow(ctx, borrowerAddr, denom)
	reserved := k.GetReserveAmount(ctx, denom)

	amountToRepay := sdk.MinInt(borrowed.Amount, reserved)

	newBorrowed := borrowed.SubAmount(amountToRepay)
	newReserved := sdk.NewCoin(denom, reserved.Sub(amountToRepay))

	if amountToRepay.IsPositive() {
		if err := k.SetBorrow(ctx, borrowerAddr, newBorrowed); err != nil {
			return false, err
		}

		if err := k.SetReserveAmount(ctx, newReserved); err != nil {
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

	if newBorrowed.IsPositive() {
		k.Logger(ctx).Debug(
			"reserves exhausted",
			"denom", denom,
		)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeReservesExhausted,
				sdk.NewAttribute(types.EventAttrDenom, denom),
			),
		)
	}

	// True is returned on full repayment
	return newBorrowed.IsZero(), nil
}
