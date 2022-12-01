package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/leverage/types"
)

// maxCollateralFromShare calculates the maximum amount of collateral a utoken denom
// is allowed to have, taking into account its associated token's MaxCollateralShare
// under current market conditions
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

	// get USD collateral value for all other denoms
	otherDenomsValue, err := k.CalculateCollateralValue(ctx, systemCollateral.Sub(thisDenomCollateral))
	if err != nil {
		return sdk.ZeroInt(), err
	}

	// determine the max USD value this uToken's collateral is allowed to have by MaxCollateralShare
	maxValue := otherDenomsValue.Quo(sdk.OneDec().Sub(token.MaxCollateralShare)).Mul(token.MaxCollateralShare)

	// determine the amount of uTokens which would be required to reach maxValue
	tokenDenom := types.ToTokenDenom(denom)
	tokenPrice, err := k.TokenBasePrice(ctx, tokenDenom)
	if err != nil {
		return sdk.ZeroInt(), err
	}
	uTokenExchangeRate := k.DeriveExchangeRate(ctx, tokenDenom)

	// in the case of a small token price smaller than the smallest sdk.Dec (10^-18),
	// this maxCollateralAmount will use the price of 10^-18 and thus derive a lower
	// (more cautious) limit than a precise price would produce
	maxCollateralAmount := maxValue.Quo(tokenPrice).Quo(uTokenExchangeRate).TruncateInt()

	// return the computed maximum or the current uToken supply, whichever is smaller
	return sdk.MinInt(k.GetUTokenSupply(ctx, denom).Amount, maxCollateralAmount), nil
}
