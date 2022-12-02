package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/leverage/types"
)

// maxWithdraw calculates the maximum amount of uTokens an account can currently withdraw.
// input should be a base token.
func (k *Keeper) maxWithdraw(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	uDenom := types.ToUTokenDenom(denom)

	availableTokens := sdk.NewCoin(denom, k.AvailableLiquidity(ctx, denom))
	availableUTokens, err := k.ExchangeToken(ctx, availableTokens)
	if err != nil {
		return sdk.Coin{}, err
	}
	walletUtokens := k.bankKeeper.SpendableCoins(ctx, addr).AmountOf(uDenom)
	totalCollateral := k.GetBorrowerCollateral(ctx, addr)
	totalBorrowed := k.GetBorrowerBorrows(ctx, addr)

	// calculate borrowed value for the account
	borrowedValue, err := k.TotalTokenValue(ctx, totalBorrowed)
	if err != nil {
		return sdk.Coin{}, err
	}

	// if no non-blacklisted tokens are borrowed, withdraw the maximum available amount
	if borrowedValue.IsZero() {
		withdrawAmount := walletUtokens.Add(totalCollateral.AmountOf(uDenom))
		withdrawAmount = sdk.MinInt(withdrawAmount, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), nil
	}

	// for nonzero borrows, calculations are based on unused borrow limit
	borrowLimit, err := k.CalculateBorrowLimit(ctx, totalCollateral)
	if err != nil {
		return sdk.Coin{}, err
	}
	// borrowers above their borrow limit cannot withdraw collateral, but can withdraw wallet uTokens
	if borrowLimit.LTE(borrowedValue) {
		withdrawAmount := sdk.MinInt(walletUtokens, availableUTokens.Amount)
		return sdk.NewCoin(uDenom, withdrawAmount), nil
	}

	// get collateral weight
	ts, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// get uToken exchange rate
	uTokenExchangeRate := k.DeriveExchangeRate(ctx, denom)

	// determine how many uTokens can be withdrawn from collateral to reduce borrow limit to borrowed value
	unusedBorrowLimit := borrowLimit.Sub(borrowedValue)
	unusedCollateral := unusedBorrowLimit.Quo(uTokenExchangeRate).Quo(ts.CollateralWeight).TruncateInt()

	// determine how many uTokens can actually be withdrawn from collateral alone
	withdrawAmount := sdk.MinInt(unusedCollateral, totalCollateral.AmountOf(uDenom))

	// add wallet uTokens to the amount from collateral
	withdrawAmount = withdrawAmount.Add(walletUtokens)

	return sdk.NewCoin(uDenom, sdk.MinInt(withdrawAmount, availableUTokens.Amount)), nil
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

	// determine the amount of uTokens which would be required to reach maxValue
	tokenDenom := types.ToTokenDenom(denom)
	tokenPrice, err := k.TokenBasePrice(ctx, tokenDenom)
	if err != nil {
		return sdk.ZeroInt(), err
	}
	uTokenExchangeRate := k.DeriveExchangeRate(ctx, tokenDenom)

	// in the case of a base token price smaller than the smallest sdk.Dec (10^-18),
	// this maxCollateralAmount will use the price of 10^-18 and thus derive a lower
	// (more cautious) limit than a precise price would produce
	maxCollateralAmount := maxValue.Quo(tokenPrice).Quo(uTokenExchangeRate).TruncateInt()

	// return the computed maximum or the current uToken supply, whichever is smaller
	return sdk.MinInt(k.GetUTokenSupply(ctx, denom).Amount, maxCollateralAmount), nil
}
