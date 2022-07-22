package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ComputeLiquidation takes the conditions preceding a liquidation and outputs the amounts
// of base token that should be repaid, collateral uToken burned, and reward token allocated
// as a result of the transaction, after accounting for limiting factors with as little
// rounding as possible. Inputs are as follows:
// - availableRepay: The lowest (in repay denom) of either liquidator balance, max repayment, or borrowed amount.
// - availableCollateral: The amount of the reward uToken denom which borrower has as collateral
// - availableReward: The amount of unreserved reward tokens in the module balance
// - repayTokenPrice: The oracle price of the base repayment denom
// - rewardTokenPrice: The oracle price of the base reward denom
// - uTokenExchangeRate: The uToken exchange rate from collateral uToken denom to reward base denom
// - liquidationIncentive: The liquidation incentive of the token reward denomination
// - closeFactor: The dynamic close factor computed from the borrower's borrowed value and liquidation threshold
// - borrowedValue: The borrower's borrowed value in USD
func ComputeLiquidation(
	availableRepay sdk.Int,
	availableCollateral sdk.Int,
	availableReward sdk.Int,
	repayTokenPrice sdk.Dec,
	rewardTokenPrice sdk.Dec,
	uTokenExchangeRate sdk.Dec,
	liquidationIncentive sdk.Dec,
	closeFactor sdk.Dec,
	borrowedValue sdk.Dec,
) (tokenRepay sdk.Int, collateralBurn sdk.Int, tokenReward sdk.Int) {
	// Prevent division by zero
	if uTokenExchangeRate.IsZero() || rewardTokenPrice.IsZero() || repayTokenPrice.IsZero() {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	}

	// Start with the maximum possible repayment amount, as a decimal
	repayDec := availableRepay.ToDec()
	// Determine the base rewardDec amount that would result from maximum repayment
	rewardDec := repayDec.Mul(repayTokenPrice).Mul(sdk.OneDec().Add(liquidationIncentive)).Quo(rewardTokenPrice)
	// Determine the collateralDec burn amount that corresponds to base reward amount
	collateralDec := rewardDec.Quo(uTokenExchangeRate)

	// Catch no-ops early
	if repayDec.IsZero() || rewardDec.IsZero() || collateralDec.IsZero() || closeFactor.IsZero() || borrowedValue.IsZero() {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	}

	// We will track limiting factors by the ratio by which the max repayment would need to be reduced to comply
	ratio := sdk.OneDec()
	// Repaid value cannot exceed borrowed value times close factor
	ratio = sdk.MinDec(ratio,
		borrowedValue.Mul(closeFactor).Quo(repayDec.Mul(repayTokenPrice)),
	)
	// Collateral burned cannot exceed borrower's collateral
	ratio = sdk.MinDec(ratio,
		availableCollateral.ToDec().Quo(collateralDec),
	)
	// Base token reward cannot exceed available unreserved module balance
	ratio = sdk.MinDec(ratio,
		availableReward.ToDec().Quo(rewardDec),
	)
	// Catch edge cases
	ratio = sdk.MaxDec(ratio, sdk.ZeroDec())

	// Reduce all three values by the most severe limiting factor encountered
	repayDec = repayDec.Mul(ratio)
	collateralDec = collateralDec.Mul(ratio)
	rewardDec = rewardDec.Mul(ratio)

	// No rounding has occurred yet. In dust scenarios, the limiting factor will be a clean integer
	// and the other two will need rounding. Otherwise, all three outputs will need rounding.

	// REQUIREMENTS:
	// - Prevent REPAY VALUE > COLLATERAL VALUE rounding attacks
	// - Prevent REWARD VALUE > COLLATERAL VALUE rounding attacks
	// - Eliminate COLLATERAL DUST (required for bad debt flag)

	// TODO: Satisfy these requirements

	// The amount of borrowed token the liquidator will repay is rounded up
	tokenRepay = repayDec.Ceil().RoundInt()
	// The amount of collateral uToken the borrower will lose is rounded down
	collateralBurn = collateralDec.TruncateInt()

	// Liquidator base token reward is derived from collateral burn then rounded down
	tokenReward = collateralBurn.ToDec().Mul(uTokenExchangeRate).TruncateInt()
	return tokenRepay, collateralBurn, tokenReward
}

// ComputeCloseFactor uses a borrower's borrowed value and liquidation threshold and
// some leverage module parameters to derive a dynamic close factor for liquidation.
func ComputeCloseFactor(
	borrowedValue sdk.Dec,
	liquidationThreshold sdk.Dec,
	smallLiquidationSize sdk.Dec,
	minimumCloseFactor sdk.Dec,
	completeLiquidationThreshold sdk.Dec,
) (closeFactor sdk.Dec) {
	if !liquidationThreshold.IsPositive() || borrowedValue.LTE(liquidationThreshold) {
		// Not eligible for liquidation
		return sdk.ZeroDec()
	}

	if borrowedValue.LTE(smallLiquidationSize) {
		// Small enough borrows should be liquidated completely to reduce dust
		return sdk.OneDec()
	}

	if completeLiquidationThreshold.IsZero() {
		// If close factor is set to unlimited by global params
		return sdk.OneDec()
	}

	// outside of special cases, close factor scales linearly between MinimumCloseFactor and 1.0,
	// reaching max value when (borrowed / threshold) = 1 + CompleteLiquidationThreshold
	closeFactor = Interpolate(
		borrowedValue.Quo(liquidationThreshold).Sub(sdk.OneDec()), // x
		sdk.ZeroDec(),                // xMin
		minimumCloseFactor,           // yMin
		completeLiquidationThreshold, // xMax
		sdk.OneDec(),                 // yMax
	)
	if closeFactor.GTE(sdk.OneDec()) {
		closeFactor = sdk.OneDec()
	}
	if closeFactor.IsNegative() {
		closeFactor = sdk.ZeroDec()
	}

	return closeFactor
}
