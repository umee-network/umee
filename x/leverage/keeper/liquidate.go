package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// liquidationOutcome takes a repayment and reward denom proposed by a liquidator and calculates
// the maximum repayment amounts a target address is eligible for, and the corresponding reward
// amounts using current oracle, params, and available balances. Inputs must be registered tokens.
// Outputs are base token repayment, uToken collateral cost, and base token reward from liquidation.
func (k Keeper) liquidationOutcome(
	ctx sdk.Context,
	liquidatorAddr sdk.AccAddress,
	targetAddr sdk.AccAddress,
	desiredRepay sdk.Coin,
	rewardDenom string,
) (sdk.Coin, sdk.Coin, sdk.Coin, error) {
	// get liquidator's available balance of base asset to repay
	availableRepay := k.bankKeeper.SpendableCoins(ctx, liquidatorAddr).AmountOf(desiredRepay.Denom)

	// get module's available balance of reward base asset
	availableReward := k.ModuleBalance(ctx, rewardDenom).Sub(k.GetReserveAmount(ctx, rewardDenom))

	// examine target address
	collateral := k.GetBorrowerCollateral(ctx, targetAddr)
	borrowed := k.GetBorrowerBorrows(ctx, targetAddr)

	// calculate position health in USD values
	borrowedValue, err := k.TotalTokenValue(ctx, borrowed)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	liquidationThreshold, err := k.CalculateLiquidationThreshold(ctx, collateral)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	if liquidationThreshold.GTE(borrowedValue) {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, types.ErrLiquidationIneligible.Wrapf(
			"%s borrowed value %s is below the liquidation threshold %s",
			targetAddr, borrowedValue, liquidationThreshold)
	}

	// get dynamic close factor and liquidation incentive
	liquidationIncentive, closeFactor, err := k.liquidationParams(ctx, rewardDenom, borrowedValue, liquidationThreshold)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// maximum repayment starts at desiredRepayment but can be lower due to limiting factors
	baseRepay := desiredRepay
	// repayment cannot exceed borrower's borrowed amount of selected denom
	baseRepay.Amount = sdk.MinInt(baseRepay.Amount, borrowed.AmountOf(baseRepay.Denom))
	// repayment cannot exceed liquidator's available balance
	baseRepay.Amount = sdk.MinInt(baseRepay.Amount, availableRepay)
	// repayment USD value cannot exceed borrowed USD value * close factor
	repayValueLimit := borrowedValue.Mul(closeFactor)
	repayValue, err := k.TokenValue(ctx, baseRepay)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	// if repayValue > repayValueLimit
	//   maxRepayment *= (repayValueLimit / repayValue)
	ReduceProportionallyDec(repayValueLimit, repayValue, &baseRepay.Amount)
	if baseRepay.IsZero() {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, types.ErrLiquidationRepayZero
	}

	// find the price ratio of repay:reward tokens
	priceRatio, err := k.PriceRatio(ctx, baseRepay.Denom, rewardDenom)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	// determine uToken collateral reward, rounding up
	uReward := sdk.NewCoin(
		k.FromTokenToUTokenDenom(ctx, rewardDenom),
		// uReward = repay * (repayPrice / rewardPrice) * (1  + incentive), rounded up
		baseRepay.Amount.ToDec().Mul(priceRatio).Mul(sdk.OneDec().Add(liquidationIncentive)).Ceil().RoundInt(),
	)

	// uToken reward cannot exceed available collateral
	availableCollateral := collateral.AmountOf(uReward.Denom)
	// if uReward > availableCollateral
	//   uReward = availableCollateral
	//   baseRepay *= (availableCollateral / uReward)
	ReduceProportionally(availableCollateral, uReward.Amount, &uReward.Amount, &baseRepay.Amount)
	if uReward.IsZero() {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, types.ErrLiquidationRewardZero
	}

	// convert uToken reward to base tokens, rounding down
	baseReward, err := k.ExchangeUToken(ctx, uReward)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	// base reward cannot exceed available reward
	// if baseReward > availableReward
	//   baseReward = availableReward
	//   uReward *= (availableReward / baseReward)
	//   baseRepay *= (availableReward / baseReward)
	ReduceProportionally(availableReward, baseReward.Amount, &baseReward.Amount, &uReward.Amount, &baseRepay.Amount)

	return baseRepay, uReward, baseReward, nil
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
