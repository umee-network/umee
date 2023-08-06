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

// demoteBorrowsAfter takes any borrows which would be sorted after an input borrowed denom
// and matches them with unpaired collateral, and then ordinary collateral starting at the
// lowest in the list. And freed up collateral is moved to the position's unpaired collateral,
// where it can be used by other operations such as fillOrdinaryCollateral.
func (ap *AccountPosition) demoteBorrowsAfter(denom string) {
	//
	//
	//
	//
	//
}