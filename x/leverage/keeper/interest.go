package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

// DeriveBorrowAPY derives the current borrow interest rate on a token denom
// using its supply utilization and token-specific params. Returns zero on
// invalid asset.
func (k Keeper) DeriveBorrowAPY(ctx sdk.Context, denom string) sdkmath.LegacyDec {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdkmath.LegacyZeroDec()
	}
	if token.Blacklist {
		// Regardless of params, AccrueAllInterest skips blacklisted denoms
		return sdkmath.LegacyZeroDec()
	}

	// Derive current supply utilization, which will always be between 0.0 and 1.0
	utilization := k.SupplyUtilization(ctx, denom)

	// Tokens which have reached or exceeded their max supply utilization always use max borrow APY
	if utilization.GTE(token.MaxSupplyUtilization) {
		return token.MaxBorrowRate
	}

	// Tokens which are past kink value but have not reached max supply utilization interpolate between the two
	if utilization.GTE(token.KinkUtilization) {
		return Interpolate(
			utilization,                // x
			token.KinkUtilization,      // x1
			token.KinkBorrowRate,       // y1
			token.MaxSupplyUtilization, // x2
			token.MaxBorrowRate,        // y2
		)
	}

	// utilization is between 0% and kink value
	return Interpolate(
		utilization,             // x
		sdkmath.LegacyZeroDec(), // x1
		token.BaseBorrowRate,    // y1
		token.KinkUtilization,   // x2
		token.KinkBorrowRate,    // y2
	)
}

// DeriveSupplyAPY derives the current supply interest rate on a token denom
// using its supply utilization and borrow APY. Returns zero on invalid asset.
func (k Keeper) DeriveSupplyAPY(ctx sdk.Context, denom string) sdkmath.LegacyDec {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdkmath.LegacyZeroDec()
	}

	borrowRate := k.DeriveBorrowAPY(ctx, denom)
	utilization := k.SupplyUtilization(ctx, denom)
	params := k.GetParams(ctx)
	reduction := params.OracleRewardFactor.Add(params.RewardsAuctionFee).Add(token.ReserveFactor)

	// supply APY = borrow APY * utilization, reduced by reserve factor and oracle reward factor
	return borrowRate.Mul(utilization).Mul(sdkmath.LegacyOneDec().Sub(reduction))
}

// AccrueAllInterest is called by EndBlock to update borrow positions.
// It accrues interest on all open borrows, increase reserves, funds
// oracle rewards, and sets LastInterestTime to BlockTime.
func (k Keeper) AccrueAllInterest(ctx sdk.Context) error {
	currentTime := ctx.BlockTime().Unix()
	prevInterestTime := k.getLastInterestTime(ctx)
	if prevInterestTime <= 0 {
		// if stored LastInterestTime is zero (or negative), either the chain has just started
		// or the genesis file has been modified intentionally. In either case, proceed as if
		// 0 seconds have passed since the last block, thus accruing no interest and setting
		// the current BlockTime as the new starting point.
		prevInterestTime = currentTime
	}

	// calculate time elapsed since last interest accrual (measured in years for APR math)
	if currentTime < prevInterestTime {
		// precaution against this and similar issues: https://github.com/tendermint/tendermint/issues/8773
		k.Logger(ctx).With("AccrueAllInterest will wait for block time > prevInterestTime").Error(
			types.ErrNegativeTimeElapsed.Error(),
			"current", currentTime,
			"prev", prevInterestTime,
		)

		// if LastInterestTime appears to be in the future, do nothing (besides logging) and leave
		// LastInterestTime at its stored value. This will repeat every block until BlockTime exceeds
		// LastInterestTime.
		return nil
	}

	yearsElapsed := sdkmath.LegacyNewDec(currentTime - prevInterestTime).QuoInt64(types.SecondsPerYear)
	if yearsElapsed.GTE(sdkmath.LegacyOneDec()) {
		// this safeguards primarily against misbehaving block time or incorrectly modified genesis states
		// which would accrue significant interest on borrows instantly. Chain will halt.
		return types.ErrExcessiveTimeElapsed.Wrapf("BlockTime: %d, LastInterestTime: %d",
			currentTime, prevInterestTime)
	}

	// fetch required parameters
	tokens := k.GetAllRegisteredTokens(ctx)
	params := k.GetParams(ctx)

	// create sdk.Coins objects to track oracle rewards, new reserves, and total interest accrued
	oracleRewards := sdk.NewCoins()
	auctionRewards := sdk.NewCoins()
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
		interestAccrued := prevTotalBorrowed.Mul(exponential.Sub(sdkmath.LegacyOneDec()))
		totalInterest = totalInterest.Add(sdk.NewCoin(
			token.BaseDenom,
			interestAccrued.TruncateInt(),
		))

		// if interest accrued on this denom is at least one base token
		if interestAccrued.GT(sdkmath.LegacyOneDec()) {
			// calculate new reserves gained for this denom, rounding up
			newReserves = newReserves.Add(sdk.NewCoin(
				token.BaseDenom,
				interestAccrued.Mul(token.ReserveFactor).Ceil().TruncateInt(),
			))
		}

		oracleRewards = oracleRewards.Add(sdk.NewCoin(
			token.BaseDenom,
			interestAccrued.Mul(params.OracleRewardFactor).TruncateInt(),
		))
		auctionRewards = auctionRewards.Add(sdk.NewCoin(
			token.BaseDenom,
			interestAccrued.Mul(params.RewardsAuctionFee).TruncateInt(),
		))
	}

	// apply all reserve increases accumulated when iterating over denoms
	for _, coin := range newReserves {
		if err := k.setReserves(ctx, coin.Add(k.GetReserves(ctx, coin.Denom))); err != nil {
			return err
		}
	}

	k.Logger(ctx).Debug("Rewards for oracle and auction", "oracleRewards", oracleRewards.String(),
		"auctionRewards", auctionRewards.String())
	if err := k.fundModules(ctx, oracleRewards, auctionRewards); err != nil {
		return err
	}

	err := k.setLastInterestTime(ctx, currentTime)
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
	sdkutil.Emit(&ctx, &types.EventInterestAccrual{
		BlockHeight:   uint64(ctx.BlockHeight()),
		Timestamp:     uint64(currentTime),
		TotalInterest: totalInterest,
		Reserved:      newReserves,
	})
	return nil
}
