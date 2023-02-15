package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

// maxWithdraw calculates the maximum amount of uTokens an account can currently withdraw.
// input denom should be a base token. If oracle prices are missing for some of the borrower's
// collateral (other than the denom being withdrawn), computes the maximum safe withdraw allowed
// by only the collateral whose prices are known
func (k *Keeper) maxWithdraw(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	uDenom := types.ToUTokenDenom(denom)
	availableTokens := sdk.NewCoin(denom, k.AvailableLiquidity(ctx, denom))
	availableUTokens, err := k.ExchangeToken(ctx, availableTokens)
	if err != nil {
		return sdk.Coin{}, err
	}
	totalBorrowed := k.GetBorrowerBorrows(ctx, addr)
	walletUtokens := k.bankKeeper.SpendableCoins(ctx, addr).AmountOf(uDenom)
	totalCollateral := k.GetBorrowerCollateral(ctx, addr)
	thisCollateral := sdk.NewCoin(uDenom, totalCollateral.AmountOf(uDenom))
	otherCollateral := totalCollateral.Sub(thisCollateral)

	// calculate borrowed value for the account, using the higher of spot or historic prices for each token
	borrowedValue, err := k.TotalTokenValue(ctx, totalBorrowed, types.PriceModeHigh)
	if nonOracleError(err) {
		// for errors besides a missing price, the whole transaction fails
		return sdk.Coin{}, err
	}
	if err != nil {
		// for missing prices on borrowed assets, we can't withdraw any collateral
		// but can withdraw non-collateral uTokens
		withdrawAmount := sdk.MinInt(walletUtokens, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), nil
	}

	// if no non-blacklisted tokens are borrowed, withdraw the maximum available amount
	if borrowedValue.IsZero() {
		withdrawAmount := walletUtokens.Add(totalCollateral.AmountOf(uDenom))
		withdrawAmount = sdk.MinInt(withdrawAmount, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), nil
	}

	// compute the borrower's borrow limit using all their collateral
	// except the denom being withdrawn (also excluding collateral missing oracle prices)
	otherBorrowLimit, err := k.VisibleBorrowLimit(ctx, otherCollateral)
	if err != nil {
		return sdk.Coin{}, err
	}
	// if their other collateral fully covers all borrows, withdraw the maximum available amount
	if borrowedValue.LT(otherBorrowLimit) {
		withdrawAmount := walletUtokens.Add(totalCollateral.AmountOf(uDenom))
		withdrawAmount = sdk.MinInt(withdrawAmount, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), nil
	}

	// for nonzero borrows, calculations are based on unused borrow limit
	// this treats collateral which is missing oracle prices as having zero value,
	// resulting in a lower borrow limit but not in an error
	borrowLimit, err := k.VisibleBorrowLimit(ctx, totalCollateral)
	if err != nil {
		return sdk.Coin{}, err
	}
	// borrowers above their borrow limit cannot withdraw collateral, but can withdraw wallet uTokens
	if borrowLimit.LTE(borrowedValue) {
		withdrawAmount := sdk.MinInt(walletUtokens, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), nil
	}

	// determine the USD amount of borrow limit that is currently unused
	unusedBorrowLimit := borrowLimit.Sub(borrowedValue)

	// calculate the contribution to borrow limit made by only the type of collateral being withdrawn
	// this WILL error on a missing price, since the cases where we know other collateral is sufficient
	// have all been eliminated
	specificBorrowLimit, err := k.CalculateBorrowLimit(ctx, sdk.NewCoins(thisCollateral))
	if err != nil {
		return sdk.Coin{}, err
	}

	// if only a portion of collateral is unused, withdraw only that portion
	unusedCollateralFraction := unusedBorrowLimit.Quo(specificBorrowLimit)
	unusedCollateral := unusedCollateralFraction.MulInt(thisCollateral.Amount).TruncateInt()

	// add wallet uTokens to the unused amount from collateral
	withdrawAmount := unusedCollateral.Add(walletUtokens)

	// reduce amount to withdraw if it exceeds available liquidity
	withdrawAmount = sdk.MinInt(withdrawAmount, availableUTokens.Amount)

	return sdk.NewCoin(uDenom, withdrawAmount), nil
}

// maxBorrow calculates the maximum amount of a given token an account can currently borrow.
// input denom should be a base token. If oracle prices are missing for some of the borrower's
// collateral, computes the maximum safe borrow allowed by only the collateral whose prices are known
func (k *Keeper) maxBorrow(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	if types.HasUTokenPrefix(denom) {
		return sdk.Coin{}, types.ErrUToken
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

	// calculate borrow limit for the account, using only collateral whose price is known
	borrowLimit, err := k.VisibleBorrowLimit(ctx, totalCollateral)
	if err != nil {
		return sdk.Coin{}, err
	}
	// borrowers above their borrow limit cannot borrow
	if borrowLimit.LTE(borrowedValue) {
		return sdk.NewCoin(denom, sdk.ZeroInt()), nil
	}

	// determine the USD amount of borrow limit that is currently unused
	unusedBorrowLimit := borrowLimit.Sub(borrowedValue)

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
