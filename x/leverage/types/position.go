package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountPosition must be created by NewAccountPosition for proper initialization.
// Contains an account's borrowed and collateral values, arranged into special asset
// pairs and normal assets. Special pairs will always be sorted by weight.
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
	// but will still be initialized with the proper weight, and sorted in order.
	// Special pairs which would reduce an asset's weight are included, but have no effect.
	specialPairs WeightedSpecialPairs
	// account's collateralValue USD value of each token, including that which is in special pairs above
	collateralValue sdk.DecCoins
	// account's borrowedValue USD value of each token, including that which is in special pairs above
	borrowedValue sdk.DecCoins
	// tracks whether the position was built using collateral weight or liquidation threshold
	isForLiquidation bool
	// caches retrieved token settings
	tokens map[string]Token
	// caches the minimum collateral weight or minimum liquidation threshold
	minimumBorrowFactor sdk.Dec
}

// NewAccountPosition creates an account position based on a user's borrowed and collateral
// values of each token, and additional information including token settings and special asset pairs.
func NewAccountPosition(
	tokens []Token,
	pairs []SpecialAssetPair,
	unsortedCollateralValue, unsortedBorrowValue sdk.DecCoins,
	forLiquidation bool,
	minimumBorrowFactor sdk.Dec,
) (AccountPosition, error) {
	position := AccountPosition{
		specialPairs:        WeightedSpecialPairs{},
		collateralValue:     sdk.DecCoins{},
		borrowedValue:       sdk.DecCoins{},
		isForLiquidation:    forLiquidation,
		tokens:              map[string]Token{},
		minimumBorrowFactor: minimumBorrowFactor,
	}

	// cache all registered tokens
	for _, t := range tokens {
		position.tokens[t.BaseDenom] = t
	}

	// cache all potentially relevant special asset pairs, and sort them by collateral weight (or liquidation threshold).
	// Initialize their amounts, which will eventually store matching asset value, to zero.
	for _, sp := range pairs {
		weight := sp.CollateralWeight
		if forLiquidation {
			weight = sp.LiquidationThreshold
		}
		if weight.LTE(
			// pairs may not reduce collateral weight or liquidation threshold
			// below what the tokens would produce without the special pair.
			sdk.MinDec(
				position.tokenWeight(sp.Collateral),
				sdk.MaxDec(position.tokenWeight(sp.Borrow), minimumBorrowFactor),
			),
		) || weight.IsZero() {
			// Such pairs as well as those with zero weight are omitted from the
			// position entirely.
			continue
		}
		wp := WeightedSpecialPair{
			Collateral:    sdk.NewDecCoinFromDec(sp.Collateral, sdk.ZeroDec()),
			Borrow:        sdk.NewDecCoinFromDec(sp.Borrow, sdk.ZeroDec()),
			SpecialWeight: weight,
		}
		// sorting is performed by Add function
		position.specialPairs = position.specialPairs.Add(wp)
	}

	// Copy borrowed and collateral value before special pairs are taken into account
	for _, cv := range unsortedCollateralValue {
		position.collateralValue = position.collateralValue.Add(cv)
	}
	for _, bv := range unsortedBorrowValue {
		position.borrowedValue = position.borrowedValue.Add(bv)
	}

	// match assets into special asset pairs, removing matched value from unsortedBorrowValue and unsortedCollateralValue
	// but not changing position.borrowedValue and position.collateralValue
	for i, p := range position.specialPairs {
		b := unsortedBorrowValue.AmountOf(p.Borrow.Denom)
		c := unsortedCollateralValue.AmountOf(p.Collateral.Denom)
		if b.IsPositive() && c.IsPositive() {
			// some unpaired assets match the special pair
			bCoin := sdk.NewDecCoinFromDec(p.Borrow.Denom, sdk.ZeroDec())
			cCoin := sdk.NewDecCoinFromDec(p.Collateral.Denom, sdk.ZeroDec())
			pairBorrowLimit := c.Mul(p.SpecialWeight)
			if pairBorrowLimit.GTE(b) {
				// all of the borrow is covered by collateral in this pair
				bCoin = sdk.NewDecCoinFromDec(bCoin.Denom, b)
				// some collateral, equal to borrow value / collateral weight, is used
				cCoin = sdk.NewDecCoinFromDec(cCoin.Denom, b.Quo(p.SpecialWeight))
			} else {
				// only some of the borrow, equal to collateral value * collateal weight is covered
				bCoin = sdk.NewDecCoinFromDec(bCoin.Denom, c.Mul(p.SpecialWeight))
				// all of the collateral is used
				cCoin = sdk.NewDecCoinFromDec(cCoin.Denom, c)
			}
			// subtract newly paired assets from unsorted assets
			unsortedBorrowValue = unsortedBorrowValue.Sub(sdk.NewDecCoins(bCoin))
			unsortedCollateralValue = unsortedCollateralValue.Sub(sdk.NewDecCoins(cCoin))
			// add newly paired assets to the appropriate special pair
			position.specialPairs[i].Borrow = bCoin
			position.specialPairs[i].Collateral = cCoin
		}
	}

	// always validates the position before returning for safety
	return position, position.Validate()
}

// validates basic properties of a position that should always be true
func (ap *AccountPosition) Validate() error {
	if err := ap.borrowedValue.Validate(); err != nil {
		return err
	}
	if err := ap.collateralValue.Validate(); err != nil {
		return err
	}
	for _, t := range ap.tokens {
		if err := t.Validate(); err != nil {
			return err
		}
	}
	if ap.minimumBorrowFactor.IsNil() || !ap.minimumBorrowFactor.IsPositive() {
		return fmt.Errorf("invalid minimum borrow factor: %s", ap.minimumBorrowFactor)
	}

	pairedBorrows := sdk.DecCoins{}
	pairedCollateral := sdk.DecCoins{}
	for _, sp := range ap.specialPairs {
		if err := sp.Validate(); err != nil {
			return err
		}
		pairedBorrows = pairedBorrows.Add(sp.Borrow)
		pairedCollateral = pairedCollateral.Add(sp.Collateral)
	}

	if _, neg := ap.borrowedValue.SafeSub(pairedBorrows); neg {
		return fmt.Errorf("special borrows exceeded total")
	}
	if _, neg := ap.collateralValue.SafeSub(pairedCollateral); neg {
		return fmt.Errorf("special collateral exceeded total")
	}

	return nil
}

// MaxBorrow finds the maximum additional amount of an asset a position can
// borrow without exceeding its borrow limit. Does not mutate position.
// If the requested token denom did not exist or the borrower was already
// at or over their borrow limit, returns zero.
// Returns zero if a position was computed with liquidation in mind.
func (ap *AccountPosition) MaxBorrow(denom string) sdk.Dec {
	if ap.isForLiquidation {
		return sdk.ZeroDec()
	}

	// Compute max borrow independently for borrow limit and borrow factor limitations
	maxBorrow := sdk.MinDec(ap.maxBorrowFromBorrowLimit(denom), ap.maxBorrowFromBorrowFactor(denom))
	// Prevent over-limit accounts from returning negative max borrow
	return sdk.MaxDec(sdk.ZeroDec(), maxBorrow)
}

// maxBorrowFromBorrowFactor is the subset of the max borrow calculation which considers
// borrow factor (not collateral weight)
func (ap *AccountPosition) maxBorrowFromBorrowFactor(denom string) sdk.Dec {
	borrowFactor := ap.borrowFactor(denom)
	if !borrowFactor.IsPositive() {
		return sdk.ZeroDec()
	}

	usage := ap.totalCollateralUsage()
	unusedCollateral := ap.CollateralValue().Sub(usage)
	unpairedCollateral := ap.unpairedCollateral()
	maxSpecialBorrow := sdk.ZeroDec()

	// if restricted by borrow factor, each special pair frees up additional collateral
	// for a given amount borrowed
	for _, wsp := range ap.specialPairs {
		if wsp.Borrow.Denom == denom {
			collateralToPair := sdk.MinDec(
				unusedCollateral,
				unpairedCollateral.AmountOf(wsp.Collateral.Denom),
			)
			borrowToPair := collateralToPair.Mul(wsp.SpecialWeight)
			// update totals and proceed to next pair
			unusedCollateral = unusedCollateral.Sub(collateralToPair)
			unpairedCollateral = unpairedCollateral.Sub(sdk.NewDecCoins(sdk.NewDecCoinFromDec(
				wsp.Collateral.Denom, collateralToPair,
			)))
			maxSpecialBorrow = maxSpecialBorrow.Add(borrowToPair)
		}
	}
	// max borrow value is equal new special pair borrows, plus remaining unused collateral's max borrow
	return unusedCollateral.Mul(borrowFactor).Add(maxSpecialBorrow)
}

// maxBorrowFromBorrowLimit is the subset of the max borrow calculation which considers
// collateral weight (not borrow factor)
func (ap *AccountPosition) maxBorrowFromBorrowLimit(denom string) sdk.Dec {
	unpairedBorrows := ap.unpairedBorrows()
	unpairedCollateral := ap.unpairedCollateral()
	unusedLimit := ap.normalBorrowLimit(unpairedCollateral).Sub(ap.total(unpairedBorrows))

	if !unusedLimit.IsPositive() {
		return sdk.ZeroDec()
	}

	// attempt to borrow using special pairs first
	specialBorrowed := sdk.ZeroDec()
	for _, wsp := range ap.specialPairs {
		if wsp.Borrow.Denom == denom && unusedLimit.IsPositive() {
			cDenom := wsp.Collateral.Denom
			cWeight := ap.tokenWeight(cDenom)
			// attempt to pair the maximum amount of collateral available
			collateralToPair := unpairedCollateral.AmountOf(cDenom)
			// limit to that collateral which is not being used by normal borrows
			collateralToPair = sdk.MinDec(collateralToPair,
				unusedLimit.Quo(cWeight),
			)
			// pair the assets and update totals
			unpairedCollateral = unpairedCollateral.Sub(sdk.NewDecCoins(sdk.NewDecCoinFromDec(
				cDenom, collateralToPair,
			)))
			unusedLimit = unusedLimit.Sub(
				collateralToPair.Mul(cWeight),
			)
			specialBorrowed = specialBorrowed.Add(
				collateralToPair.Mul(wsp.SpecialWeight),
			)
		}
	}

	// max borrow value is equal new special pair borrows, plus remaining normal borrow limit
	return unusedLimit.Add(specialBorrowed)
}

// MaxWithdraw finds the maximum additional amount of an asset a position can
// withdraw without exceeding its borrow limit. Does not mutate position.
// Returns zero if a position was computed with liquidation in mind.
// Also returns a boolean indicating whether total withdrawal is possible,
// to prevent downstream rounding errors when converting back to tokens.
func (ap *AccountPosition) MaxWithdraw(denom string) (sdk.Dec, bool) {
	if ap.isForLiquidation {
		return sdk.ZeroDec(), false
	}
	owned := ap.collateralValue.AmountOf(denom)
	if ap.borrowedValue.IsZero() {
		// return early on trivial case
		return owned, true
	}

	limit := ap.totalBorrowLimit()     // borrow limit after special pairs
	usage := ap.totalCollateralUsage() // collateral usage after special pairs

	collateralWeight := ap.tokenWeight(denom)
	if !collateralWeight.IsPositive() {
		// TODO: might not be accurate if special pairs exist - move this statement lower.
		return owned, false
	}

	//
	// TODO: withdraw first from normal, then from special pairs, one at a time.
	// this and verifying limit (and keeper logic) are all that's left, I think
	//

	// - for borrow limit, subtracting [collat * weight] from borrow limit
	//		- TODO: for special pairs, subtracting additional [collateral * delta weight]
	unusedLimit := limit.Sub(ap.BorrowedValue())
	max1 := unusedLimit.Quo(collateralWeight)

	// - for borrow factor, subtracting [collat] from TC
	//		- TODO: for special pairs, adding additional [borrow * delta factor] to collateral usage
	unusedCollateral := ap.CollateralValue().Sub(usage)
	max2 := unusedCollateral

	maxWithdraw := sdk.MinDec(max1, max2)                // lower of borrow limit and borrow factor results
	maxWithdraw = sdk.MinDec(maxWithdraw, owned)         // capped at owned amount
	maxWithdraw = sdk.MaxDec(maxWithdraw, sdk.ZeroDec()) // prevent negative value
	return maxWithdraw, maxWithdraw.GTE(owned)
}

// HasCollateral returns true if a position contains any collateral of a given
// type.
func (ap *AccountPosition) HasCollateral(denom string) bool {
	return ap.collateralValue.AmountOf(denom).IsPositive()
}

// Limit calculates the borrow limit of an account position
// (or liquidation threshold if ap.isForLiquidation is true).
func (ap *AccountPosition) Limit() sdk.Dec {
	collateralValue := ap.CollateralValue()
	if !collateralValue.IsPositive() {
		return sdk.ZeroDec()
	}

	// compute limit due to collateral weights
	limit := ap.totalBorrowLimit()

	// if no borrows, borrow factor limits will not apply
	borrowedValue := ap.BorrowedValue()
	if borrowedValue.IsZero() {
		return limit
	}

	// compute limit due to borrow factors
	usage := ap.totalCollateralUsage()
	unusedCollateralValue := collateralValue.Sub(usage) // can be negative

	var avgWeight sdk.Dec
	if unusedCollateralValue.IsNegative() {
		// if user if above limit, overused collateral is being borrowed against at
		// the borrow factor of the surplus borrows (which are by definition not paired)
		avgWeight = ap.averageBorrowFactor(ap.unpairedBorrows())
	} else {
		// if user is below limit, unused collateral can be borrowed against at the
		// average collateral weight of its unpaired collateral at most. But that is
		// borrow limit logic. Borrow factor considers the maximum possible borrow factor,
		// which if we are not specifying a borrow denom, is 1.0.
		avgWeight = sdk.OneDec()
	}
	borrowFactorLimit := ap.BorrowedValue().Add(unusedCollateralValue.Mul(avgWeight))

	// return the minimum of the two limits
	return sdk.MinDec(limit, borrowFactorLimit)
}

// BorrowedValue sums the total borrowed value in a position.
func (ap *AccountPosition) BorrowedValue() sdk.Dec {
	sum := sdk.ZeroDec()
	for _, b := range ap.borrowedValue {
		sum = sum.Add(b.Amount)
	}
	return sum
}

// CollateralValue sums the total collateral value in a position.
func (ap *AccountPosition) CollateralValue() sdk.Dec {
	sum := sdk.ZeroDec()
	for _, c := range ap.collateralValue {
		sum = sum.Add(c.Amount)
	}
	return sum
}

// IsHealthy() returns true if a position's borrowed value is below
// its borrow limit (or liquidation threshold if ap.isForLiquidation is true).
func (ap *AccountPosition) IsHealthy() bool {
	return ap.BorrowedValue().LTE(ap.Limit())
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

// specialBorrows returns the sum of all borrowed value in an account's special asset pairs
func (ap *AccountPosition) specialBorrows() sdk.DecCoins {
	special := sdk.NewDecCoins()
	for _, wsp := range ap.specialPairs {
		special = special.Add(wsp.Borrow)
	}
	return special
}

// unpairedBorrows returns an account's borrowed value minus any borrows tied up in special asset pairs
func (ap *AccountPosition) unpairedBorrows() sdk.DecCoins {
	total := sdk.NewDecCoins(ap.borrowedValue...)
	special := ap.specialBorrows()
	return total.Sub(special)
}

// specialCollateral returns the sum of all collateral value in an account's special asset pairs
func (ap *AccountPosition) specialCollateral() sdk.DecCoins {
	special := sdk.NewDecCoins()
	for _, wsp := range ap.specialPairs {
		special = special.Add(wsp.Collateral)
	}
	return special
}

// unpairedCollateral returns an account's collateral value minus any collateral tied up in special asset pairs
func (ap *AccountPosition) unpairedCollateral() sdk.DecCoins {
	total := sdk.NewDecCoins(ap.collateralValue...)
	special := ap.specialCollateral()
	return total.Sub(special)
}

// averageWeight gets the weighted average collateral weight (or liquidation threshold) of a set of tokens
func (ap *AccountPosition) averageWeight(coins sdk.DecCoins) sdk.Dec {
	if coins.IsZero() {
		return sdk.OneDec()
	}
	amountSum := sdk.ZeroDec()
	weightedSum := sdk.ZeroDec()
	for _, c := range coins {
		weightedSum = weightedSum.Add(c.Amount.Mul(ap.tokenWeight(c.Denom)))
		amountSum = amountSum.Add(c.Amount)
	}
	return weightedSum.Quo(amountSum)
}

// averageBorrowFactor gets the weighted average borrow factor of a set of tokens
func (ap *AccountPosition) averageBorrowFactor(coins sdk.DecCoins) sdk.Dec {
	if coins.IsZero() {
		return sdk.OneDec()
	}
	amountSum := sdk.ZeroDec()
	weightedSum := sdk.ZeroDec()
	for _, c := range coins {
		weightedSum = weightedSum.Add(c.Amount.Mul(ap.borrowFactor(c.Denom)))
		amountSum = amountSum.Add(c.Amount)
	}
	return weightedSum.Quo(amountSum)
}

// borrowFactor gets a token's collateral weight or liquidation threshold (or minimumBorrowFactor if greater)
// if the token is registered, else zero.
func (ap *AccountPosition) borrowFactor(denom string) sdk.Dec {
	if t, ok := ap.tokens[denom]; ok {
		if ap.isForLiquidation {
			return sdk.MaxDec(t.LiquidationThreshold, ap.minimumBorrowFactor)
		}
		return sdk.MaxDec(t.CollateralWeight, ap.minimumBorrowFactor)
	}
	return sdk.ZeroDec()
}

// totalCollateralUsage computes collateral usage of a position's unpaired borrows
// and adds collateral value from special asset pairs.
func (ap *AccountPosition) totalCollateralUsage() sdk.Dec {
	normal := ap.normalCollateralUsage(ap.unpairedBorrows())
	special := ap.total(ap.specialCollateral())
	return normal.Add(special)
}

// normalCollateralUsage calculated the minimum collateral value that can support borrowed sdk.DecCoins
// based on the borrow factor of those coins, without any special asset pairs being applied.
// Uses either collateral weight or liquidation threshold (if ap.isForLiquidation), or minimumBorrowFactor if greater.
func (ap *AccountPosition) normalCollateralUsage(borrowed sdk.DecCoins) sdk.Dec {
	sum := sdk.ZeroDec()
	for _, b := range borrowed {
		sum = sum.Add(
			b.Amount.Quo(sdk.MaxDec(
				ap.tokenWeight(b.Denom),
				ap.minimumBorrowFactor,
			)),
		)
	}
	return sum
}

// totalBorrowLimit computes the borrow limit of a position's unpaired collateral
// and then adds borrowed value from special asset pairs.
func (ap *AccountPosition) totalBorrowLimit() sdk.Dec {
	normal := ap.normalBorrowLimit(ap.unpairedCollateral())
	special := ap.total(ap.specialBorrows())
	return normal.Add(special)
}

// normalBorrowLimit is the total borrowed value which could be supported by
// collateral sdk.DecCoins, without any special asset pairs being applied.
// Uses either collateral weight or liquidation threshold (if ap.isForLiquidation)
func (ap *AccountPosition) normalBorrowLimit(collateral sdk.DecCoins) sdk.Dec {
	sum := sdk.ZeroDec()
	for _, b := range collateral {
		sum = sum.Add(b.Amount.Mul(ap.tokenWeight(b.Denom)))
	}
	return sum
}

// total sums the amounts in an sdk.DecCoins, regardless of denom
func (ap *AccountPosition) total(coins sdk.DecCoins) sdk.Dec {
	total := sdk.ZeroDec()
	for _, c := range coins {
		total = total.Add(c.Amount)
	}
	return total
}
