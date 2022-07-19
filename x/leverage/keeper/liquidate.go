package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// reduceLiquidation takes the conditions preceding a liquidation and outputs the amounts
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
func reduceLiquidation(
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
	// Start with the maximum possible repayment amount, as a decimal
	repay := availableRepay.ToDec()
	// Determine the base reward amount that would result from maximum repayment
	reward := repay.Mul(repayTokenPrice).Mul(sdk.OneDec().Add(liquidationIncentive)).Quo(rewardTokenPrice)
	// Determine the collateral burn amount that corresponds to base reward amount
	collateral := reward.Quo(uTokenExchangeRate)

	// We will track limiting factors by the ratio by which the max repayment would need to be reduced to comply
	ratio := sdk.OneDec()
	// Repaid value cannot exceed borrowed value times close factor
	ratio = sdk.MinDec(ratio,
		borrowedValue.Mul(closeFactor).Quo(repay.Mul(repayTokenPrice)),
	)
	// Collateral burned cannot exceed borrower's collateral
	ratio = sdk.MinDec(ratio,
		availableCollateral.ToDec().Quo(collateral),
	)
	// Base token reward cannot exceed available unreserved module balance
	ratio = sdk.MinDec(ratio,
		availableReward.ToDec().Quo(reward),
	)
	// Catch edge cases
	ratio = sdk.MaxDec(ratio, sdk.ZeroDec())

	// Reduce all three values by the most severe limiting factor encountered
	repay = repay.Mul(ratio)
	collateral = collateral.Mul(ratio)
	reward = reward.Mul(ratio)

	// The amount of borrowed token the liquidator will repay is rounded up
	tokenRepay = repay.Ceil().RoundInt()
	// The amount of collateral uToken the borrower will lose is rounded up
	collateralBurn = collateral.Ceil().RoundInt()
	// The amount of reward token the liquidator will receive is rounded down
	tokenReward = reward.TruncateInt()

	return tokenRepay, collateralBurn, tokenReward
}

// computeLiquidation takes a repayment and reward denom proposed by a liquidator and calculates
// the actual repayment amount a target address is eligible for, and the corresponding collateral
// to burn and rewards to return to the liquidator.
func (k Keeper) computeLiquidation(
	ctx sdk.Context,
	liquidatorAddr sdk.AccAddress,
	targetAddr sdk.AccAddress,
	maxRepay sdk.Coin,
	rewardDenom string,
) (tokenRepay sdk.Coin, collateralBurn sdk.Coin, tokenReward sdk.Coin, err error) {
	repayDenom := maxRepay.Denom
	collateralDenom := k.FromTokenToUTokenDenom(ctx, rewardDenom)

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

	// get dynamic close factor and liquidation incentive
	liquidationIncentive, closeFactor, err := k.liquidationParams(ctx, rewardDenom, borrowedValue, liquidationThreshold)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

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

	// compute final liquidation amounts
	repay, burn, reward := reduceLiquidation(
		sdk.MinInt(sdk.MinInt(availableRepay, maxRepay.Amount), totalBorrowed.AmountOf(repayDenom)),
		borrowerCollateral.AmountOf(collateralDenom),
		k.ModuleBalance(ctx, rewardDenom).Sub(k.GetReserveAmount(ctx, rewardDenom)),
		repayTokenPrice,
		rewardTokenPrice,
		exchangeRate,
		liquidationIncentive,
		closeFactor,
		borrowedValue,
	)

	return sdk.NewCoin(repayDenom, repay), sdk.NewCoin(collateralDenom, burn), sdk.NewCoin(rewardDenom, reward), nil
}

// liquidationParams computes dynamic liquidation parameters based on a collateral reward
// denomination, and a borrower's borrowed value and liquidation threshold. Returns
// liquidationIncentive (the ratio of bonus collateral awarded during Liquidate transactions,
// and closeFactor (the fraction of a borrower's total borrowed value that can be repaid
// by a liquidator in a single liquidation event.)
func (k Keeper) liquidationParams(
	ctx sdk.Context,
	rewardDenom string,
	borrowedValue sdk.Dec,
	liquidationThreshold sdk.Dec,
) (sdk.Dec, sdk.Dec, error) {
	if borrowedValue.IsNegative() {
		return sdk.ZeroDec(), sdk.ZeroDec(), sdkerrors.Wrap(types.ErrBadValue, borrowedValue.String())
	}
	if liquidationThreshold.IsNegative() {
		return sdk.ZeroDec(), sdk.ZeroDec(), sdkerrors.Wrap(types.ErrBadValue, liquidationThreshold.String())
	}

	if liquidationThreshold.GTE(borrowedValue) {
		return sdk.ZeroDec(), sdk.ZeroDec(), types.ErrLiquidationIneligible.Wrapf(
			"borrowed value %s is below the liquidation threshold %s",
			borrowedValue, liquidationThreshold)
	}

	ts, err := k.GetTokenSettings(ctx, rewardDenom)
	if err != nil {
		return sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// special case: If liquidation threshold is zero, close factor is always 1
	if liquidationThreshold.IsZero() {
		return ts.LiquidationIncentive, sdk.OneDec(), nil
	}

	params := k.GetParams(ctx)

	// special case: If borrowed value is less than small liquidation size,
	// close factor is always 1
	if borrowedValue.LTE(params.SmallLiquidationSize) {
		return ts.LiquidationIncentive, sdk.OneDec(), nil
	}

	// special case: If complete liquidation threshold is zero, close factor is always 1
	if params.CompleteLiquidationThreshold.IsZero() {
		return ts.LiquidationIncentive, sdk.OneDec(), nil
	}

	// outside of special cases, close factor scales linearly between MinimumCloseFactor and 1.0,
	// reaching max value when (borrowed / threshold) = 1 + CompleteLiquidationThreshold
	var closeFactor sdk.Dec
	closeFactor = Interpolate(
		borrowedValue.Quo(liquidationThreshold).Sub(sdk.OneDec()), // x
		sdk.ZeroDec(),                       // xMin
		params.MinimumCloseFactor,           // yMin
		params.CompleteLiquidationThreshold, // xMax
		sdk.OneDec(),                        // yMax
	)
	if closeFactor.GTE(sdk.OneDec()) {
		closeFactor = sdk.OneDec()
	}
	if closeFactor.IsNegative() {
		closeFactor = sdk.ZeroDec()
	}

	return ts.LiquidationIncentive, closeFactor, nil
}
