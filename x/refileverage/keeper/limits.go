package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v5/x/refileverage/types"
)

// userMaxWithdraw calculates the maximum amount of uTokens an account can currently withdraw and the amount of
// these uTokens is non-collateral. Input denom should be a base token. If oracle prices are missing for some of the
// borrower's collateral (other than the denom being withdrawn), computes the maximum safe withdraw allowed by only
// the collateral whose prices are known.
func (k *Keeper) userMaxWithdraw(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, sdk.Coin, error) {
	uDenom := types.ToUTokenDenom(denom)
	borrowed := k.GetBorrowerBorrows(ctx, addr)
	walletUtokenAmount := k.bankKeeper.SpendableCoins(ctx, addr).AmountOf(uDenom)
	walletUCoins := sdk.NewCoin(uDenom, walletUtokenAmount)
	totalCollateral := k.GetBorrowerCollateral(ctx, addr)
	thisCollateral := sdk.NewCoin(uDenom, totalCollateral.AmountOf(uDenom))
	otherCollateral := totalCollateral.Sub(thisCollateral)
	unbondedCollateral := k.unbondedCollateral(ctx, addr, uDenom)

	// if no non-blacklisted tokens are borrowed, withdraw the maximum available amount
	if borrowed.IsZero() {
		withdrawAmount := walletUtokenAmount.Add(unbondedCollateral.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), walletUCoins, nil
	}

	// compute the borrower's borrow limit using all their collateral
	// except the denom being withdrawn (also excluding collateral missing oracle prices)
	otherBorrowLimit, err := k.VisibleBorrowLimit(ctx, otherCollateral)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	// if their other collateral fully covers all borrows, withdraw the maximum available amount
	if borrowed.LT(otherBorrowLimit) {
		withdrawAmount := walletUtokenAmount.Add(unbondedCollateral.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), walletUCoins, nil
	}

	// for nonzero borrows, calculations are based on unused borrow limit
	// this treats collateral which is missing oracle prices as having zero value,
	// resulting in a lower borrow limit but not in an error
	borrowLimit, err := k.VisibleBorrowLimit(ctx, totalCollateral)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	// borrowers above their borrow limit cannot withdraw collateral, but can withdraw wallet uTokens
	if borrowLimit.LTE(borrowed) {
		return walletUCoins, walletUCoins, nil
	}

	unusedBorrowLimit := borrowLimit.Sub(borrowed)

	// calculate the contribution to borrow limit made by only the type of collateral being withdrawn
	// this WILL error on a missing price, since the cases where we know other collateral is sufficient
	// have all been eliminated
	specificBorrowLimit, err := k.CalculateBorrowLimit(ctx, sdk.NewCoins(thisCollateral))
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// if only a portion of collateral is unused, withdraw only that portion
	unusedCollateralFraction := unusedBorrowLimit.Quo(specificBorrowLimit)
	unusedCollateral := unusedCollateralFraction.MulInt(thisCollateral.Amount).TruncateInt()

	// find the minimum of unused collateral (due to borrows) or unbonded collateral (incentive module)
	if unbondedCollateral.Amount.LT(unusedCollateral) {
		unusedCollateral = unbondedCollateral.Amount
	}

	// add wallet uTokens to the unused amount from collateral
	withdrawAmount := unusedCollateral.Add(walletUtokenAmount)

	return sdk.NewCoin(uDenom, withdrawAmount), walletUCoins, nil
}

// userMaxBorrow calculates the maximum amount of a given token an account can currently borrow.
// input denom should be a base token. If oracle prices are missing for some of the borrower's
// collateral, computes the maximum safe borrow allowed by only the collateral whose prices are known
func (k *Keeper) userMaxBorrow(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	if types.HasUTokenPrefix(denom) {
		return sdk.Coin{}, types.ErrUToken
	}

	borrowed := k.GetBorrowerBorrows(ctx, addr)
	totalCollateral := k.GetBorrowerCollateral(ctx, addr)

	// calculate borrow limit for the account, using only collateral whose price is known
	borrowLimit, err := k.VisibleBorrowLimit(ctx, totalCollateral)
	if err != nil {
		return sdk.Coin{}, err
	}
	// borrowers above their borrow limit cannot borrow
	if borrowLimit.LTE(borrowed) {
		return sdk.NewCoin(denom, sdk.ZeroInt()), nil
	}

	// determine the USD amount of borrow limit that is currently unused
	unusedBorrowLimit := borrowLimit.Sub(borrowed)

	// determine max borrow, using the higher of spot or historic prices for the token to borrow
	maxBorrow, err := k.TokenWithValue(ctx, denom, unusedBorrowLimit, types.PriceModeHigh)
	if nonOracleError(err) {
		// non-oracle errors fail the transaction (or query)
		return sdk.Coin{}, err
	}
	if err != nil {
		// oracle errors cause max borrow to be zero
		return sdk.NewCoin(denom, sdk.ZeroInt()), nil
	}

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
	otherDenomsValue, err := k.VisibleCollateralValue(ctx, systemCollateral.Sub(thisDenomCollateral))
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
