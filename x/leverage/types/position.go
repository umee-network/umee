package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// minimum borrow factor is the minimum collateral weight and minimum liquidation threshold
// allowed when a borrowed token is limiting the efficiency of a pair of assets
var minimumBorrowFactor = sdk.MustNewDecFromStr("0.5")

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
	// isForLiquidation tracks whether the position was built using collateral weight
	// or liquidation threshold
	isForLiquidation bool
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
	isLiquidation bool,
) AccountPosition {
	position := AccountPosition{
		specialPairs:     WeightedSpecialPairs{},
		collateral:       WeightedDecCoins{},
		borrowed:         WeightedDecCoins{},
		tokens:           map[string]Token{},
		collateralValue:  sdk.ZeroDec(),
		borrowedValue:    sdk.ZeroDec(),
		isForLiquidation: isLiquidation,
	}

	// cache all registered tokens
	for _, t := range tokens {
		position.tokens[t.BaseDenom] = t
	}

	// cache all potentially relevant special asset pairs, and sort them by collateral weight (or liquidation threshold).
	// Initialize their amounts, which will eventually store matching asset value, to zero.
	for _, sp := range pairs {
		weight := sdk.MaxDec(
			sp.CollateralWeight,
			// pairs may not reduce collateral weight or liquidation threshold
			// below what the tokens would produce without the special pair
			sdk.MinDec(
				position.tokenWeight(sp.Collateral),
				sdk.MaxDec(position.tokenWeight(sp.Borrow), minimumBorrowFactor),
			),
		)
		wp := WeightedSpecialPair{
			Collateral:    sdk.NewDecCoinFromDec(sp.Collateral, sdk.ZeroDec()),
			Borrow:        sdk.NewDecCoinFromDec(sp.Borrow, sdk.ZeroDec()),
			SpecialWeight: weight,
		}
		// sorting is performed by Add function
		position.specialPairs = position.specialPairs.Add(wp)
	}

	for _, cv := range collateralValue {
		// track total collateral value
		position.collateralValue = position.collateralValue.Add(cv.Amount)
	}
	for _, bv := range borrowedValue {
		// track total borrowed value
		position.borrowedValue = position.borrowedValue.Add(bv.Amount)
	}

	// match assets into special asset pairs, removing matched value from collateralValue and borrowedValue
	for _, p := range position.specialPairs {
		b := borrowedValue.AmountOf(p.Borrow.Denom)
		c := collateralValue.AmountOf(p.Collateral.Denom)
		if b.IsPositive() && c.IsPositive() {
			// some unmatched assets match the special pair
			pairBorrowLimit := c.Mul(p.SpecialWeight)
			if pairBorrowLimit.GTE(b) {
				// all of the borrow is covered by collateral in this pair
				borrowedValue = borrowedValue.Sub(sdk.NewDecCoins(sdk.NewDecCoinFromDec(p.Borrow.Denom, b)))
				collateralValue = collateralValue.Sub(sdk.NewDecCoins(sdk.NewDecCoinFromDec(
					// some collateral, equal to borrow value / collateral weight, is used
					p.Collateral.Denom, b.Quo(p.SpecialWeight),
				)))
			} else {
				// only some of the borrow, equal to collateral value * collateal weight is covered
				borrowedValue = borrowedValue.Sub(sdk.NewDecCoins(sdk.NewDecCoinFromDec(
					p.Borrow.Denom, c.Mul(p.SpecialWeight),
				)))
				collateralValue = collateralValue.Sub(sdk.NewDecCoins(sdk.NewDecCoinFromDec(p.Collateral.Denom, c)))
			}
		}
	}

	for _, cv := range collateralValue {
		// sort collateral by collateral weight (or liquidation threshold) using Add function
		position.collateral = position.collateral.Add(
			WeightedDecCoin{
				Asset:  sdk.NewDecCoinFromDec(cv.Denom, cv.Amount),
				Weight: position.tokenWeight(cv.Denom),
			},
		)
	}

	for _, bv := range borrowedValue {
		// sort borrows by collateral weight (or liquidation threshold) using Add function
		position.borrowed = position.borrowed.Add(
			WeightedDecCoin{
				Asset:  sdk.NewDecCoinFromDec(bv.Denom, bv.Amount),
				Weight: position.tokenWeight(bv.Denom),
			},
		)
	}

	return position
}

// BorrowedValue returns an account's total USD value borrowed
func (ap *AccountPosition) BorrowedValue() sdk.Dec {
	return ap.borrowedValue
}

// CollateralValue returns an account's collateral's total USD value
func (ap *AccountPosition) CollateralValue() sdk.Dec {
	return ap.collateralValue
}

// tokenWeight gets a token's collateral weight or liquidation threshold if it is registered, else zero
func (ap *AccountPosition) tokenWeight(denom string) sdk.Dec {
	if t, ok := ap.tokens[denom]; ok {
		if ap.isForLiquidation {
			return t.LiquidationThreshold
		}
		return t.CollateralWeight
	}
	return sdk.ZeroDec()
}

// Limit computes the borrow limit or liquidation threshold of a position, depending on position.isForLiquidation.
// The result may be less or more than its borrowed value.
func (ap *AccountPosition) Limit() sdk.Dec {
	// An initialized account position already has special asset pairs matched up, so
	// this function only needs to deal with assets outside of special pairs.

	// These slices of position's weighted coins will be mutated in the matching process
	remainingBorrow := ap.borrowed
	remainingCollateral := ap.collateral

	var i, j int
	for i < len(remainingCollateral) && j < len(remainingBorrow) {
		cDenom := remainingCollateral[i].Asset.Denom
		bDenom := remainingBorrow[j].Asset.Denom
		c := remainingCollateral[i].Asset.Amount
		b := remainingBorrow[j].Asset.Amount
		w := sdk.MinDec(
			remainingCollateral[i].Weight,
			sdk.MaxDec(remainingBorrow[j].Weight, minimumBorrowFactor),
		)
		// match collateral and borrow at indexes i and j, exhausting at least one of them
		pairBorrowLimit := c.Mul(w)
		if pairBorrowLimit.GTE(b) {
			// all of the borrow is covered by collateral in this pair
			remainingBorrow = remainingBorrow.Sub(sdk.NewDecCoinFromDec(bDenom, b))
			remainingCollateral = remainingCollateral.Sub(sdk.NewDecCoinFromDec(
				// some collateral, equal to borrow value / collateral weight, is used
				cDenom, b.Quo(w),
			))
			// next borrow
			j++
		} else {
			// only some of the borrow, equal to collateral value * collateral weight is covered
			// if in a previous step the collateral and borrow both reached zero, this is zero
			remainingBorrow = remainingBorrow.Sub(sdk.NewDecCoinFromDec(
				bDenom, c.Mul(w),
			))
			remainingCollateral = remainingCollateral.Sub(sdk.NewDecCoinFromDec(cDenom, c))
			// next collateral
			i++
		}
	}

	// if any borrows remain after matching, user is over borrow limit (or LT) by remaining value
	remainingBorrowValue := remainingBorrow.Total()
	if remainingBorrowValue.IsPositive() {
		return ap.borrowedValue.Sub(remainingBorrowValue)
	}

	// if no borrows remain after matching, user may have additional borrow limit (or LT) available
	limit := ap.borrowedValue
	for _, c := range remainingCollateral {
		// the borrow limit calculation assumes no additional limitations based on borrow factor
		limit = limit.Add(c.Asset.Amount.Mul(c.Weight))
	}
	return limit
}

// MaxBorrow
func (ap *AccountPosition) MaxBorrow(denom string) sdk.Dec {
	return sdk.ZeroDec()
}

// MaxWithdraw
func (ap *AccountPosition) MaxWithdraw(denom string) sdk.Dec {
	return sdk.ZeroDec()
}

// TODO: bump to the bottom, or top, when computing max borrow
// TODO: similar when computing max withdraw
// TODO: isolate special pairs and bump
