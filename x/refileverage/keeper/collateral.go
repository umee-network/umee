package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/refileverage/types"
)

// unbondedCollateral returns the collateral an account has which is neither bonded nor currently unbonding.
// This collateral is available for immediate decollateralizing or withdrawal.
func (k Keeper) unbondedCollateral(ctx sdk.Context, addr sdk.AccAddress, uDenom string) sdk.Coin {
	collateralAmount := k.GetBorrowerCollateral(ctx, addr).AmountOf(uDenom)
	unavailable := k.bondedCollateral(ctx, addr, uDenom)
	available := sdk.MaxInt(collateralAmount.Sub(unavailable.Amount), sdk.ZeroInt())
	return sdk.NewCoin(uDenom, available)
}

// liquidateCollateral burns uToken collateral and sends the base token reward to the liquidator.
// This occurs during direct liquidation.
func (k Keeper) liquidateCollateral(ctx sdk.Context, borrower, liquidator sdk.AccAddress, uToken, token sdk.Coin,
) error {
	if err := k.burnCollateral(ctx, borrower, uToken); err != nil {
		return err
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidator, sdk.NewCoins(token))
}

// burnCollateral removes some uTokens from an account's collateral and burns them. This occurs
// during direct liquidations and during donateCollateral.
func (k Keeper) burnCollateral(ctx sdk.Context, addr sdk.AccAddress, uToken sdk.Coin) error {
	err := k.setCollateral(ctx, addr, k.GetCollateral(ctx, addr, uToken.Denom).Sub(uToken))
	if err != nil {
		return err
	}
	if err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(uToken)); err != nil {
		return err
	}
	return k.setUTokenSupply(ctx, k.GetUTokenSupply(ctx, uToken.Denom).Sub(uToken))
}

// decollateralize removes fromAddr's uTokens from the module and sends them to toAddr.
// It occurs when decollateralizing uTokens (in which case fromAddr and toAddr are the
// same) as well as during non-direct liquidations, where toAddr is the liquidator.
func (k Keeper) decollateralize(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, uToken sdk.Coin) error {
	err := k.setCollateral(ctx, fromAddr, k.GetCollateral(ctx, fromAddr, uToken.Denom).Sub(uToken))
	if err != nil {
		return err
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toAddr, sdk.NewCoins(uToken))
}

// GetTotalCollateral returns an sdk.Coin representing how much of a given uToken
// the x/refileverage module account currently holds as collateral. Non-uTokens return zero.
func (k Keeper) GetTotalCollateral(ctx sdk.Context, denom string) sdk.Coin {
	if !types.HasUTokenPrefix(denom) {
		// non-uTokens cannot be collateral
		return sdk.Coin{}
	}

	// uTokens in the module account are always from collateral
	return k.ModuleBalance(ctx, denom)
}

// CalculateCollateralValue uses the price oracle to determine the value (in USD) provided by
// collateral sdk.Coins, using each token's uToken exchange rate. Always uses spot price.
// An error is returned if any input coins are not uTokens or if value calculation fails.
func (k Keeper) CalculateCollateralValue(ctx sdk.Context, collateral sdk.Coins) (sdk.Dec, error) {
	total := sdk.ZeroDec()

	for _, coin := range collateral {
		// convert uToken collateral to base assets
		baseAsset, err := k.ExchangeUToken(ctx, coin)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		// get USD value of base assets
		v, err := k.TokenValue(ctx, baseAsset, types.PriceModeSpot)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		// add each collateral coin's weighted value to borrow limit
		total = total.Add(v)
	}

	return total, nil
}

// VisibleCollateralValue uses the price oracle to determine the value (in USD) provided by
// collateral sdk.Coins, using each token's uToken exchange rate. Always uses spot price.
// Unlike CalculateCollateralValue, this function will not return an error if value calculation
// fails on a token - instead, that token will contribute zero value to the total.
func (k Keeper) VisibleCollateralValue(ctx sdk.Context, collateral sdk.Coins) (sdk.Dec, error) {
	total := sdk.ZeroDec()

	for _, coin := range collateral {
		// convert uToken collateral to base assets
		baseAsset, err := k.ExchangeUToken(ctx, coin)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		// get USD value of base assets
		v, err := k.TokenValue(ctx, baseAsset, types.PriceModeSpot)
		if err == nil {
			// for coins that did not error, add their value to the total
			total = total.Add(v)
		}
		if nonOracleError(err) {
			return sdk.ZeroDec(), err
		}
	}

	return total, nil
}

// GetAllTotalCollateral returns total collateral across all uTokens.
func (k Keeper) GetAllTotalCollateral(ctx sdk.Context) sdk.Coins {
	total := sdk.NewCoins()

	tokens := k.GetAllRegisteredTokens(ctx)
	for _, t := range tokens {
		uDenom := types.ToUTokenDenom(t.BaseDenom)
		total = total.Add(k.GetTotalCollateral(ctx, uDenom))
	}
	return total
}

// VisibleCollateralShare calculates the portion of overall collateral (measured in USD value) that a
// given uToken denom represents. If an asset other than the denom requested is missing an oracle
// price, it ignores that asset's contribution to the system's overall collateral, thus potentially
// overestimating the requested denom's collateral share while improving availability.
func (k *Keeper) VisibleCollateralShare(ctx sdk.Context, denom string) (sdk.Dec, error) {
	systemCollateral := k.GetAllTotalCollateral(ctx)
	thisCollateral := sdk.NewCoins(sdk.NewCoin(denom, systemCollateral.AmountOf(denom)))

	// get USD collateral value for all uTokens combined, except those experiencing price outages
	totalValue, err := k.VisibleCollateralValue(ctx, systemCollateral)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	// get USD collateral value for this uToken only
	thisValue, err := k.CalculateCollateralValue(ctx, thisCollateral)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	if !totalValue.IsPositive() {
		return sdk.ZeroDec(), nil
	}
	return thisValue.Quo(totalValue), nil
}

// checkCollateralShare returns an error if a given uToken is above its collateral share
// as calculated using only tokens whose oracle prices exist
func (k *Keeper) checkCollateralShare(ctx sdk.Context, denom string) error {
	token, err := k.GetTokenSettings(ctx, types.ToTokenDenom(denom))
	if err != nil {
		return err
	}

	if token.MaxCollateralShare.Equal(sdk.OneDec()) {
		// skip computation when collateral share is unrestricted
		return nil
	}

	share, err := k.VisibleCollateralShare(ctx, denom)
	if err != nil {
		return err
	}

	if share.GT(token.MaxCollateralShare) {
		return types.ErrMaxCollateralShare.Wrapf("%s share is %s", denom, share)
	}
	return nil
}
