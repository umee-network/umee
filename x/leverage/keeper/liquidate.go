package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

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
	closeFactor := types.ComputeCloseFactor(
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

	// compute final liquidation amounts
	repay, burn, reward := types.ComputeLiquidation(
		sdk.MinInt(sdk.MinInt(availableRepay, maxRepay.Amount), totalBorrowed.AmountOf(repayDenom)),
		borrowerCollateral.AmountOf(collateralDenom),
		k.ModuleBalance(ctx, rewardDenom).Sub(k.GetReserveAmount(ctx, rewardDenom)),
		repayTokenPrice,
		rewardTokenPrice,
		exchangeRate,
		ts.LiquidationIncentive,
		closeFactor,
		borrowedValue,
	)

	return sdk.NewCoin(repayDenom, repay), sdk.NewCoin(collateralDenom, burn), sdk.NewCoin(rewardDenom, reward), nil
}
