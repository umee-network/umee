package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// DeriveBorrowAPY derives the current borrow interest rate on a token denom
// using its borrow utilization and token-specific params. Returns zero on
// invalid asset.
func (k Keeper) DeriveBorrowAPY(ctx sdk.Context, denom string) sdk.Dec {
	token, err := k.GetRegisteredToken(ctx, denom)
	if err != nil {
		return sdk.ZeroDec()
	}

	if token.Blacklist {
		// Regardless of params, AccrueAllInterest skips blacklisted denoms
		return sdk.ZeroDec()
	}

	utilization := k.DeriveBorrowUtilization(ctx, denom)

	if utilization.GTE(token.KinkUtilizationRate) {
		return Interpolate(
			utilization,               // x
			token.KinkUtilizationRate, // x1
			token.KinkBorrowRate,      // y1
			sdk.OneDec(),              // x2
			token.MaxBorrowRate,       // y2
		)
	}

	// utilization is between 0% and kink value
	return Interpolate(
		utilization,               // x
		sdk.ZeroDec(),             // x1
		token.BaseBorrowRate,      // y1
		token.KinkUtilizationRate, // x2
		token.KinkBorrowRate,      // y2
	)
}

// DeriveLendAPY derives the current lend interest rate on a token denom
// using its borrow utilization borrow APY. Returns zero on invalid asset.
func (k Keeper) DeriveLendAPY(ctx sdk.Context, denom string) sdk.Dec {
	token, err := k.GetRegisteredToken(ctx, denom)
	if err != nil {
		return sdk.ZeroDec()
	}

	borrowRate := k.DeriveBorrowAPY(ctx, denom)
	borrowUtilization := k.DeriveBorrowUtilization(ctx, denom)
	reduction := k.GetParams(ctx).OracleRewardFactor.Add(token.ReserveFactor)

	// lend APY = borrow APY * utilization, reduced by reserve factor and oracle reward factor
	return borrowRate.Mul(borrowUtilization).Mul(sdk.OneDec().Sub(reduction))
}

// AccrueAllInterest is called by EndBlock to rebase the borrowed positions.
// It accrue interest on all open borrows, increase reserves, and set
// LastInterestTime to BlockTime.
func (k Keeper) AccrueAllInterest(ctx sdk.Context) error {
	currentTime := ctx.BlockTime().Unix()
	prevInterestTime := k.GetLastInterestTime(ctx)
	if prevInterestTime == 0 {
		prevInterestTime = currentTime
	}

	// calculate time elapsed since last interest accrual (measured in years for APR math)
	secondsElapsed := currentTime - prevInterestTime
	if secondsElapsed < 0 {
		return sdkerrors.Wrap(types.ErrNegativeTimeElapsed, fmt.Sprintf("%d seconds", secondsElapsed))
	}
	yearsElapsed := sdk.NewDec(secondsElapsed).QuoInt64(types.SecondsPerYear)

	// fetch required parameters
	tokens := k.GetAllRegisteredTokens(ctx)
	oracleRewardFactor := k.GetParams(ctx).OracleRewardFactor

	// create sdk.Coins objects to track oracle rewards, new reserves, and total interest accrued
	oracleRewards := sdk.NewCoins()
	newReserves := sdk.NewCoins()
	totalInterest := sdk.NewCoins()

	// iterate over all accepted token denominations
	for _, token := range tokens {
		if token.Blacklist {
			// skip accruing interest on blacklisted assets
			continue
		}

		// interest is accrued by continuous compound interest on each denom's Interest Scalar
		scalar := k.getInterestScalar(ctx, token.BaseDenom)
		// calculate e^(APY*time)
		exponential := ApproxExponential(k.DeriveBorrowAPY(ctx, token.BaseDenom).Mul(yearsElapsed))
		// multiply interest scalar by e^(APY*time)
		if err := k.setInterestScalar(ctx, token.BaseDenom, scalar.Mul(exponential)); err != nil {
			return err
		}

		// apply (pre-accural) interest scalar to borrows to get total borrowed before interest accrued
		prevTotalBorrowed := k.getAdjustedTotalBorrowed(ctx, token.BaseDenom).Mul(scalar)

		// calculate total interest accrued for this denom
		interestAccrued := prevTotalBorrowed.Mul(exponential.Sub(sdk.OneDec()))
		totalInterest = totalInterest.Add(sdk.NewCoin(
			token.BaseDenom,
			interestAccrued.TruncateInt(),
		))

		// calculate new reserves accrued for this denom
		newReserves = newReserves.Add(sdk.NewCoin(
			token.BaseDenom,
			interestAccrued.Mul(token.ReserveFactor).TruncateInt(),
		))

		// calculate oracle rewards accrued for this denom
		oracleRewards = oracleRewards.Add(sdk.NewCoin(
			token.BaseDenom,
			interestAccrued.Mul(oracleRewardFactor).TruncateInt(),
		))
	}

	// apply all reserve increases accumulated when iterating over denoms
	for _, coin := range newReserves {
		if err := k.setReserveAmount(ctx, coin.AddAmount(k.GetReserveAmount(ctx, coin.Denom))); err != nil {
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
		"interest accrued",
		"block_height", fmt.Sprintf("%d", ctx.BlockHeight()),
		"unix_time", fmt.Sprintf("%d", currentTime),
		"interest", totalInterest.String(),
		"reserved", newReserves.String(),
	)

	// TODO: use typed events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeInterestAccrual,
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

// GetLastInterestTime returns unix timestamp (in seconds) when the lasat interest was accrued.
// Returns 0 if the value if the value is absent.
func (k Keeper) GetLastInterestTime(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	timeKey := types.CreateLastInterestTimeKey()
	bz := store.Get(timeKey)
	if bz == nil {
		return 0
	}

	val := gogotypes.Int64Value{}
	if err := k.cdc.Unmarshal(bz, &val); err != nil {
		panic(err)
	}

	return val.Value
}
