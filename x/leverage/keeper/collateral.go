package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

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
// during liquidations.
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
// the x/leverage module account currently holds as collateral. Non-uTokens return zero.
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

// CollateralLiquidity calculates the current collateral liquidity of a token denom,
// which is defined as the token's liquidity, divided by the base token equivalent
// of associated uToken's total collateral. Ranges from 0 to 1.0
func (k Keeper) CollateralLiquidity(ctx sdk.Context, denom string) sdk.Dec {
	totalCollateral := k.GetTotalCollateral(ctx, types.ToUTokenDenom(denom))
	exchangeRate := k.DeriveExchangeRate(ctx, denom)
	liquidity := k.AvailableLiquidity(ctx, denom)

	// Zero collateral will be interpreted as full collateral liquidity. This encompasses two cases:
	// - liquidity / collateral = 0/0: Empty market, system is considered healthy by default
	// - liquidity / collateral = x/0: No collateral but nonzero liquidity, also considered healthy
	// In both cases, "all collateral is liquid" is technically true, given that there is no collateral.
	if totalCollateral.IsZero() {
		return sdk.OneDec()
	}

	collateralLiquidity := toDec(liquidity).Quo(exchangeRate.MulInt(totalCollateral.Amount))

	// Liquidity above 100% is ignored
	return sdk.MinDec(collateralLiquidity, sdk.OneDec())
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

// checkCollateralLiquidity returns the appropriate error if a token denom's
// collateral liquidity is below its MinCollateralLiquidity
func (k Keeper) checkCollateralLiquidity(ctx sdk.Context, denom string) error {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return err
	}

	collateralLiquidity := k.CollateralLiquidity(ctx, denom)
	if collateralLiquidity.LT(token.MinCollateralLiquidity) {
		return types.ErrMinCollateralLiquidity.Wrap(collateralLiquidity.String())
	}
	return nil
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

//// AvailableCollateralLiquidity
//func (k Keeper) AvailableCollateralLiquidity(ctx sdk.Context, denom string) (sdk.Dec, error) {
//	// Definition of collateral liquidity:
//	// CL = liquidity / collateral
//	//		where liquidity = (module balance - reserves)
//
//	// Effects of a collateral withdrawal:
//	//	liquidity decreases by token amount. collateral decreases by token amount.
//	//
//	// CL = L / C
//	//		= (L1 - X) / (C1 - X)
//
//	// Effects of a non-collateral withdrawal:
//	//	liquidity decreases by token amount. collateral is unchanged
//	//
//	// CL = L / C
//	//		= (L1 - X) / (C1)
//
//	// Effects of withdrawing X collateral and Y non-collateral:
//	//
//	// CL = L / C
//	//		= (L1 - X - Y) / (C1- X)
//
//	token, err := k.GetTokenSettings(ctx, types.ToTokenDenom(denom))
//	if err != nil {
//		return sdk.ZeroDec(), err
//	}
//
//
//
//	availableCollateralLiquidity := liquidity.
//}

// AvailableCollateralLiquidity2 calculates ...
func (k Keeper) AvailableCollateralLiquidity(ctx sdk.Context, denom string) (sdkmath.Int, error) {
	// CL = L / C
	//	  = (L1 - X - Y) / (C1 - X)

	// If targeting a given maximize (X+Y), should we try to withdraw X or Y first?

	// X = amount of collateral to withdraw
	// Y = amount of non-collateral to withdraw

	uDenom := types.ToUTokenDenom(denom)

	token, err := k.GetTokenSettings(ctx, types.ToTokenDenom(denom))
	if err != nil {
		return sdk.ZeroInt(), err
	}

	// L1
	liquidity := k.AvailableLiquidity(ctx, denom)

	totalCollateral := k.GetTotalCollateral(ctx, uDenom)

	// C1
	totalTokenCollateral, err := k.ExchangeUTokens(ctx, sdk.NewCoins(totalCollateral))
	if err != nil {
		return sdk.ZeroInt(), err
	}

	// CL
	minCollateralLiquidity := token.MinCollateralLiquidity

	// x = (L1 - CL*C1) / (1 - CL)
	totalAvailable := (sdk.NewDec(liquidity.Int64()).Sub(sdk.NewDec(totalTokenCollateral.AmountOf(denom).Int64()).Mul(minCollateralLiquidity))).Quo(sdk.NewDec(1).Sub(minCollateralLiquidity))
	uTokenTotalAvailable, err := k.ExchangeToken(ctx, sdk.NewCoin(uDenom, sdkmath.NewInt(totalAvailable.RoundInt64())))
	if err != nil {
		return sdk.ZeroInt(), err
	}

	return uTokenTotalAvailable.Amount, nil
}
