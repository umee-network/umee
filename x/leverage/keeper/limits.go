package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/leverage/types"
)

// userMaxWithdraw calculates the maximum amount of uTokens an account can currently withdraw and the amount of
// these uTokens which are non-collateral. Input denom should be a base token. If oracle prices are missing for
// some of the borrower's collateral (other than the denom being withdrawn), computes the maximum safe withdraw
// allowed by only the collateral whose prices are known.
func (k *Keeper) newMaxWithdraw(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, sdk.Coin, error) {
	uDenom := types.ToUTokenDenom(denom)
	position, err := k.getAccountPosition(ctx, addr, false)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	maxWithdrawValue := position.MaxWithdraw(denom)
	maxWithdraw, err := k.UTokenWithValue(ctx, uDenom, maxWithdrawValue, types.PriceModeLow)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	collateralUTokens, err := k.ExchangeToken(ctx, maxWithdraw)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// always withdraw non-collateral uTokens
	walletUtokens := sdk.NewCoin(
		uDenom,
		k.bankKeeper.SpendableCoins(ctx, addr).AmountOf(uDenom),
	)

	// apply incentive module limit
	collateralUTokens.Amount = sdk.MinInt(collateralUTokens.Amount, k.unbondedCollateral(ctx, addr, uDenom).Amount)
	return collateralUTokens.Add(walletUtokens), walletUtokens, nil
}

// userMaxBorrow calculates the maximum amount of a given token an account can currently borrow.
// input denom should be a base token. If oracle prices are missing for some of the borrower's
// collateral, computes the maximum safe borrow allowed by only the collateral whose prices are known.
func (k *Keeper) newMaxBorrow(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	if types.HasUTokenPrefix(denom) {
		return sdk.Coin{}, types.ErrUToken
	}

	position, err := k.getAccountPosition(ctx, addr, false)
	if err != nil {
		return sdk.Coin{}, err
	}
	maxBorrowValue := position.MaxBorrow(denom)
	maxBorrow, err := k.TokenWithValue(ctx, denom, maxBorrowValue, types.PriceModeHigh)
	if err != nil {
		return sdk.Coin{}, err
	}

	// also cap borrow amount at available liquidity
	maxBorrow.Amount = sdk.MinInt(maxBorrow.Amount, k.AvailableLiquidity(ctx, denom))
	return maxBorrow, nil
}

// maxCollateralFromShare calculates the maximum amount of collateral a utoken denom
// is allowed to have, taking into account its associated token's MaxCollateralShare
// under current market conditions. If any collateral denoms other than this are missing
// oracle prices, calculates a (lower) maximum amount using the collateral with known prices.
func (k *Keeper) maxCollateralFromShare(ctx sdk.Context, denom string) (sdkmath.Int, error) {
	token, err := k.GetTokenSettings(ctx, types.ToTokenDenom(denom))
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
	udenom := types.ToUTokenDenom(denom)
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
	totalCollateral := k.GetTotalCollateral(ctx, types.ToUTokenDenom(denom))
	totalTokenCollateral, err := k.ExchangeUTokens(ctx, sdk.NewCoins(totalCollateral))
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
