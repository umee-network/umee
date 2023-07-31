package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountPosition must be created by NewAccountPosition for proper initialization.
// Contains an account's borrowed and collateral values, arranged into special asset
// pairs and regular assets. Each list will always be sorted by collateral weight.
// Also caches some relevant values that will be reused in computation, like token
// settings and the borrower's total borrowed value and collateral value.
type AccountPosition struct {
	// all special asset pairs which apply to the account. Specifically, this should
	// cache any special pairs which match the account's collateral, even if they do
	// not match one of its borrows. This is the widest set of information that could
	// be needed when calculating maxWithdraw of an already collateralized asset, or
	// maxBorrow of any asset (even if not initially present in the position).
	// A pair that does not currently apply to the account (but could do so if a borrow
	// were added or existing pairs were rearranged) will have zero USD value
	// but will still be initialized with the proper weights, and sorted in order.
	specialPairs WeightedSpecialPairs
	// all collateral value not being used for special asset pairs
	collateral WeightedDecCoins
	// all borrowed value not being used for special asset pairs
	borrowed WeightedDecCoins
	// caches retrieved token settings
	tokens map[string]Token
	// collateral value (using PriceModeLow and interpreting unknown prices as zero)
	collateralValue sdk.Dec
	// borrowed value (using PriceModeHigh and requiring all prices known)
	borrowedValue sdk.Dec
}

// NewAccountPosition creates and sorts an account position based on token settings,
// special asset pairs, and the collateral and borrowed value of each token in an account.
// Once this structure is created, borrow limit calculations can be performed without
// keeper or context.
func NewAccountPosition(
	tokens []Token,
	pairs []SpecialAssetPair,
	collateralValue,
	borrowedValue sdk.DecCoins,
) AccountPosition {
	position := AccountPosition{
		specialPairs:    WeightedSpecialPairs{},
		collateral:      WeightedDecCoins{},
		borrowed:        WeightedDecCoins{},
		tokens:          map[string]Token{},
		collateralValue: sdk.ZeroDec(),
		borrowedValue:   sdk.ZeroDec(),
	}

	// cache all registered tokens
	for _, t := range tokens {
		position.tokens[t.BaseDenom] = t
	}

	// cache all potentially relevant special asset pairs, and sort them by collateral weight.
	// Initialize their amounts, which will eventually store matching asset value, to zero.
	for _, sp := range pairs {
		wp := WeightedSpecialPair{
			Collateral:           sdk.NewDecCoinFromDec(sp.Collateral, sdk.ZeroDec()),
			Borrow:               sdk.NewDecCoinFromDec(sp.Borrow, sdk.ZeroDec()),
			SpecialWeight:        sp.CollateralWeight,
			LiquidationThreshold: sp.LiquidationThreshold,
		}
		// sorting is performed by Add function
		position.specialPairs = position.specialPairs.Add(wp)
	}

	for _, cv := range collateralValue {
		// track total collateral value
		position.collateralValue = position.collateralValue.Add(cv.Amount)
		// sort collateral by collateral weight using Add function
		position.collateral = position.collateral.Add(
			WeightedDecCoin{
				Asset:       sdk.NewDecCoinFromDec(cv.Denom, cv.Amount),
				Weight:      position.tokenCollateralWeight(cv.Denom),
				Liquidation: position.tokenLiquidationThreshold(cv.Denom),
			},
		)
	}

	for _, bv := range borrowedValue {
		// track total borrowed value
		position.borrowedValue = position.borrowedValue.Add(bv.Amount)
		// sort borrows by collateral weight using Add function
		position.borrowed = position.borrowed.Add(
			WeightedDecCoin{
				Asset:       sdk.NewDecCoinFromDec(bv.Denom, bv.Amount),
				Weight:      position.tokenCollateralWeight(bv.Denom),
				Liquidation: position.tokenLiquidationThreshold(bv.Denom),
			},
		)
	}

	// TODO: match regular assets into special asset pairs on init

	return position
}

// tokenCollateralWeight gets a token's collateral weight if it is registered, else zero
func (ap *AccountPosition) tokenCollateralWeight(denom string) sdk.Dec {
	if t, ok := ap.tokens[denom]; ok {
		return t.CollateralWeight
	}
	return sdk.ZeroDec()
}

// tokenLiquidationThreshold gets a token's liquidation threshold if it is registered, else zero
func (ap *AccountPosition) tokenLiquidationThreshold(denom string) sdk.Dec {
	if t, ok := ap.tokens[denom]; ok {
		return t.LiquidationThreshold
	}
	return sdk.ZeroDec()
}

// TODO: bump to the bottom, or top, when computing max borrow
// TODO: similar when computing max withdraw
// TODO: isolate special pairs and bump
