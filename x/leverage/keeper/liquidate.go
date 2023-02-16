package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/leverage/types"
)

// getLiquidationAmounts takes a repayment and reward denom proposed by a liquidator and calculates
// the actual repayment amount a target address is eligible for, and the corresponding collateral
// to liquidate and equivalent base rewards to send to the liquidator.
func (k Keeper) getLiquidationAmounts(
	ctx sdk.Context,
	liquidatorAddr,
	targetAddr sdk.AccAddress,
	requestedRepay sdk.Coin,
	rewardDenom string,
	directLiquidation bool,
) (tokenRepay sdk.Coin, collateralLiquidate sdk.Coin, tokenReward sdk.Coin, err error) {
	repayDenom := requestedRepay.Denom
	collateralDenom := types.ToUTokenDenom(rewardDenom)

	// get relevant liquidator, borrower, and module balances
	borrowerCollateral := k.GetBorrowerCollateral(ctx, targetAddr)
	totalBorrowed := k.GetBorrowerBorrows(ctx, targetAddr)
	availableRepay := k.bankKeeper.SpendableCoins(ctx, liquidatorAddr).AmountOf(repayDenom)
	repayDenomBorrowed := sdk.NewCoin(repayDenom, totalBorrowed.AmountOf(repayDenom))

	// calculate borrower health in USD values, using spot prices only (no historic)
	// borrowed value will skip borrowed tokens with unknown oracle prices, treating them as zero value
	borrowedValue, err := k.VisibleTokenValue(ctx, totalBorrowed, types.PriceModeSpot)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	collateralValue, err := k.CalculateCollateralValue(ctx, borrowerCollateral)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	liquidationThreshold, err := k.CalculateLiquidationThreshold(ctx, borrowerCollateral)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	if borrowedValue.LT(liquidationThreshold) {
		// borrower is healthy and cannot be liquidated
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, types.ErrLiquidationIneligible
	}
	repayDenomBorrowedValue, err := k.TokenValue(ctx, repayDenomBorrowed, types.PriceModeSpot)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
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
		collateralValue,
		liquidationThreshold,
		params.SmallLiquidationSize,
		params.MinimumCloseFactor,
		params.CompleteLiquidationThreshold,
	)
	// maximum USD value that can be repaid
	maxRepayValue := borrowedValue.Mul(closeFactor)
	// determine fraction of borrowed repayDenom which can be repaid after close factor
	maxRepayAfterCloseFactor := totalBorrowed.AmountOf(repayDenom)
	if maxRepayValue.LT(repayDenomBorrowedValue) {
		maxRepayRatio := maxRepayValue.Quo(repayDenomBorrowedValue)
		maxRepayAfterCloseFactor = maxRepayRatio.MulInt(totalBorrowed.AmountOf(repayDenom)).RoundInt()
	}

	// get precise (less rounding at high exponent) price ratio
	priceRatio, err := k.PriceRatio(ctx, repayDenom, rewardDenom, types.PriceModeSpot)
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

	// max repayment amount is limited by a number of factors
	maxRepay := requestedRepay.Amount                                   // maximum allowed by liquidator
	maxRepay = sdk.MinInt(maxRepay, availableRepay)                     // liquidator account balance
	maxRepay = sdk.MinInt(maxRepay, totalBorrowed.AmountOf(repayDenom)) // borrower position
	maxRepay = sdk.MinInt(maxRepay, maxRepayAfterCloseFactor)           // close factor

	// compute final liquidation amounts
	repay, burn, reward := ComputeLiquidation(
		maxRepay,
		borrowerCollateral.AmountOf(collateralDenom),
		k.AvailableLiquidity(ctx, rewardDenom),
		priceRatio,
		exchangeRate,
		liqudationIncentive,
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
// - priceRatio: The ratio of repayPrice / rewardPrice, which is used when computing rewards
// - uTokenExchangeRate: The uToken exchange rate from collateral uToken denom to reward base denom
// - liquidationIncentive: The liquidation incentive of the token reward denomination
func ComputeLiquidation(
	availableRepay,
	availableCollateral,
	availableReward sdkmath.Int,
	priceRatio,
	uTokenExchangeRate,
	liquidationIncentive sdk.Dec,
) (tokenRepay sdkmath.Int, collateralBurn sdkmath.Int, tokenReward sdkmath.Int) {
	// Prevent division by zero
	if uTokenExchangeRate.IsZero() || priceRatio.IsZero() {
		return sdkmath.ZeroInt(), sdkmath.ZeroInt(), sdkmath.ZeroInt()
	}

	// Start with the maximum possible repayment amount, as a decimal
	maxRepay := toDec(availableRepay)
	// Determine the base maxReward amount that would result from maximum repayment

	maxReward := maxRepay.Mul(priceRatio).Mul(sdk.OneDec().Add(liquidationIncentive))
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
