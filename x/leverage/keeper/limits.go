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

	position, err := k.GetAccountPosition(ctx, addr, false)
	if nonOracleError(err) {
		// non-oracle errors fail the transaction (or query)
		return sdk.Coin{}, sdk.Coin{}, err
	}
	if err != nil {
		// oracle errors cause max withdraw to only be wallet uTokens
		withdrawAmount := sdk.MinInt(walletUtokens, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), sdk.NewCoin(uDenom, withdrawAmount), nil
	}
	maxWithdrawValue := position.MaxWithdraw(denom)

	maxWithdraw := coin.Zero(uDenom)
	if position.IsHealthy() && !position.HasCollateral(denom) {
		// if after max withdraw, the position has no more collateral of the requested denom
		// but is still under its borrow limit, then withdraw everything.
		// this works with missing collateral price

		// TODO: since max withdraw no longer mutates, this needs refactor

		maxWithdraw = k.GetCollateral(ctx, addr, uDenom)
	} else {
		// for partial withdrawal, must have collateral price to withdraw anything more than wallet uTokens
		maxWithdraw, err = k.UTokenWithValue(ctx, uDenom, maxWithdrawValue, types.PriceModeLow)
		if nonOracleError(err) {
			// non-oracle errors fail the transaction (or query)
			return sdk.Coin{}, sdk.Coin{}, err
		}
		if err != nil {
			// oracle errors cause max withdraw to only be wallet uTokens
			withdrawAmount := sdk.MinInt(walletUtokens, availableUTokens.Amount)
			return sdk.NewCoin(uDenom, withdrawAmount), sdk.NewCoin(uDenom, withdrawAmount), nil
		}
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
// respecting the MinCollateralLiquidity set for given Token.
func (k Keeper) ModuleAvailableLiquidity(ctx sdk.Context, denom string) (sdkmath.Int, error) {
	// Get module liquidity for the Token
	liquidity := k.AvailableLiquidity(ctx, denom)

	// Get module collateral for the associated uToken
	totalCollateral := k.GetTotalCollateral(ctx, coin.ToUTokenDenom(denom))
	totalTokenCollateral, err := k.ToTokens(ctx, sdk.NewCoins(totalCollateral))
	if err != nil {
		return sdkmath.Int{}, err
	}

	// Get min_collateral_liquidity for the denom
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdkmath.Int{}, err
	}
	minCollateralLiquidity := token.MinCollateralLiquidity

	// The formula to calculate the module_available_liquidity is as follows:
	//
	// 	min_collateral_liquidity = (module_liquidity - module_available_liquidity) / module_collateral
	// 	module_available_liquidity = module_liquidity - min_collateral_liquidity * module_collateral
	moduleAvailableLiquidity := sdk.NewDecFromInt(liquidity).Sub(
		minCollateralLiquidity.MulInt(totalTokenCollateral.AmountOf(denom)),
	)

	return sdk.MaxInt(moduleAvailableLiquidity.TruncateInt(), sdk.ZeroInt()), nil
}
