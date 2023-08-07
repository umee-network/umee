package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// fillOrdinaryCollateral finds all unused collateral assets in a position
// and borrows the maximum amount of an input denom against them. Does not
// interact with special asset pairs or move borrows around between collateral,
// so must be called only after a position has been rearranged to accomodate
// a max borrow. Returns the amount of borrows added. The account position
// is mutated to include the new borrows, and will be at its borrow limit.
// If the requested token denom did not exist or the borrower was already
// at or over their borrow limit, this is a no-op which returns zero.
func (ap *AccountPosition) fillOrdinaryCollateral(denom string) sdk.Dec {
	if len(ap.unpairedCollateral) == 0 {
		return sdk.ZeroDec()
	}
	if !ap.hasToken(denom) {
		return sdk.ZeroDec()
	}
	borrowFactor := sdk.MaxDec(
		minimumBorrowFactor,
		ap.tokenWeight(denom),
	)
	total := sdk.ZeroDec()
	// ignores collateral with weight of zero
	ineligible := WeightedDecCoins{}
	// converts unpaired collateral into normal asset pairs with new borrow
	for i, uc := range ap.unpairedCollateral {
		weight := sdk.MinDec(uc.Weight, borrowFactor)
		if weight.IsPositive() {
			bCoin := sdk.NewDecCoinFromDec(denom, uc.Asset.Amount.Mul(weight))
			ap.normalPairs = ap.normalPairs.Add(WeightedNormalPair{
				Collateral: WeightedDecCoin{
					Asset:  uc.Asset,
					Weight: ap.tokenWeight(uc.Asset.Denom),
				},
				Borrow: WeightedDecCoin{
					Asset:  bCoin,
					Weight: ap.tokenWeight(denom),
				},
			})
			// tracks how much was borrowed
			total = total.Add(bCoin.Amount)
			// clears unpaired collateral which has now been borrowed against
			ap.unpairedCollateral[i].Asset.Amount = sdk.ZeroDec()
		} else {
			ineligible = ineligible.Add(uc)
		}
	}
	// the only remaining unpaired collateral is that which cannot be borrowed against
	ap.unpairedCollateral = ineligible
	return total
}

// displaceBorrowsAfter takes any borrows which would be sorted after an input borrowed denom
// and matches them with unpaired collateral, and then ordinary collateral starting at the
// lowest in the list. And freed up collateral is moved to the position's unpaired collateral,
// where it can be used by other operations such as fillOrdinaryCollateral.
func (ap *AccountPosition) displaceBorrowsAfter(denom string) error {
	if len(ap.normalPairs) == 0 || len(ap.unpairedBorrows) > 0 {
		// no-op if there are no normal assets to sort or if the borrower is over limit
		return nil
	}
	// all of the position's borrows and collateral will be rearranged
	normalPairs := WeightedNormalPairs{}
	unpairedCollateral := WeightedDecCoins{}
	unpairedBorrows := WeightedDecCoins{}
	// collateral and borrows are taken from normal pairs
	for _, np := range ap.normalPairs {
		// if the borrow in each ordinary pair should be displaced
		if !np.Borrow.before(WeightedDecCoin{
			Asset:  sdk.NewDecCoin(denom, sdk.ZeroInt()),
			Weight: ap.tokenWeight(denom),
		}) {
			// displaced normal asset pairs are taken apart
			unpairedBorrows = unpairedBorrows.Add(np.Borrow)
			unpairedCollateral = unpairedCollateral.Add(np.Collateral)
		} else {
			// non-displaced normal asset pairs stay the same
			normalPairs = normalPairs.Add(np)
		}
	}
	// initial unpaired collateral is also added to the pre-sorted pool of collateral to pair
	for _, c := range ap.unpairedCollateral {
		unpairedCollateral = unpairedCollateral.Add(c)
	}
	// match assets into normal asset pairs, removing matched value from sortedCollateral and sortedBorrows
	// unlike NewAccountPosition, this prioritizes the lowest weight borrows and collateral first
	i := len(unpairedCollateral) - 1
	j := len(unpairedBorrows) - 1
	// TODO: refactor mostly duplicate code from NewAccountPosition (only the iteration order is reversed here)
	for i >= 0 && j >= 0 {
		cDenom := unpairedCollateral[i].Asset.Denom
		bDenom := unpairedBorrows[j].Asset.Denom
		c := unpairedCollateral[i].Asset.Amount
		b := unpairedBorrows[j].Asset.Amount
		w := sdk.MinDec(
			// for normal asset pairs, both tokens limit the collateral weight of the pair
			ap.tokenWeight(cDenom),
			sdk.MaxDec(ap.tokenWeight(bDenom), minimumBorrowFactor),
		)
		// match collateral and borrow at indexes i and j, exhausting at least one of them
		pairBorrowLimit := c.Mul(w)
		var bCoin, cCoin sdk.DecCoin
		if pairBorrowLimit.LTE(b) {
			// all of the collateral is used (note: this case includes collateral with zero weight)
			cCoin = sdk.NewDecCoinFromDec(cDenom, c)
			// only some of the borrow, equal to collateral value * collateral weight is covered
			bCoin = sdk.NewDecCoinFromDec(bDenom, pairBorrowLimit)
			// next collateral
			i--
		} else {
			// some collateral, equal to borrow value / collateral weight, is used
			cCoin = sdk.NewDecCoinFromDec(cDenom, b.Quo(w))
			// all of the borrow is covered by collateral in this pair
			bCoin = sdk.NewDecCoinFromDec(bDenom, b)
			// next borrow
			j--
		}

		// skip zero positions.
		if cCoin.IsPositive() || bCoin.IsPositive() {
			// subtract newly paired assets from unsorted assets
			unpairedBorrows = unpairedBorrows.Sub(bCoin)
			unpairedCollateral = unpairedCollateral.Sub(cCoin)
			// create a normal asset pair and add it to the account position
			normalPairs = normalPairs.Add(WeightedNormalPair{
				Collateral: WeightedDecCoin{
					Asset:  cCoin,
					Weight: ap.tokenWeight(cDenom),
				},
				Borrow: WeightedDecCoin{
					Asset:  bCoin,
					Weight: ap.tokenWeight(bDenom),
				},
			})
		}
	}

	// now there should be unpaired collateral left over, but no unpaired borrows.
	// all normal pairs have been created in their new order.

	// overwrite the position's affected assets
	ap.normalPairs = normalPairs
	// overwrites unpaired collateral with newly freed assets from displacement
	ap.unpairedCollateral = WeightedDecCoins{}
	for _, cv := range unpairedCollateral {
		if cv.Asset.IsPositive() {
			// sort collateral by collateral weight (or liquidation threshold) using Add function
			ap.unpairedCollateral = ap.unpairedCollateral.Add(cv)
		}
	}
	// any remaining borrows could not be paired (should not occur)
	for _, bv := range unpairedBorrows {
		if bv.Asset.IsPositive() {
			return ErrInvalidPosition.Wrap("borrow position over limit following displaceBorrowsAfter")
		}
	}
	return nil
}

// TODO: displaceBorrowsFrom(denom)
// used for max withdraw from ordinary
// ""fromSpecial separate function?
