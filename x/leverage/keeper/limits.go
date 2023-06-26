package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v5/x/leverage/types"
)

// userMaxWithdraw calculates the maximum amount of uTokens an account can currently withdraw and the amount of
// these uTokens which are non-collateral. Input denom should be a base token. If oracle prices are missing for some of the
// borrower's collateral (other than the denom being withdrawn), computes the maximum safe withdraw allowed by only
// the collateral whose prices are known.
func (k *Keeper) userMaxWithdraw(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, sdk.Coin, error) {
	uDenom := types.ToUTokenDenom(denom)
	availableTokens := sdk.NewCoin(denom, k.AvailableLiquidity(ctx, denom))
	availableUTokens, err := k.ExchangeToken(ctx, availableTokens)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	totalBorrowed := k.GetBorrowerBorrows(ctx, addr)
	walletUtokens := k.bankKeeper.SpendableCoins(ctx, addr).AmountOf(uDenom)
	totalCollateral := k.GetBorrowerCollateral(ctx, addr)
	thisCollateral := sdk.NewCoin(uDenom, totalCollateral.AmountOf(uDenom))
	otherCollateral := totalCollateral.Sub(thisCollateral)
	unbondedCollateral := k.unbondedCollateral(ctx, addr, uDenom)

	// calculate borrowed value for the account, using the higher of spot or historic prices for each token
	borrowedValue, err := k.TotalTokenValue(ctx, totalBorrowed, types.PriceModeHigh)
	if nonOracleError(err) {
		// for errors besides a missing price, the whole transaction fails
		return sdk.Coin{}, sdk.Coin{}, err
	}
	if err != nil {
		// for missing prices on borrowed assets, we can't withdraw any collateral
		// but can withdraw non-collateral uTokens
		withdrawAmount := sdk.MinInt(walletUtokens, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), sdk.NewCoin(uDenom, walletUtokens), nil
	}

	// calculate collateral value for the account, using the lower of spot or historic prices for each token
	// will count collateral with missing prices as zero value without returning an error
	collateralValue, err := k.VisibleCollateralValue(ctx, totalCollateral, types.PriceModeLow)
	if err != nil {
		// for errors besides a missing price, the whole transaction fails
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// calculate weighted borrowed value - used by the borrow factor limit
	weightedBorrowValue, err := k.WeightedBorrowValue(ctx, totalBorrowed, types.PriceModeHigh)
	if nonOracleError(err) {
		// for errors besides a missing price, the whole transaction fails
		return sdk.Coin{}, sdk.Coin{}, err
	}
	if err != nil {
		// for missing prices on borrowed assets, we can't withdraw any collateral
		// but can withdraw non-collateral uTokens
		withdrawAmount := sdk.MinInt(walletUtokens, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), sdk.NewCoin(uDenom, walletUtokens), nil
	}

	// if no non-blacklisted tokens are borrowed, withdraw the maximum available amount
	if borrowedValue.IsZero() {
		withdrawAmount := walletUtokens.Add(unbondedCollateral.Amount)
		withdrawAmount = sdk.MinInt(withdrawAmount, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), sdk.NewCoin(uDenom, walletUtokens), nil
	}

	// compute the borrower's borrow limit using all their collateral
	// except the denom being withdrawn (also excluding collateral missing oracle prices)
	otherCollateralBorrowLimit, err := k.VisibleBorrowLimit(ctx, otherCollateral)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	// if their other collateral fully covers all borrows, withdraw the maximum available amount
	if borrowedValue.LT(otherCollateralBorrowLimit) {
		// also check collateral value vs weighted borrow (borrow factor limit)
		otherCollateralValue, err := k.VisibleCollateralValue(ctx, otherCollateral, types.PriceModeLow)
		if err != nil {
			return sdk.Coin{}, sdk.Coin{}, err
		}
		// if weighted borrow does not exceed other collateral value, this collateral can be fully withdrawn
		if otherCollateralValue.GTE(weightedBorrowValue) {
			// in this case, both borrow limits will not be exceeded even if all collateral is withdrawn
			withdrawAmount := walletUtokens.Add(unbondedCollateral.Amount)
			withdrawAmount = sdk.MinInt(withdrawAmount, availableUTokens.Amount)
			return sdk.NewCoin(uDenom, withdrawAmount), sdk.NewCoin(uDenom, walletUtokens), nil
		}
	}

	// for nonzero borrows, calculations are based on unused borrow limit
	// this treats collateral which is missing oracle prices as having zero value,
	// resulting in a lower borrow limit but not in an error
	borrowLimit, err := k.VisibleBorrowLimit(ctx, totalCollateral)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	// borrowers above either of their borrow limits cannot withdraw collateral, but can withdraw wallet uTokens
	if borrowLimit.LTE(borrowedValue) || collateralValue.LTE(weightedBorrowValue) {
		withdrawAmount := sdk.MinInt(walletUtokens, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), sdk.NewCoin(uDenom, walletUtokens), nil
	}

	// determine the USD amount of borrow limit that is currently unused
	unusedBorrowLimit := borrowLimit.Sub(borrowedValue)

	// calculate the contribution to borrow limit made by only the type of collateral being withdrawn
	// this WILL error on a missing price, since the cases where we know other collateral is sufficient
	// have all been eliminated
	specificBorrowLimit, err := k.CalculateBorrowLimit(ctx, sdk.NewCoins(thisCollateral))
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// if only a portion of collateral is unused, withdraw only that portion (regular borrow limit)
	unusedCollateralFraction := unusedBorrowLimit.Quo(specificBorrowLimit)

	// calculate value of this collateral specifically, which is used in borrow factor's borrow limit
	specificCollateralValue, err := k.CalculateCollateralValue(ctx, sdk.NewCoins(thisCollateral), types.PriceModeLow)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	unusedCollateralValue := collateralValue.Sub(weightedBorrowValue)
	// Find the more restrictive of either borrow factor limit or borrow limit
	unusedCollateralFraction = sdk.MinDec(unusedCollateralFraction, unusedCollateralValue.Quo(specificCollateralValue))

	// Both borrow limits are satisfied by this withdrawl amount. The restrictions below relate to neither.
	unusedCollateral := unusedCollateralFraction.MulInt(thisCollateral.Amount).TruncateInt()

	// find the minimum of unused collateral (due to borrows) or unbonded collateral (incentive module)
	if unbondedCollateral.Amount.LT(unusedCollateral) {
		unusedCollateral = unbondedCollateral.Amount
	}

	// add wallet uTokens to the unused amount from collateral
	withdrawAmount := unusedCollateral.Add(walletUtokens)

	// reduce amount to withdraw if it exceeds available liquidity
	withdrawAmount = sdk.MinInt(withdrawAmount, availableUTokens.Amount)

	return sdk.NewCoin(uDenom, withdrawAmount), sdk.NewCoin(uDenom, walletUtokens), nil
}

// userMaxBorrow calculates the maximum amount of a given token an account can currently borrow.
// input denom should be a base token. If oracle prices are missing for some of the borrower's
// collateral, computes the maximum safe borrow allowed by only the collateral whose prices are known.
func (k *Keeper) userMaxBorrow(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	if types.HasUTokenPrefix(denom) {
		return sdk.Coin{}, types.ErrUToken
	}
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	availableTokens := k.AvailableLiquidity(ctx, denom)

	totalBorrowed := k.GetBorrowerBorrows(ctx, addr)
	totalCollateral := k.GetBorrowerCollateral(ctx, addr)

	// calculate borrowed value for the account, using the higher of spot or historic prices
	borrowedValue, err := k.TotalTokenValue(ctx, totalBorrowed, types.PriceModeHigh)
	if nonOracleError(err) {
		// non-oracle errors fail the transaction (or query)
		return sdk.Coin{}, err
	}
	if err != nil {
		// oracle errors cause max borrow to be zero
		return sdk.NewCoin(denom, sdk.ZeroInt()), nil
	}

	// calculate weighted borrowed value for the account, using the higher of spot or historic prices
	weightedBorrowedValue, err := k.WeightedBorrowValue(ctx, totalBorrowed, types.PriceModeHigh)
	if nonOracleError(err) {
		// non-oracle errors fail the transaction (or query)
		return sdk.Coin{}, err
	}
	if err != nil {
		// oracle errors cause max borrow to be zero
		return sdk.NewCoin(denom, sdk.ZeroInt()), nil
	}

	// calculate borrow limit for the account, using only collateral whose price is known
	borrowLimit, err := k.VisibleBorrowLimit(ctx, totalCollateral)
	if err != nil {
		return sdk.Coin{}, err
	}
	// borrowers above their borrow limit cannot borrow
	if borrowLimit.LTE(borrowedValue) {
		return sdk.NewCoin(denom, sdk.ZeroInt()), nil
	}

	// calculate collateral value limit for the account, using only collateral whose price is known
	collateralValue, err := k.VisibleCollateralValue(ctx, totalCollateral, types.PriceModeLow)
	if err != nil {
		return sdk.Coin{}, err
	}
	// borrowers above their borrow factor borrow limit cannot borrow
	if collateralValue.LTE(weightedBorrowedValue) {
		return sdk.NewCoin(denom, sdk.ZeroInt()), nil
	}

	// determine the USD amount of borrow limit that is currently unused
	unusedBorrowLimit := borrowLimit.Sub(borrowedValue)

	// determine the USD amount that can be borrowed according to borrow factor limit
	maxBorrowValueIncrease := collateralValue.Sub(weightedBorrowedValue).Quo(token.BorrowFactor())

	// finds the most restrictive of regular borrow limit and borrow factor limit
	valueToBorrow := sdk.MinDec(unusedBorrowLimit, maxBorrowValueIncrease)

	// determine max borrow, using the higher of spot or historic prices for the token to borrow
	maxBorrow, err := k.TokenWithValue(ctx, denom, valueToBorrow, types.PriceModeHigh)
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
