package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"

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
	prevInterestTime := k.GetLastInterestTime(ctx)

	// Calculate time elapsed since last interest accrual (measured in years for APR math)
	currentTime := ctx.BlockTime().Unix()
	yearsElapsed := sdk.NewDec(currentTime - prevInterestTime).QuoInt64(types.SecondsPerYear)
	if yearsElapsed.IsNegative() {
		return sdkerrors.Wrap(types.ErrNegativeTimeElapsed, yearsElapsed.String()+" years")
	}

	// Compute total borrows across all borrowers, which are used when calculating
	// borrow utilization.
	totalBorrowed := k.GetTotalBorrows(ctx)

	interestToApply := map[string]sdk.Dec{}
	reserveFactors := map[string]sdk.Dec{}

	// Derive interest rate from utilization and parameters, for each denom found
	// in totalBorrowed, then multiply it by YearsElapsed to create the amount of
	// interest (expressed as a multiple of borrow amount) that will be applied to
	// each borrow position. Also collect reserve factors.
	for _, coin := range totalBorrowed {
		reserveFactor, err := k.GetReserveFactor(ctx, coin.Denom)
		if err != nil {
			return err
		}

		utilization, err := k.GetBorrowUtilization(ctx, coin.Denom, coin.Amount)
		if err != nil {
			return err
		}

		borrowRate, err := k.GetDynamicBorrowInterest(ctx, coin.Denom, utilization)
		if err != nil {
			return err
		}

		reserveFactors[coin.Denom] = reserveFactor
		interestToApply[coin.Denom] = borrowRate.Mul(yearsElapsed)

		if err := k.SetBorrowAPY(ctx, coin.Denom, borrowRate); err != nil {
			return err
		}

		lendRate := borrowRate.Mul(utilization).Mul(sdk.OneDec().Sub(reserveFactor))
		if err := k.SetLendAPY(ctx, coin.Denom, lendRate); err != nil {
			return err
		}
	}

	oracleRewards := sdk.NewCoins()
	newReserves := sdk.NewCoins()
	totalInterest := sdk.NewCoins()
	borrowPrefix := types.CreateLoanKeyNoAddress()

	iter := sdk.KVStorePrefixIterator(store, borrowPrefix)
	defer iter.Close()

	// Iterate over all open borrows, accruing interest on each and collecting new
	// reserves.
	for ; iter.Valid(); iter.Next() {
		// key is borrowPrefix | lengthPrefixed(borrowerAddr) | denom | 0x00
		key, val := iter.Key(), iter.Value()

		// remove prefix | lengthPrefixed(addr) and null-terminator
		denom := types.DenomFromKeyWithAddress(key, borrowPrefix)

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

		// Track total interest across all borrowers for logging
		totalInterest = totalBorrowed.Add(sdk.NewCoin(denom, amountToAccrue))

		// A portion of amountToAccrue defined by the denom's reserve factor will be
		// set aside as reserves.
		newReserves = newReserves.Add(sdk.NewCoin(denom, reserveFactors[denom].MulInt(amountToAccrue).TruncateInt()))

		// A portion of amountToAccrue defined by the param OracleRewardFactor will be
		// sent to the oracle module account to fund the reward pool.
		oracleRewards = oracleRewards.Add(
			sdk.NewCoin(denom, k.GetParams(ctx).OracleRewardFactor.MulInt(amountToAccrue).TruncateInt()),
		)
	}

	// apply all reserve increases accumulated when iterating over borrows
	for _, coin := range newReserves {
		if coin.IsNegative() {
			return sdkerrors.Wrap(types.ErrInvalidAsset, coin.String())
		}

		if err := k.SetReserveAmount(ctx, coin.AddAmount(k.GetReserveAmount(ctx, coin.Denom))); err != nil {
			return err
		}
	}

	// fund oracle reward pool
	if err := k.FundOracle(ctx, oracleRewards); err != nil {
		return err
	}

	// set LastInterestTime
	err := k.SetLastInterestTime(ctx, currentTime)
	if err != nil {
		return err
	}

	// Because this action is not caused by a message, logging and
	// events are here instead of msg_server.go
	k.Logger(ctx).Debug(
		"interest epoch",
		"block_height", fmt.Sprintf("%d", ctx.BlockHeight()),
		"unix_time", fmt.Sprintf("%d", currentTime),
		"interest", totalInterest.String(),
		"reserved", newReserves.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeInterestEpoch,
			sdk.NewAttribute(types.EventAttrBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.EventAttrUnixTime, fmt.Sprintf("%d", currentTime)),
			sdk.NewAttribute(types.EventAttrInterest, totalInterest.String()),
			sdk.NewAttribute(types.EventAttrReserved, newReserves.String()),
		),
	})

	return nil
}

// SetLastInterestTime sets LastInterestTime to a given value
func (k *Keeper) SetLastInterestTime(ctx sdk.Context, interestTime int64) error {
	store := ctx.KVStore(k.storeKey)
	timeKey := types.CreateLastInterestTimeKey()

	bz, err := k.cdc.Marshal(&gogotypes.Int64Value{Value: interestTime})
	if err != nil {
		return err
	}

	store.Set(timeKey, bz)
	return nil
}

// GetLastInterestTime gets last time at which interest was accrued
func (k Keeper) GetLastInterestTime(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	timeKey := types.CreateLastInterestTimeKey()
	bz := store.Get(timeKey)

	val := gogotypes.Int64Value{}

	if err := k.cdc.Unmarshal(bz, &val); err != nil {
		panic(err)
	}

	return val.Value
}
