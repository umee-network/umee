package types

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// A list of WeightedDecCoin sorted by collateral weight (descending) and denom (alphabetical) to break ties.
type WeightedDecCoins []WeightedDecCoin

// A list of WeightedSpecialPair sorted by collateral weight (descending), and denom (alphabetical) to break ties.
type WeightedSpecialPairs []WeightedSpecialPair

// A list of WeightedNormalPair sorted by collateral weight (of the collateral, then of the borrow),
// and denom (alphabetical) of the two to break ties.
type WeightedNormalPairs []WeightedNormalPair

// WeightedDecCoin holds an sdk.DecCoin representing a USD value amount of a given token denom, with
// no information on the underlying token amount. It also holds the token's collateral weight OR
// liquidation threshold, depending on usage.
type WeightedDecCoin struct {
	// the USD value of an Asset in a position
	Asset sdk.DecCoin
	// the collateral Weight or liquidation threshold
	Weight sdk.Dec
}

// WeightedSpecialPair contains borrowed and collateral value that has been matched
// together as part of a special asset pair in an account's position. The collateral
// weight OR liquidation threshold of the special pair, depending on usage, is also included.
type WeightedSpecialPair struct {
	// the collateral asset and its value
	Collateral sdk.DecCoin
	// the borrowed asset and its value
	Borrow sdk.DecCoin
	// the collateral weight (or liquidation treshold) of the special pair
	SpecialWeight sdk.Dec
}

// WeightedNormalPair contains borrowed and collateral value that has been matched together
// using regular collateral weights after special asset pairs are taken from an account's position.
type WeightedNormalPair struct {
	// the collateral asset and its weight and value
	Collateral WeightedDecCoin
	// the borrowed asset and its weight and value
	Borrow WeightedDecCoin
}

// higher returns true if a WeightedDecCoin should be sorted after
// another WeightedDecCoin b. Always use sort.SliceStable to preserve
// order of coins with equal weight.
func (wdc WeightedDecCoin) higher(b WeightedDecCoin) bool {
	if wdc.Weight.GT(b.Weight) {
		return true // sort by collateral weight, descending
	}
	return false
}

// higher returns true if a WeightedSpecialPair should be sorted after
// another WeightedSpecialPair b. Always use sort.SliceStable to preserve
// order of pairs with equal weight.
func (wsp WeightedSpecialPair) higher(b WeightedSpecialPair) bool {
	if wsp.SpecialWeight.GT(b.SpecialWeight) {
		return true // sort by special collateral weight, descending
	}
	return false
}

// higher returns true if a WeightedNormalPair should be sorted after
// another WeightedNormalPair b. Always use sort.SliceStable to preserve
// order of pairs with equal weight.
func (wnp WeightedNormalPair) higher(b WeightedNormalPair) bool {
	if wnp.Collateral.higher(b.Collateral) {
		return true // sort first by collateral
	}
	if wnp.Collateral.Weight.Equal(b.Collateral.Weight) {
		// break ties by borrow
		if wnp.Borrow.higher(b.Borrow) {
			return true
		}
	}
	return false
}

// Add returns the sum of a weightedDecCoins and an additional weightedDecCoin.
// The result is sorted.
func (wdc WeightedDecCoins) Add(add WeightedDecCoin) (sum WeightedDecCoins) {
	if len(wdc) == 0 {
		return WeightedDecCoins{add}
	}
	for _, c := range wdc {
		if c.Asset.Denom == add.Asset.Denom {
			sum = append(sum, WeightedDecCoin{
				Asset:  c.Asset.Add(add.Asset),
				Weight: c.Weight,
			})
		} else {
			sum = append(sum, c)
		}
	}
	// sorts the sum. Fixes unsorted input as well.
	sort.SliceStable(sum, func(i, j int) bool {
		return sum[i].higher(sum[j])
	})
	return sum
}

// Total returns the total USD value in a WeightedDecCoins, unaffected by collateral weight
func (wdc WeightedDecCoins) Total() sdk.Dec {
	total := sdk.ZeroDec()
	for _, c := range wdc {
		total = total.Add(c.Asset.Amount)
	}
	return total
}

// Sub subtracts a sdk.DecCoin from a WeightedDecCoins. Panics if the result would be negative.
func (wdc WeightedDecCoins) Sub(sub sdk.DecCoin) (diff WeightedDecCoins) {
	found := false
	for _, c := range wdc {
		if c.Asset.Denom == sub.Denom {
			diff = append(diff, WeightedDecCoin{
				Asset:  c.Asset.Sub(sub), // sdk.DecCoin.Sub panics on negative result
				Weight: c.Weight,
			})
			found = true
		} else {
			diff = append(diff, c)
		}
	}
	if !found {
		panic("WeightedDecCoins: sub denom not present")
	}

	// sorts the diff. Fixes unsorted input as well.
	sort.SliceStable(diff, func(i, j int) bool {
		return diff[i].higher(diff[j])
	})
	return diff
}

// Add returns the sum of a WeightedSpecialPairs and an additional WeightedSpecialPair.
// The result is sorted.
func (wsp WeightedSpecialPairs) Add(add WeightedSpecialPair) (sum WeightedSpecialPairs) {
	if len(wsp) == 0 {
		return WeightedSpecialPairs{add}
	}
	for _, wp := range wsp {
		if wp.canCombine(add) {
			sum = append(sum, WeightedSpecialPair{
				Collateral:    wp.Collateral.Add(add.Collateral),
				Borrow:        wp.Borrow.Add(add.Borrow),
				SpecialWeight: wp.SpecialWeight,
			})
		} else {
			sum = append(sum, wp)
		}
	}
	// sorts the sum. Fixes unsorted input as well.
	sort.SliceStable(sum, func(i, j int) bool {
		return sum[i].higher(sum[j])
	})
	return sum
}

// canCombine returns true if the borrow and collateral denoms of two WeightedSpecialPair are equal
func (wsp WeightedSpecialPair) canCombine(b WeightedSpecialPair) bool {
	return wsp.Collateral.Denom == b.Collateral.Denom && wsp.Borrow.Denom == b.Borrow.Denom
}
