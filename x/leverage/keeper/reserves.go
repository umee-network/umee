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

// GetAllReserves returns all reserves.
func (k Keeper) GetAllReserves(ctx sdk.Context) sdk.Coins {
	prefix := types.KeyPrefixReserveAmount
	reserves := sdk.NewCoins()

	iterator := func(key, val []byte) error {
		denom := types.DenomFromKey(key, prefix)

		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled reserve amount should never happen
			return err
		}

		reserves = reserves.Add(sdk.NewCoin(denom, amount))
		return nil
	}

	err := k.iterate(ctx, prefix, iterator)
	if err != nil {
		panic(err)
	}

	return reserves
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
// stopping for any denomination where reserves become exhausted. It should be
// called every Interest Epoch.
func (k Keeper) SweepBadDebts(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)

	// Iterator will iterate over all addresses with bad debt
	prefix := types.CreateBadDebtKeyNoAddress()
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// key is prefix | lengthPrefixed(addr) | denom | 0x00
		key := iter.Key()

		// remove prefix | lengthPrefix and denom | 0x00
		addr := key[len(prefix)+1 : len(prefix)+int(key[len(prefix)]+1)]

		// remove prefix | lengthPrefixed(addr) and null-terminator
		denom := string(key[len(prefix)+int(key[len(prefix)]+1) : len(key)-1])

		// attempt to repay a single address's debt in this denom
		fullyRepaid, err := k.RepayBadDebt(ctx, addr, denom)
		if err != nil {
			return err
		}
		if fullyRepaid {
			// If the bad debt of this denom at the given address was fully repaid,
			// clear the address|denom pair from the bad debt address list
			k.SetBadDebtAddress(ctx, denom, addr, false)
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
	amountToRepay = sdk.MinInt(amountToRepay, k.ModuleBalance(ctx, denom))

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
