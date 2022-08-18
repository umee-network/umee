package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// getLiquidationAmounts takes a repayment and reward denom proposed by a liquidator and calculates
// the actual repayment amount a target address is eligible for, and the corresponding collateral
// to liquidate and equivalent base rewards to send to the liquidator.
func (k Keeper) getLiquidationAmounts(
	ctx sdk.Context,
	liquidatorAddr,
	targetAddr sdk.AccAddress,
	maxRepay sdk.Coin,
	rewardDenom string,
	directLiquidation bool,
) (tokenRepay sdk.Coin, collateralLiquidate sdk.Coin, tokenReward sdk.Coin, err error) {
	repayDenom := maxRepay.Denom
	collateralDenom := types.ToUTokenDenom(rewardDenom)

	// get relevant liquidator, borrower, and module balances
	borrowerCollateral := k.GetBorrowerCollateral(ctx, targetAddr)
	totalBorrowed := k.GetBorrowerBorrows(ctx, targetAddr)
	availableRepay := k.bankKeeper.SpendableCoins(ctx, liquidatorAddr).AmountOf(repayDenom)

	// calculate borrower health in USD values
	borrowedValue, err := k.TotalTokenValue(ctx, totalBorrowed)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	liquidationThreshold, err := k.CalculateLiquidationThreshold(ctx, borrowerCollateral)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	if liquidationThreshold.GTE(borrowedValue) {
		// borrower is healthy and cannot be liquidated
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, types.ErrLiquidationIneligible
	}

	// get liquidation incentive
	ts, err := k.GetTokenSettings(ctx, rewardDenom)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// get dynamic close factor
	params := k.GetParams(ctx)
	closeFactor := ComputeCloseFactor(
		borrowedValue,
		liquidationThreshold,
		params.SmallLiquidationSize,
		params.MinimumCloseFactor,
		params.CompleteLiquidationThreshold,
	)

	// get oracle prices for the reward and repay denoms
	repayTokenPrice, err := k.TokenPrice(ctx, repayDenom)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	rewardTokenPrice, err := k.TokenPrice(ctx, rewardDenom)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
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

	// compute final liquidation amounts
	repay, burn, reward := ComputeLiquidation(
		sdk.MinInt(sdk.MinInt(availableRepay, maxRepay.Amount), totalBorrowed.AmountOf(repayDenom)),
		borrowerCollateral.AmountOf(collateralDenom),
		k.ModuleBalance(ctx, rewardDenom).Sub(k.GetReserveAmount(ctx, rewardDenom)),
		repayTokenPrice,
		rewardTokenPrice,
		exchangeRate,
		liqudationIncentive,
		closeFactor,
		borrowedValue,
	)

	return sdk.NewCoin(repayDenom, repay), sdk.NewCoin(collateralDenom, burn), sdk.NewCoin(rewardDenom, reward), nil
}

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
	availableRepay,
	availableCollateral,
	availableReward sdkmath.Int,
	repayTokenPrice,
	rewardTokenPrice,
	uTokenExchangeRate,
	liquidationIncentive,
	closeFactor,
	borrowedValue sdk.Dec,
) (tokenRepay sdkmath.Int, collateralBurn sdkmath.Int, tokenReward sdkmath.Int) {
	// Prevent division by zero
	if uTokenExchangeRate.IsZero() || rewardTokenPrice.IsZero() || repayTokenPrice.IsZero() {
		return sdkmath.ZeroInt(), sdkmath.ZeroInt(), sdkmath.ZeroInt()
	}

	// Start with the maximum possible repayment amount, as a decimal
	maxRepay := toDec(availableRepay)
	// Determine the base maxReward amount that would result from maximum repayment
	maxReward := maxRepay.Mul(repayTokenPrice).Mul(sdk.OneDec().Add(liquidationIncentive)).Quo(rewardTokenPrice)
	// Determine the maxCollateral burn amount that corresponds to base reward amount
	maxCollateral := maxReward.Quo(uTokenExchangeRate)

	// Catch no-ops early
	if maxRepay.IsZero() ||
		maxReward.IsZero() ||
		maxCollateral.IsZero() ||
		closeFactor.IsZero() ||
		borrowedValue.IsZero() {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	}

	// We will track limiting factors by the ratio by which the max repayment would need to be reduced to comply
	ratio := sdk.OneDec()
	// Repaid value cannot exceed borrowed value times close factor
	ratio = sdk.MinDec(ratio,
		borrowedValue.Mul(closeFactor).Quo(maxRepay.Mul(repayTokenPrice)),
	)
	// Collateral burned cannot exceed borrower's collateral
	ratio = sdk.MinDec(ratio,
		toDec(availableCollateral).Quo(maxCollateral),
	)
	// Base token reward cannot exceed available unreserved module balance
	ratio = sdk.MinDec(ratio,
		toDec(availableReward).Quo(maxReward),
	)
	// Catch edge cases
	ratio = sdk.MaxDec(ratio, sdk.ZeroDec())

	// Reduce repay and collateral limits by the most severe limiting factor encountered
	maxRepay = maxRepay.Mul(ratio)
	maxCollateral = maxCollateral.Mul(ratio)

	// No rounding has occurred yet, but both values are now within the
	// limits defined by available balances and module parameters.

	// First, the amount of borrowed token the liquidator must repay is rounded up.
	// This is a slight disadvantage to the liquidator in favor of the borrower and
	// the module. It also ensures borrow dust is always eliminated when encountered.
	tokenRepay = maxRepay.Ceil().RoundInt()

	// Next, the amount of collateral uToken the borrower will lose is rounded down.
	// This is favors the borrower over the liquidator, and also protects the module.
	collateralBurn = maxCollateral.TruncateInt()

	// One danger to rounding collateral burn down is that of collateral dust. This
	// can be considered in two scenarios:
	// 1) If collateral was the limiting factor above, then it will have already been
	// an integer amount and truncating is a no-op.
	// 2) If collateral was not the limiting factor, then there will be a non-dust
	// quantity left over anyway.

	// Finally, the base token reward amount is derived directly from the collateral
	// to burn. This will round down identically to MsgWithdraw, favoring the module
	// over the liquidator.
	tokenReward = toDec(collateralBurn).Mul(uTokenExchangeRate).TruncateInt()

	return tokenRepay, collateralBurn, tokenReward
}

// ComputeCloseFactor derives the maximum portion of a borrower's current
// borrowed value can currently be repaid in a single liquidate transaction.
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
