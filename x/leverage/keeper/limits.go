package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

// maxWithdraw calculates the maximum amount of uTokens an account can currently withdraw.
// input denom should be a base token.
func (k *Keeper) maxWithdraw(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	if types.HasUTokenPrefix(denom) {
		return sdk.Coin{}, types.ErrUToken
	}
	uDenom := types.ToUTokenDenom(denom)

	availableTokens := sdk.NewCoin(denom, k.AvailableLiquidity(ctx, denom))
	availableUTokens, err := k.ExchangeToken(ctx, availableTokens)
	if err != nil {
		return sdk.Coin{}, err
	}
	totalBorrowed := k.GetBorrowerBorrows(ctx, addr)
	walletUtokens := k.bankKeeper.SpendableCoins(ctx, addr).AmountOf(uDenom)
	totalCollateral := k.GetBorrowerCollateral(ctx, addr)
	specificCollateral := sdk.NewCoin(uDenom, totalCollateral.AmountOf(uDenom))

	// calculate borrowed value for the account, using the higher of spot or historic prices for each token
	borrowedValue, err := k.TotalTokenValue(ctx, totalBorrowed, types.PriceModeHigh)
	if err != nil {
		return sdk.Coin{}, err
	}

	// if no non-blacklisted tokens are borrowed, withdraw the maximum available amount
	if borrowedValue.IsZero() {
		withdrawAmount := walletUtokens.Add(totalCollateral.AmountOf(uDenom))
		withdrawAmount = sdk.MinInt(withdrawAmount, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), nil
	}

	// for nonzero borrows, calculations are based on unused borrow limit, using the lower of spot or historic prices
	borrowLimit, err := k.CalculateBorrowLimit(ctx, totalCollateral, types.PriceModeLow)
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
	specificBorrowLimit, err := k.CalculateBorrowLimit(ctx, sdk.NewCoins(specificCollateral), types.PriceModeLow)
	if err != nil {
		return sdk.Coin{}, err
	}
	if unusedBorrowLimit.GT(specificBorrowLimit) {
		// If borrow limit is sufficiently high even without this collateral, withdraw the full amount
		withdrawAmount := walletUtokens.Add(specificCollateral.Amount)
		withdrawAmount = sdk.MinInt(withdrawAmount, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), nil
	}

	// if only a portion of collateral is unused, withdraw only that portion
	unusedCollateralFraction := unusedBorrowLimit.Quo(specificBorrowLimit)
	unusedCollateral := unusedCollateralFraction.MulInt(specificCollateral.Amount).TruncateInt()

	// add wallet uTokens to the unused amount from collateral
	withdrawAmount := unusedCollateral.Add(walletUtokens)

	// reduce amount to withdraw if it exceeds available liquidity
	withdrawAmount = sdk.MinInt(withdrawAmount, availableUTokens.Amount)

	return sdk.NewCoin(uDenom, withdrawAmount), nil
}

// maxBorrow calculates the maximum amount of a given token an account can currently borrow.
// input denom should be a base token.
func (k *Keeper) maxBorrow(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	if types.HasUTokenPrefix(denom) {
		return sdk.Coin{}, types.ErrUToken
	}
	availableTokens := k.AvailableLiquidity(ctx, denom)

	totalBorrowed := k.GetBorrowerBorrows(ctx, addr)
	totalCollateral := k.GetBorrowerCollateral(ctx, addr)

	// calculate borrowed value for the account, using the higher of spot or historic prices
	borrowedValue, err := k.TotalTokenValue(ctx, totalBorrowed, types.PriceModeHigh)
	if err != nil {
		return sdk.Coin{}, err
	}

	// calculate borrow limit for the account, using the lower of spot or historic prices for collateral
	borrowLimit, err := k.CalculateBorrowLimit(ctx, totalCollateral, types.PriceModeLow)
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
	if err != nil {
		return sdk.Coin{}, err
	}

	// also cap borrow amount at available liquidity
	maxBorrow.Amount = sdk.MinInt(maxBorrow.Amount, availableTokens)

	return maxBorrow, nil
}

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

	// determine the amount of base tokens which would be required to reach maxValue,
	// using the hgiher of spot or historic prices
	udenom := types.ToUTokenDenom(denom)
	maxUTokens, err := k.UTokenWithValue(ctx, udenom, maxValue, types.PriceModeHigh)
	if err != nil {
		return sdk.ZeroInt(), err
	}

	// return the computed maximum or the current uToken supply, whichever is smaller
	return sdk.MinInt(k.GetUTokenSupply(ctx, denom).Amount, maxUTokens.Amount), nil
}
