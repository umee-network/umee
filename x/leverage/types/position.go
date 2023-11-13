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
	mapTokens := map[string]Token{}
	weightedPairs := WeightedSpecialPairs{}

	// arrange all registered tokens
	for _, t := range tokens {
		mapTokens[t.BaseDenom] = t
	}

	// arrange all potentially relevant special asset pairs, and sort them by collateral weight (or liquidation threshold).
	// Initialize their amounts, which will eventually store matching asset value, to zero.
	temp := AccountPosition{
		tokens: mapTokens, // temp position to use tokenWeight function
	}
	for _, sp := range pairs {
		weight := sp.CollateralWeight
		if forLiquidation {
			weight = sp.LiquidationThreshold
		}
		if weight.LTE(
			// pairs may not reduce collateral weight or liquidation threshold
			// below what the tokens would produce without the special pair.
			sdk.MinDec(
				temp.tokenWeight(sp.Collateral),
				sdk.MaxDec(temp.tokenWeight(sp.Borrow), minimumBorrowFactor),
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
		weightedPairs = weightedPairs.Add(wp)
	}

	return newAccountPosition(
		mapTokens,
		weightedPairs,
		unsortedCollateralValue,
		unsortedBorrowValue,
		forLiquidation,
		minimumBorrowFactor,
	)
}

func newAccountPosition(
	tokens map[string]Token,
	pairs []WeightedSpecialPair,
	unsortedCollateralValue, unsortedBorrowValue sdk.DecCoins,
	forLiquidation bool,
	minimumBorrowFactor sdk.Dec,
) (AccountPosition, error) {
	position := AccountPosition{
		specialPairs:        WeightedSpecialPairs{},
		collateralValue:     sdk.DecCoins{},
		borrowedValue:       sdk.DecCoins{},
		isForLiquidation:    forLiquidation,
		tokens:              tokens,
		minimumBorrowFactor: minimumBorrowFactor,
	}

	for _, sp := range pairs {
		// initialize all amounts to zero. Constructing each as a new struct ensures the original
		// slice's contents cannot be modified if this position is mutated.
		wp := WeightedSpecialPair{
			Collateral:    sdk.NewDecCoinFromDec(sp.Collateral.Denom, sdk.ZeroDec()),
			Borrow:        sdk.NewDecCoinFromDec(sp.Borrow.Denom, sdk.ZeroDec()),
			SpecialWeight: sp.SpecialWeight,
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
		if wsp.Borrow.Denom == denom && unusedCollateral.IsPositive() {
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
			if cWeight.IsPositive() {
				collateralToPair = sdk.MinDec(collateralToPair,
					unusedLimit.Quo(cWeight),
				)
			}
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
// This function is fairly complex, but perfectly handles edge cases including
// those with overlapping special asset pairs.
func (ap *AccountPosition) MaxWithdraw(denom string) (sdk.Dec, bool) {
	if ap.isForLiquidation {
		// liquidation calculations should not support max withdraw
		return sdk.ZeroDec(), false
	}
	if !ap.IsHealthy() {
		// accounts over their borrow limit cannot withdraw any collateral
		return sdk.ZeroDec(), false
	}

	// first try withdrawing everything
	maxWithdraw := ap.collateralValue.AmountOf(denom)
	if ap.canHealthyWithdraw(denom, maxWithdraw) {
		// position would be healthy even after withdrawing all of this denom's collateral
		return ap.collateralValue.AmountOf(denom), true
	}

	// next try withdrawing normal collateral only
	maxWithdraw = ap.unpairedCollateral().AmountOf(denom)
	if ap.canHealthyWithdraw(denom, maxWithdraw) {
		// position would be healthy after withdrawing all unpaired collateral of this denom:
		// proceed to test each special asset pair matching this collateral
		for i := len(ap.specialPairs) - 1; i >= 0; i-- {
			sp := ap.specialPairs[i]
			// for special pairs matching collateral denom to withdraw, starting at the lowest weighted
			if sp.Collateral.Denom == denom {
				// test if the collateral in this pair can be fully withdrawn
				if !ap.canHealthyWithdraw(denom, maxWithdraw.Add(sp.Collateral.Amount)) {
					// prepare for partial withdrawal from this pair by simulating withdrawal of
					// normal collateral and any previous special pairs which were completely withdrawn
					intermediatePosition, err := newAccountPosition(
						ap.tokens,
						ap.specialPairs,
						ap.collateralValue.Sub(sdk.NewDecCoins(sdk.NewDecCoinFromDec(denom, maxWithdraw))),
						ap.borrowedValue,
						ap.isForLiquidation,
						ap.minimumBorrowFactor,
					)
					if err != nil {
						return maxWithdraw, false
					}
					// when withdrawing exclusively from a single special pair, the borrowed assets from
					// the pair are effectively borrowed against its normal collateral
					borrowToDisplace := sdk.MinDec(
						sp.Borrow.Amount,
						intermediatePosition.MaxBorrow(sp.Borrow.Denom),
					)
					// derive collateral to withdraw from displaced borrow amount
					partialWithdraw := borrowToDisplace.Quo(sp.SpecialWeight)
					// partially withdraw from this special pair in addition to completed unpaired and special withdrawals
					return maxWithdraw.Add(partialWithdraw), false
				}
				// if full withdraw was possible, add to maxWithdraw and proceed to next special pair
				maxWithdraw = maxWithdraw.Add(sp.Collateral.Amount)
			}
		}
		return maxWithdraw, true
	}
	// position would not be healthy after withdrawing all unpaired collateral of this denom
	// only calculate unpaired collateral max withdraw
	unusedLimit := ap.totalBorrowLimit().Sub(ap.BorrowedValue())            // unused borrow limit by collateral weight
	unusedCollateral := ap.CollateralValue().Sub(ap.totalCollateralUsage()) // unused collateral by borrow factor
	// - for borrow limit, withdraw subtracts [collat * weight] from borrow limit
	max1 := unusedLimit.Quo(ap.tokenWeight(denom))
	// - for borrow factor, withdraw subtracts [collat] from TC
	max2 := unusedCollateral
	// replace maxWithdraw with the lower of borrow limit and borrow factor results
	maxWithdraw = sdk.MinDec(max1, max2)
	return maxWithdraw, false
}

// canHealthyWithdraw simulates an account position with a specified amount of
// collateral withdrawn, and returns true if it is still at or below its borrow limit
func (ap *AccountPosition) canHealthyWithdraw(denom string, amount sdk.Dec) bool {
	hypotheticalPosition, err := newAccountPosition(
		ap.tokens,
		ap.specialPairs,
		ap.collateralValue.Sub(sdk.NewDecCoins(sdk.NewDecCoinFromDec(denom, amount))),
		ap.borrowedValue,
		ap.isForLiquidation,
		ap.minimumBorrowFactor,
	)
	return err == nil && hypotheticalPosition.IsHealthy()
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
