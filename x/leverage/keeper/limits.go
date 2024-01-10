package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

// userMaxWithdraw calculates the maximum amount of uTokens an account can currently withdraw and the amount of
// these uTokens which are non-collateral. Input denom should be a base token. If oracle prices are missing for
// some of the borrower's collateral (other than the denom being withdrawn), computes the maximum safe withdraw
// allowed by only the collateral whose prices are known.
func (k *Keeper) userMaxWithdraw(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, sdk.Coin, error) {
	uDenom := coin.ToUTokenDenom(denom)
	availableTokens := sdk.NewCoin(denom, k.AvailableLiquidity(ctx, denom))
	availableUTokens, err := k.ToUToken(ctx, availableTokens)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	walletUtokens := k.bankKeeper.SpendableCoins(ctx, addr).AmountOf(uDenom)
	unbondedCollateral := k.unbondedCollateral(ctx, addr, uDenom)

	maxWithdraw := coin.Zero(uDenom)
	position, err := k.GetAccountPosition(ctx, addr, false)
	if err == nil {
		maxWithdrawValue, fullWithdrawal := position.MaxWithdraw(denom)
		if fullWithdrawal {
			maxWithdraw = k.GetCollateral(ctx, addr, uDenom)
		} else {
			maxWithdraw, err = k.UTokenWithValue(ctx, uDenom, maxWithdrawValue, types.PriceModeLow)
		}
	}
	if nonOracleError(err) {
		// non-oracle errors fail the transaction (or query)
		return sdk.Coin{}, sdk.Coin{}, err
	}
	if err != nil {
		// oracle errors cause max withdraw to only be wallet uTokens
		withdrawAmount := sdk.MinInt(walletUtokens, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), sdk.NewCoin(uDenom, withdrawAmount), nil
	}

	// find the minimum of max withdraw (from positions) or unbonded collateral (incentive module)
	if unbondedCollateral.Amount.LT(maxWithdraw.Amount) {
		maxWithdraw = unbondedCollateral
	}

	// add wallet uTokens to the unused amount from collateral
	maxWithdraw.Amount = maxWithdraw.Amount.Add(walletUtokens)
	// reduce amount to withdraw if it exceeds available liquidity
	maxWithdraw.Amount = sdk.MinInt(maxWithdraw.Amount, availableUTokens.Amount)
	return maxWithdraw, sdk.NewCoin(uDenom, walletUtokens), nil
}

// userMaxBorrow calculates the maximum amount of a given token an account can currently borrow.
// input denom should be a base token. If oracle prices are missing for some of the borrower's
// collateral, computes the maximum safe borrow allowed by only the collateral whose prices are known.
func (k *Keeper) userMaxBorrow(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	if coin.HasUTokenPrefix(denom) {
		return sdk.Coin{}, types.ErrUToken
	}

	availableTokens := k.AvailableLiquidity(ctx, denom)

	position, err := k.GetAccountPosition(ctx, addr, false)
	if nonOracleError(err) {
		// non-oracle errors fail the transaction (or query)
		return sdk.Coin{}, err
	}
	if err != nil {
		// oracle errors cause max borrow to be zero
		return sdk.NewCoin(denom, sdk.ZeroInt()), nil
	}

	maxBorrowValue := position.MaxBorrow(denom)
	maxBorrow, err := k.TokenWithValue(ctx, denom, maxBorrowValue, types.PriceModeHigh)
	if nonOracleError(err) {
		// non-oracle errors fail the transaction (or query)
		return sdk.Coin{}, err
	}
	if err != nil {
		// oracle errors cause max borrow to be zero
		return sdk.NewCoin(denom, sdk.ZeroInt()), nil
	}

	// also cap borrow amount at available liquidity
	maxBorrow.Amount = sdk.MinInt(maxBorrow.Amount, availableTokens)

	return maxBorrow, nil
}

// maxCollateralFromShare calculates the maximum amount of collateral a utoken denom
// is allowed to have, taking into account its associated token's MaxCollateralShare
// under current market conditions. If any collateral denoms other than this are missing
// oracle prices, calculates a (lower) maximum amount using the collateral with known prices.
func (k *Keeper) maxCollateralFromShare(ctx sdk.Context, denom string) (sdkmath.Int, error) {
	token, err := k.GetTokenSettings(ctx, coin.StripUTokenDenom(denom))
	if err != nil {
		return sdk.ZeroInt(), err
	}

	// if a token's max collateral share is zero, max collateral is zero
	if token.MaxCollateralShare.LTE(sdk.ZeroDec()) {
		return sdk.ZeroInt(), nil
	}

	// if a token's max collateral share is 100%, max collateral is its uToken supply
	if token.MaxCollateralShare.GTE(sdk.OneDec()) {
		return k.GetUTokenSupply(ctx, denom).Amount, nil
	}

	// if a token's max collateral share is less than 100%, additional restrictions apply
	systemCollateral := k.GetAllTotalCollateral(ctx)
	thisDenomCollateral := sdk.NewCoin(denom, systemCollateral.AmountOf(denom))

	// get USD collateral value for all other denoms, skipping those which are missing oracle prices
	otherDenomsValue, err := k.VisibleCollateralValue(
		ctx,
		systemCollateral.Sub(thisDenomCollateral),
		types.PriceModeSpot,
	)
	if err != nil {
		return sdk.ZeroInt(), err
	}

	// determine the max USD value this uToken's collateral is allowed to have by MaxCollateralShare
	maxValue := otherDenomsValue.Quo(sdk.OneDec().Sub(token.MaxCollateralShare)).Mul(token.MaxCollateralShare)

	// determine the amount of base tokens which would be required to reach maxValue,
	// using the higher of spot or historic prices
	udenom := coin.ToUTokenDenom(denom)
	maxUTokens, err := k.UTokenWithValue(ctx, udenom, maxValue, types.PriceModeHigh)
	if err != nil {
		return sdk.ZeroInt(), err
	}

	// return the computed maximum or the current uToken supply, whichever is smaller
	return sdk.MinInt(k.GetUTokenSupply(ctx, denom).Amount, maxUTokens.Amount), nil
}

// ModuleAvailableLiquidity calculates the maximum available liquidity of a Token denom from the module can be used,
// respecting the Token's MinCollateralLiquidity as well as the tokens supplied from the meToken module.
func (k Keeper) ModuleAvailableLiquidity(ctx sdk.Context, denom string) (sdkmath.Int, error) {
	// Get unreserved module balance for the Token
	liquidity := k.AvailableLiquidity(ctx, denom)

	// Determine how much liquidity must be kept due to min_collateral_liquidity
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.ZeroInt(), err
	}
	totalCollateral, err := k.ToToken(ctx, k.GetTotalCollateral(ctx, coin.ToUTokenDenom(denom)))
	if err != nil {
		return sdk.ZeroInt(), err
	}
	requiredLiquidityFromCollateral := token.MinCollateralLiquidity.MulInt(totalCollateral.Amount).TruncateInt()

	// Determine how much liquidity must be kept for potential withdrawals by the meToken module
	meTokenBacking, err := k.GetSupplied(ctx, k.meTokenAddr, denom)
	if err != nil {
		return sdk.ZeroInt(), err
	}
	requiredLiquidityFromMeTokens := meTokenBacking.Amount

	// ModuleAvailableLiquidity is the unreserved module balance,
	// minus additional tokens kept for collateral liquidity and metoken liquidity
	return sdk.MaxInt(
		sdk.ZeroInt(),
		liquidity.Sub(requiredLiquidityFromCollateral).Sub(requiredLiquidityFromMeTokens),
	), nil
}

// ModuleMaxWithdraw calculates the maximum available amount of uToken to withdraw from the module given the amount of
// user's spendable tokens. The calculation first finds the maximum amount of non-collateral uTokens the user can
// withdraw up to the amount in their wallet, then determines how much collateral can be withdrawn in addition to that.
// The returned value is the sum of the two values.
func (k Keeper) ModuleMaxWithdraw(ctx sdk.Context, spendableUTokens sdk.Coin, withdrawalAddr sdk.AccAddress) (
	sdkmath.Int,
	error,
) {
	denom := coin.StripUTokenDenom(spendableUTokens.Denom)

	moduleAvailableLiquidity, err := k.ModuleAvailableLiquidity(ctx, denom)
	if err != nil {
		return sdk.ZeroInt(), err
	}

	if spendableUTokens.Amount.GTE(moduleAvailableLiquidity) {
		return moduleAvailableLiquidity, nil
	}

	// Get module collateral for the uDenom
	totalCollateral := k.GetTotalCollateral(ctx, spendableUTokens.Denom)
	totalTokenCollateral, err := k.ToTokens(ctx, sdk.NewCoins(totalCollateral))
	if err != nil {
		return sdk.ZeroInt(), err
	}

	// If after subtracting all the user_spendable_utokens from the module_available_liquidity,
	// the result is higher than the total module_collateral,
	// we can withdraw user_spendable_utokens + module_collateral.
	if moduleAvailableLiquidity.Sub(spendableUTokens.Amount).GTE(totalTokenCollateral.AmountOf(denom)) {
		return spendableUTokens.Amount.Add(totalTokenCollateral.AmountOf(denom)), nil
	}

	// MeToken module supply is fully protected in order to guarantee its availability for redemption.
	var liquidity sdkmath.Int
	if withdrawalAddr.Equals(k.meTokenAddr) {
		liquidity = k.AvailableLiquidity(ctx, denom)
	} else {
		liquidity, err = k.AvailableLiquiditySubMetokenSupply(ctx, denom)
		if err != nil {
			return sdk.ZeroInt(), err
		}
	}

	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.ZeroInt(), err
	}
	minCollateralLiquidity := token.MinCollateralLiquidity

	// At this point we know that there is enough module_available_liquidity to withdraw user_spendable_utokens.
	// Now we need to get the module_available_collateral after withdrawing user_spendable_utokens:
	//
	// min_collateral_liquidity = (module_liquidity - user_spendable_utokens - module_available_collateral)
	//									/ (module_collateral  - module_available_collateral)
	//
	// module_available_collateral = (module_liquidity - user_spendable_utokens - min_collateral_liquidity
	//									* module_collateral) / (1 - min_collateral_liquidity)
	moduleAvailableCollateral := (sdk.NewDecFromInt(liquidity.Sub(spendableUTokens.Amount)).Sub(
		minCollateralLiquidity.MulInt(
			totalTokenCollateral.AmountOf(denom),
		),
	)).Quo(sdk.NewDec(1).Sub(minCollateralLiquidity))

	// Adding (user_spendable_utokens + module_available_collateral) we obtain the max uTokens the account can
	// withdraw from the module.
	return spendableUTokens.Amount.Add(moduleAvailableCollateral.TruncateInt()), nil
}
