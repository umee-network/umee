package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/refileverage/types"
)

// getLiquidationAmounts takes a repayment and reward denom proposed by a liquidator and calculates
// the actual repayment amount a target address is eligible for, and the corresponding collateral
// to liquidate and equivalent base rewards to send to the liquidator.
func (k Keeper) getLiquidationAmounts(
	ctx sdk.Context,
	liquidatorAddr,
	targetAddr sdk.AccAddress,
	repay sdk.Int,
	rewardDenom string,
	directLiquidation bool,
) (tokenRepay sdk.Int, collateralLiquidate sdk.Coin, tokenReward sdk.Coin, err error) {
	// get relevant liquidator, borrower, and module balances
	borrowerCollateral := k.GetBorrowerCollateral(ctx, targetAddr)
	borrowedAmount := k.GetBorrowed(ctx, targetAddr)
	borrowedVal := GhoIntToDec(borrowedAmount)

	collateralValue, err := k.CalculateCollateralValue(ctx, borrowerCollateral)
	if err != nil {
		return sdk.Int{}, sdk.Coin{}, sdk.Coin{}, err
	}
	liquidationThreshold, err := k.CalculateLiquidationThreshold(ctx, borrowerCollateral)
	if err != nil {
		return sdk.Int{}, sdk.Coin{}, sdk.Coin{}, err
	}
	if borrowedVal.LT(liquidationThreshold) {
		// borrower is healthy and cannot be liquidated
		return sdk.Int{}, sdk.Coin{}, sdk.Coin{}, types.ErrLiquidationIneligible
	}

	// get liquidation incentive
	ts, err := k.GetTokenSettings(ctx, rewardDenom)
	if err != nil {
		return sdk.Int{}, sdk.Coin{}, sdk.Coin{}, err
	}
	repayVal := GhoIntToDec(repay)

	// get dynamic close factor
	params := k.GetParams(ctx)
	closeFactor := ComputeCloseFactor(
		borrowedVal,
		collateralValue,
		liquidationThreshold,
		params.SmallLiquidationSize,
		params.MinimumCloseFactor,
		params.CompleteLiquidationThreshold,
	)
	// maximum USD value that can be repaid
	maxRepayValue := borrowedVal.Mul(closeFactor)
	// determine fraction of borrowed repayDenom which can be repaid after close factor
	maxRepayAfterCloseFactor := borrowedAmount
	if maxRepayValue.LT(repayVal) {
		maxRepayRatio := maxRepayValue.Quo(repayVal)
		maxRepayAfterCloseFactor = maxRepayRatio.MulInt(borrowedAmount).RoundInt()
	}

	rewardPrice, _, err := k.TokenPrice(ctx, rewardDenom, types.PriceModeSpot)
	if err != nil {
		return sdk.Int{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// get collateral uToken exchange rate
	exchangeRate := k.DeriveExchangeRate(ctx, rewardDenom)

	// Reduce liquidation incentive if the liquidator has specified they would like to directly receive base assets.
	// Since this fee also reduces the amount of collateral that must be burned, it is applied before any other
	// computations, as if the token itself had a smaller liquidation incentive.
	liqudationIncentive := ts.LiquidationIncentive
	if directLiquidation {
		liqudationIncentive = liqudationIncentive.Mul(sdk.OneDec().Sub(params.DirectLiquidationFee))
	}

	maxRepay := repay // maximum allowed by liquidator
	// TODO reconcillate repay bank
	// availableRepay := k.bankKeeper.SpendableCoins(ctx, liquidatorAddr).AmountOf(repayDenom)

	// TODO: maxRepay = sdk.MinInt(maxRepay, availableRepay)           // liquidator account balance
	maxRepay = sdk.MinInt(maxRepay, borrowedAmount)           // borrower position
	maxRepay = sdk.MinInt(maxRepay, maxRepayAfterCloseFactor) // close factor

	collateralDenom := types.ToUTokenDenom(rewardDenom)
	repay, burn, reward := ComputeLiquidation(
		maxRepay,
		borrowerCollateral.AmountOf(collateralDenom),
		rewardPrice,
		exchangeRate,
		liqudationIncentive,
	)

	return repay, sdk.NewCoin(collateralDenom, burn), sdk.NewCoin(rewardDenom, reward), nil
}

// ComputeLiquidation takes the conditions preceding a liquidation and outputs the amounts
// of base token that should be repaid, collateral uToken burned, and reward token allocated
// as a result of the transaction, after accounting for limiting factors with as little
// rounding as possible. Inputs are as follows:
// - availableRepay: The lowest (in repay denom) of either liquidator balance, max repayment, or borrowed amount.
// - availableRewardUtokens: The amount of the reward uToken denom which borrower has as collateral
// - priceRatio: The ratio of repayPrice / rewardPrice, which is used when computing rewards
// - uTokenExchangeRate: The uToken exchange rate from collateral uToken denom to reward base denom
// - liquidationIncentive: The liquidation incentive of the token reward denomination
func ComputeLiquidation(
	availableRepay,
	availableRewardUtokens sdkmath.Int,
	rewardPrice,
	uTokenExchangeRate,
	liquidationIncentive sdk.Dec,
) (tokenRepay sdkmath.Int, collateralBurn sdkmath.Int, tokenReward sdkmath.Int) {
	// Prevent division by zero
	if uTokenExchangeRate.IsZero() || rewardPrice.IsZero() {
		return sdkmath.ZeroInt(), sdkmath.ZeroInt(), sdkmath.ZeroInt()
	}

	// Start with the maximum possible repayment amount, as a decimal
	maxRepay := toDec(availableRepay)
	// Determine the base maxReward amount that would result from maximum repayment

	maxReward := maxRepay.Quo(rewardPrice).Mul(sdk.OneDec().Add(liquidationIncentive))
	// Determine the maxCollateral burn amount that corresponds to base reward amount
	maxCollateral := maxReward.Quo(uTokenExchangeRate)

	// Catch no-ops early
	if maxRepay.IsZero() ||
		maxReward.IsZero() ||
		maxCollateral.IsZero() {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	}

	// We will track limiting factors by the ratio by which the max repayment would need to be reduced to comply
	ratio := sdk.OneDec()

	// Collateral burned cannot exceed borrower's collateral
	ratio = sdk.MinDec(ratio,
		toDec(availableRewardUtokens).Quo(maxCollateral),
	)
	// Catch edge cases
	ratio = sdk.MaxDec(ratio, sdk.ZeroDec())

	maxRepay = maxRepay.Mul(ratio)
	maxCollateral = maxCollateral.Mul(ratio)

	tokenRepay = maxRepay.Ceil().RoundInt()
	collateralBurn = maxCollateral.TruncateInt()
	tokenReward = toDec(collateralBurn).Mul(uTokenExchangeRate).TruncateInt()

	return tokenRepay, collateralBurn, tokenReward
}

// ComputeCloseFactor derives the maximum portion of a borrower's current borrowedValue
// that can currently be repaid in a single liquidate transaction.
//
// closeFactor scales linearly between minimumCloseFactor and 1.0,
// reaching its maximum when borrowedValue has reached a critical value
// between liquidationThreshold to collateralValue.
// This critical value is defined as:
//
//	B = critical borrowedValue
//	C = collateralValue
//	L = liquidationThreshold
//	CLT = completeLiquidationThreshold
//
//	B = L + (C-L) * CLT
//
// closeFactor is zero for borrowers that are not eligible for liquidation,
// i.e. borrowedValue < liquidationThreshold
//
// Finally, if borrowedValue is less than smallLiquidationSize,
// closeFactor will always be 1 as long as the borrower is eligible for liquidation.
func ComputeCloseFactor(
	borrowedValue sdk.Dec,
	collateralValue sdk.Dec,
	liquidationThreshold sdk.Dec,
	smallLiquidationSize sdk.Dec,
	minimumCloseFactor sdk.Dec,
	completeLiquidationThreshold sdk.Dec,
) (closeFactor sdk.Dec) {
	if borrowedValue.LT(liquidationThreshold) {
		// Not eligible for liquidation
		return sdk.ZeroDec()
	}

	if borrowedValue.LTE(smallLiquidationSize) {
		// Small enough borrows should be liquidated completely to reduce dust
		return sdk.OneDec()
	}

	if completeLiquidationThreshold.IsZero() {
		// If close factor is set to unlimited
		return sdk.OneDec()
	}

	// Calculate the borrowed value at which close factor reaches 1.0
	criticalValue := liquidationThreshold.Add(completeLiquidationThreshold.Mul(collateralValue.Sub(liquidationThreshold)))

	closeFactor = Interpolate(
		borrowedValue,        // x
		liquidationThreshold, // xMin
		minimumCloseFactor,   // yMin
		criticalValue,        // xMax
		sdk.OneDec(),         // yMax
	)

	if closeFactor.GTE(sdk.OneDec()) {
		closeFactor = sdk.OneDec()
	}
	if closeFactor.IsNegative() {
		closeFactor = sdk.ZeroDec()
	}

	return closeFactor
}
