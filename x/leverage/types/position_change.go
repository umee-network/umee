package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: unit-test this file

// fillOrdinaryCollateral finds all unused collateral assets in a position
// and borrows the maximum amount of an input denom against them. Does not
// interact with special asset pairs or move borrows around between collateral,
// so must be called only after a position has been rearranged to accommodate
// a max borrow. Returns the amount of borrows added. The account position
// is mutated to include the new borrows, and will be at its borrow limit.
// If the requested token denom did not exist or the borrower was already
// at or over their borrow limit, this is a no-op which returns zero.
// Also accepts a maximum amount of asset to borrow, which should be set
// to the position's total collateral value for max borrow.
func (ap *AccountPosition) fillOrdinaryCollateral(denom string, max sdk.Dec) sdk.Dec {
	if len(ap.unpairedCollateral) == 0 {
		return sdk.ZeroDec()
	}
	if !ap.hasToken(denom) {
		return sdk.ZeroDec()
	}
	borrowFactor := sdk.MaxDec(
		ap.minimumBorrowFactor,
		ap.tokenWeight(denom),
	)
	newBorrow := sdk.ZeroDec()
	// converts unpaired collateral into normal asset pairs with new borrow
	for i, uc := range ap.unpairedCollateral {
		weight := sdk.MinDec(uc.Weight, borrowFactor)
		if weight.IsPositive() && newBorrow.LT(max) {
			cCoin := uc.Asset
			bCoin := sdk.NewDecCoinFromDec(denom, uc.Asset.Amount.Mul(weight))
			remainingToBorrow := max.Sub(newBorrow)
			if bCoin.Amount.GT(remainingToBorrow) {
				// when partially borrowing for this collateral will reach max
				bCoin.Amount = bCoin.Amount.Mul(remainingToBorrow)
				cCoin.Amount = cCoin.Amount.Mul(remainingToBorrow.Quo(bCoin.Amount))
			}
			// create a normal pair with a new borrow and some previously unpaired collateral
			ap.normalPairs = ap.normalPairs.Add(WeightedNormalPair{
				Collateral: WeightedDecCoin{
					Asset:  cCoin,
					Weight: ap.tokenWeight(uc.Asset.Denom),
				},
				Borrow: WeightedDecCoin{
					Asset:  bCoin,
					Weight: ap.tokenWeight(denom),
				},
			})
			// tracks how much was borrowed, and adds it to position
			newBorrow = newBorrow.Add(bCoin.Amount)
			ap.borrowedValue = newBorrow.Add(bCoin.Amount)
			// reduces unpaired collateral by amount that was moved to normal pair
			ap.unpairedCollateral[i].Asset = uc.Asset.Sub(cCoin)
		}
	}
	ap.sortNormalAssets()
	return newBorrow
}

// displaceBorrowsAfterBorrowDenom takes any borrows which would be sorted after an input borrowed denom
// and matches them with unpaired collateral, and then ordinary collateral starting at the
// lowest in the list. And freed up collateral is moved to the position's unpaired collateral,
// where it can be used by other operations such as fillOrdinaryCollateral. Note that due to the
// displacement of assets from their normal order, the account position is not sorted.
func (ap *AccountPosition) displaceBorrowsAfterBorrowDenom(denom string) error {
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
		if np.Borrow.before(WeightedDecCoin{
			Asset:  sdk.NewDecCoin(denom, sdk.ZeroInt()),
			Weight: ap.tokenWeight(denom),
		}) {
			// non-displaced normal asset pairs stay the same
			normalPairs = normalPairs.Add(np)
		} else {
			// displaced normal asset pairs are taken apart
			unpairedBorrows = unpairedBorrows.Add(np.Borrow)
			unpairedCollateral = unpairedCollateral.Add(np.Collateral)
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
			sdk.MaxDec(ap.tokenWeight(bDenom), ap.minimumBorrowFactor),
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
	// unpaired collateral is not necessarily of lower weight than paired collateral anymore.
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
			return fmt.Errorf("borrow position over limit following displaceBorrowsAfter(%s)", denom)
		}
	}
	return nil
}

// withdrawNormalCollateral attempts to displace as many borrowed assets as possible away from
// normal pairs with specified collateral. There are two cases: one where the entire collateral amount of
// the input denom is freed up, and another where a partial amount must remain paired with existing
// borrows. Returns the value withdrawn and any errors. The account position is then sorted to fix
// the order of normal assets that were not withdrawn.
func (ap *AccountPosition) withdrawNormalCollateral(denom string) (sdk.Dec, error) {
	if len(ap.unpairedCollateral) == 0 || len(ap.unpairedBorrows) > 0 {
		// no-op if there are no normal assets to sort or if the borrower is over limit
		return sdk.ZeroDec(), nil
	}
	// all of the position's borrows and collateral will be rearranged
	normalPairs := WeightedNormalPairs{}
	unpairedCollateral := WeightedDecCoins{}
	unpairedBorrows := WeightedDecCoins{}
	// collateral and borrows are taken from normal pairs
	for _, np := range ap.normalPairs {
		if np.Collateral.before(WeightedDecCoin{
			Asset:  sdk.NewDecCoin(denom, sdk.ZeroInt()),
			Weight: ap.tokenWeight(denom),
		}) {
			// non-displaced normal asset pairs stay the same
			normalPairs = normalPairs.Add(np)
		} else {
			// displaced normal asset pairs are taken apart
			unpairedBorrows = unpairedBorrows.Add(np.Borrow)
			unpairedCollateral = unpairedCollateral.Add(np.Collateral)
		}
	}
	// initial unpaired collateral is also added to the pool of collateral to pair
	for _, c := range ap.unpairedCollateral {
		unpairedCollateral = unpairedCollateral.Add(c)
	}
	// match assets into normal asset pairs, removing matched value from sortedCollateral and sortedBorrows
	// unlike NewAccountPosition, this prioritizes the lowest weight borrows and collateral first
	i := len(unpairedCollateral) - 1
	j := len(unpairedBorrows) - 1
	// TODO: refactor mostly duplicate code from sortNormalAssets (only the iteration order is reversed here)
	for i >= 0 && j >= 0 {
		cDenom := unpairedCollateral[i].Asset.Denom
		bDenom := unpairedBorrows[j].Asset.Denom
		c := unpairedCollateral[i].Asset.Amount
		b := unpairedBorrows[j].Asset.Amount
		w := sdk.MinDec(
			// for normal asset pairs, both tokens limit the collateral weight of the pair
			ap.tokenWeight(cDenom),
			sdk.MaxDec(ap.tokenWeight(bDenom), ap.minimumBorrowFactor),
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
			// subtract newly paired assets from unpaired assets
			unpairedBorrows = unpairedBorrows.Sub(bCoin)
			unpairedCollateral = unpairedCollateral.Sub(cCoin)
			// create a normal asset pair and add it to the new list of normal pairs
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
	// the maximum amount of input denom collateral has been moved to unpaired collateral,
	// but any additional unpaired collateral is also abnormal, having been paired from lowest
	// to highest. all normal pairs have been created in their new order.

	// overwrite the position's affected assets
	ap.normalPairs = normalPairs
	// overwrites unpaired collateral with newly freed assets from displacement
	ap.unpairedCollateral = WeightedDecCoins{}
	withdrawn := sdk.ZeroDec()
	for _, cv := range unpairedCollateral {
		if cv.Asset.IsPositive() {
			if cv.Asset.Denom == denom {
				// withdrawn collateral is not added back to the position
				withdrawn = withdrawn.Add(cv.Asset.Amount)
				ap.collateralValue = ap.collateralValue.Sub(cv.Asset.Amount)
			} else {
				// sort remaining unpaired collateral using Add function
				ap.unpairedCollateral = ap.unpairedCollateral.Add(cv)
			}
		}
	}
	// any remaining borrows could not be paired (should not occur)
	for _, bv := range unpairedBorrows {
		if bv.Asset.IsPositive() {
			return sdk.ZeroDec(), fmt.Errorf(
				"borrow position over limit following displaceBorrowsFromCollateralDenom(%s)", denom,
			)
		}
	}

	// fix the order of the collateral which was shuffled around due to withdrawal
	ap.sortNormalAssets()
	return withdrawn, ap.Validate()
}

// withdrawFromSpecialPair attempts to displace as many borrowed assets from a given special
// asset pair as possible. This is used to free up the collateral in that pair so that it may be
// withdrawn. Displaced borrows must be absorbed by normal collateral. Returns the amount of
// collateral removed from the pair and an error. Special pair to withdraw from is identified by
// its index in AccountPosition (which will not change even if collateral is completely withdrawn.)
func (ap *AccountPosition) withdrawFromSpecialPair(index int) (sdk.Dec, error) {
	// General steps:
	// 1) max borrow from normal assets with a cap of this pair's borrow amount
	// 2) subtract borrwed amount and equivalent collateral from pair
	// 3) collateral amount is returned, after being subtracted from total value
	if len(ap.normalPairs) == 0 || len(ap.unpairedBorrows) > 0 {
		// no-op if there are no normal assets to sort or if the borrower is over limit
		return sdk.ZeroDec(), nil
	}
	sp := ap.specialPairs[index]
	// rearrange normal assets such that borrows which are lower weight than the
	// borrow denom are pushed below unpaired collateral, and any collateral
	// which can be used to borrow the that denom becomes the new unpaired
	err := ap.displaceBorrowsAfterBorrowDenom(sp.Borrow.Denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	// borrow against all remaining unpaired collateral until collateral exhausted or
	// the special pair's borrow amount is reached
	borrowed := ap.fillOrdinaryCollateral(sp.Borrow.Denom, sp.Borrow.Amount)
	withdrawn := borrowed.Quo(sp.SpecialWeight)
	// remove borrowed assets and withdrawn collateral from special asset pair
	// note that with the new borrow (from fillOrdinaryCollateral) increasing and
	// the special borrow decreasing, this position's total borrowed did not change
	ap.specialPairs[index].Collateral.Amount = sp.Collateral.Amount.Sub(withdrawn)
	ap.specialPairs[index].Borrow.Amount = sp.Borrow.Amount.Sub(borrowed)
	ap.borrowedValue = ap.borrowedValue.Sub(borrowed)
	ap.collateralValue = ap.collateralValue.Sub(withdrawn)
	return withdrawn, ap.Validate()
}
