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
	borrowFactor := ap.borrowFactor(denom)
	if ap.isForLiquidation || !borrowFactor.IsPositive() {
		return sdk.ZeroDec()
	}

	limit := ap.totalBorrowLimit()     // borrow limit after special pairs
	usage := ap.totalCollateralUsage() // collateral usage after special pairs

	maxNormalBorrow := sdk.MinDec(
		// limited by collateral weight: borrow up to remaining limit
		limit.Sub(ap.BorrowedValue()),
		// limited by borrow factor: borrow up to unused collateral / borrow factor
		ap.CollateralValue().Sub(usage).Quo(
			borrowFactor,
		),
	)

	//
	// TODO: improve for cases where new special assets can be paired
	//

	// Prevent over-limit accounts from returning negative max borrow
	return sdk.MaxDec(sdk.ZeroDec(), maxNormalBorrow)
}

// MaxWithdraw finds the maximum additional amount of an asset a position can
// withdraw without exceeding its borrow limit. Does not mutate position.
// Returns zero if a position was computed with liquidation in mind.
func (ap *AccountPosition) MaxWithdraw(denom string) sdk.Dec {
	if ap.isForLiquidation {
		return sdk.ZeroDec()
	}

	limit := ap.totalBorrowLimit()     // borrow limit after special pairs
	usage := ap.totalCollateralUsage() // collateral usage after special pairs
	owned := ap.collateralValue.AmountOf(denom)

	collateralWeight := ap.tokenWeight(denom)
	if !collateralWeight.IsPositive() {
		return owned
	}

	//
	// TODO: withdraw first from normal, then from special pairs, one at a time.
	//

	// - for borrow limit, subtracting [collat * weight] from borrow limit
	//		- TODO: for special pairs, subtracting additional [collateral * delta weight]
	unusedLimit := limit.Sub(ap.BorrowedValue())
	max1 := unusedLimit.Quo(collateralWeight)

	// - for borrow factor, subtracting [collat] from TC
	//		- TODO: for special pairs, adding additional [borrow * delta factor] to collateral usage
	unusedCollateral := ap.CollateralValue().Sub(usage)
	max2 := unusedCollateral

	maxWithdraw := sdk.MinDec(max1, max2)                // lower of borrow limit anf borrow factor results
	maxWithdraw = sdk.MinDec(maxWithdraw, owned)         // capped at owned amount
	maxWithdraw = sdk.MaxDec(maxWithdraw, sdk.ZeroDec()) // prevent negative value
	return maxWithdraw
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

	// compute limit due to borrow factors
	usage := ap.totalCollateralUsage()
	avgWeight := ap.normalBorrowLimit().Quo(collateralValue)
	unusedCollateralValue := ap.CollateralValue().Sub(usage) // can be negative
	borrowFactorLimit := ap.BorrowedValue().Add(unusedCollateralValue.Mul(avgWeight))

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

// totalCollateralUsage computes normalCollateralUsage and then applies the effects of special asset pairs.
func (ap *AccountPosition) totalCollateralUsage() sdk.Dec {
	usage := ap.normalCollateralUsage()
	for _, wsp := range ap.specialPairs {
		usage = usage.Sub(ap.collateralUsageDecrease(wsp))
	}
	return usage
}

// normalCollateralUsage sums the total borrowed value in a position,
// increased according to each token's borrow factor (collateral weight or liquidation threshold),
// or ap.minimumBorrowFactor if greater. Does not use special asset weights for paired assets.
// The resulting value is the total collateral value which would be required to support all
// borrowed assets, without any special asset pairs being applied.
func (ap *AccountPosition) normalCollateralUsage() sdk.Dec {
	sum := sdk.ZeroDec()
	for _, b := range ap.borrowedValue {
		sum = sum.Add(
			b.Amount.Quo(sdk.MaxDec(
				ap.tokenWeight(b.Denom),
				ap.minimumBorrowFactor,
			)),
		)
	}
	return sum
}

// totalBorrowLimit computes normalBorrowLimit and then applies the effects of special asset pairs.
func (ap *AccountPosition) totalBorrowLimit() sdk.Dec {
	limit := ap.normalBorrowLimit()
	for _, wsp := range ap.specialPairs {
		limit = limit.Add(ap.borrowLimitIncrease(wsp))
	}
	return limit
}

// normalBorrowLimit sums the total collateral value in a position,
// reduced according to each token's collateral weight or liquidation threshold.
// Does not use special asset weights for paired assets.
// The resulting value is the total borrowed value which could be supported by
// these collateral assets, without any special asset pairs being applied.
func (ap *AccountPosition) normalBorrowLimit() sdk.Dec {
	sum := sdk.ZeroDec()
	for _, b := range ap.collateralValue {
		sum = sum.Add(b.Amount.Mul(ap.tokenWeight(b.Denom)))
	}
	return sum
}

// borrowLimitIncrease calculates the amount above an account's normalBorrowLimit
// it is allowed to borrow due to the effects of a single special asset pair.
func (ap *AccountPosition) borrowLimitIncrease(wsp WeightedSpecialPair) sdk.Dec {
	additionalWeight := sdk.MaxDec(
		// collateral weight (or liquidation threshold) is increased if the
		// special pair's weight is greater than that of the collateral token
		wsp.SpecialWeight.Sub(ap.tokenWeight(wsp.Collateral.Denom)),
		sdk.ZeroDec(), // prevent negative effects
	)
	// the increase in borrow limit is each affected collateral amount times
	// the additional weight.
	return wsp.Collateral.Amount.Mul(additionalWeight)
}

// collateralUsageDecrease calculates amount below an account's normalCollateralUsage
// its effective utilization is due to the effects of existing special asset pairs.
func (ap *AccountPosition) collateralUsageDecrease(wsp WeightedSpecialPair) sdk.Dec {
	// initial borrow factor comes from token settings (and minimum)
	borrowFactor := ap.borrowFactor(wsp.Borrow.Denom)
	// ignore negative effects
	if borrowFactor.GTE(wsp.SpecialWeight) {
		return sdk.ZeroDec()
	}
	// decreases effective collateral usage due to the difference in parameters
	return wsp.Borrow.Amount.Quo(borrowFactor).Sub( // original usage
		wsp.Borrow.Amount.Quo(wsp.SpecialWeight)) // special usage
}
