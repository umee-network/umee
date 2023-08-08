package types

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: unit-test this file

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

// before returns true if a WeightedDecCoin should be sorted before
// another WeightedDecCoin b. It should be used as a sort.Less
func (wdc WeightedDecCoin) before(b WeightedDecCoin) bool {
	if !wdc.Weight.Equal(b.Weight) {
		// sort by collateral weight, descending
		return wdc.Weight.GT(b.Weight)
	}
	// for the same weight, sort by asset denom, alphabetical
	return wdc.Asset.Denom < b.Asset.Denom
}

// before returns true if a WeightedNormalPair should be sorted before
// another WeightedNormalPair b. It should be used as a sort.Less
func (wnp WeightedNormalPair) before(b WeightedNormalPair) bool {
	if !wnp.Collateral.Weight.Equal(b.Collateral.Weight) {
		// first sort by collateral weight, descending
		return wnp.Collateral.Weight.GT(b.Collateral.Weight)
	}
	if !wnp.Borrow.Weight.Equal(b.Borrow.Weight) {
		// if collateral weights are the same, sort by borrow weight, descending
		return wnp.Borrow.Weight.GT(b.Borrow.Weight)
	}
	if wnp.Collateral.Asset.Denom != b.Collateral.Asset.Denom {
		// if both weight are equal, sort by collateral denom, alphabetical
		return wnp.Collateral.Asset.Denom != b.Collateral.Asset.Denom
	}
	// if all of the above were equal, sort by borrow denom, alphabetical
	return wnp.Borrow.Asset.Denom < b.Borrow.Asset.Denom
}

// before returns true if a WeightedSpecialPair should be sorted before
// another WeightedSpecialPair b. It should be used as a sort.Less
func (wsp WeightedSpecialPair) before(b WeightedSpecialPair) bool {
	if !wsp.SpecialWeight.Equal(b.SpecialWeight) {
		// sort first by special collateral weight, descending
		return wsp.SpecialWeight.GT(b.SpecialWeight)
	}
	if wsp.Collateral.Denom != b.Collateral.Denom {
		// for the same weight, sort by collateral denom, alphabetical
		return wsp.Collateral.Denom < b.Collateral.Denom
	}
	// for the same collateral denom and weight, sort by borrow denom, alphabetical
	return wsp.Borrow.Denom < b.Borrow.Denom
}

// Add returns the sum of a weightedDecCoins and an additional weightedDecCoin.
// The result is sorted.
func (wdc WeightedDecCoins) Add(add WeightedDecCoin) (sum WeightedDecCoins) {
	if len(wdc) == 0 {
		return WeightedDecCoins{add}
	}
	found := false
	for _, c := range wdc {
		if c.Asset.Denom == add.Asset.Denom {
			sum = append(sum, WeightedDecCoin{
				Asset:  c.Asset.Add(add.Asset),
				Weight: c.Weight,
			})
			found = true
		} else {
			sum = append(sum, c)
		}
	}
	if !found {
		sum = append(sum, add)
	}
	// sorts the sum. Fixes unsorted input as well.
	sort.SliceStable(sum, func(i, j int) bool {
		return sum[i].before(sum[j])
	})
	return sum
}

// Total returns the total USD value in a WeightedDecCoins, unaffected by collateral weight.
// If denom is not empty, returns a the amount of that denom.
func (wdc WeightedDecCoins) Total(denom string) sdk.Dec {
	total := sdk.ZeroDec()
	for _, c := range wdc {
		if denom == "" || c.Asset.Denom == denom {
			total = total.Add(c.Asset.Amount)
		}
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
		return diff[i].before(diff[j])
	})
	return diff
}

// Add returns the sum of a WeightedSpecialPairs and an additional WeightedSpecialPair.
// The result is sorted.
func (wsp WeightedSpecialPairs) Add(add WeightedSpecialPair) (sum WeightedSpecialPairs) {
	if len(wsp) == 0 {
		return WeightedSpecialPairs{add}
	}
	found := false
	for _, wp := range wsp {
		if wp.canCombine(add) {
			sum = append(sum, WeightedSpecialPair{
				Collateral:    wp.Collateral.Add(add.Collateral),
				Borrow:        wp.Borrow.Add(add.Borrow),
				SpecialWeight: wp.SpecialWeight,
			})
			found = true
		} else {
			sum = append(sum, wp)
		}
	}
	if !found {
		sum = append(sum, add)
	}
	// sorts the sum. Fixes unsorted input as well.
	sort.SliceStable(sum, func(i, j int) bool {
		return sum[i].before(sum[j])
	})
	return sum
}

// Add returns the sum of a WeightedNormalPairs and an additional WeightedNormalPair.
// The result is sorted.
func (wnp WeightedNormalPairs) Add(add WeightedNormalPair) (sum WeightedNormalPairs) {
	if len(wnp) == 0 {
		return WeightedNormalPairs{add}
	}
	found := false
	for _, wp := range wnp {
		if wp.canCombine(add) {
			sum = append(sum, WeightedNormalPair{
				Collateral: WeightedDecCoin{
					Asset:  wp.Collateral.Asset.Add(add.Collateral.Asset),
					Weight: wp.Collateral.Weight,
				},
				Borrow: WeightedDecCoin{
					Asset:  wp.Borrow.Asset.Add(add.Borrow.Asset),
					Weight: wp.Borrow.Weight,
				},
			})
			found = true
		} else {
			sum = append(sum, wp)
		}
	}
	if !found {
		sum = append(sum, add)
	}
	// sorts the sum. Fixes unsorted input as well.
	sort.SliceStable(sum, func(i, j int) bool {
		return sum[i].before(sum[j])
	})
	return sum
}

// canCombine returns true if the borrow and collateral denoms of two WeightedSpecialPair are equal
func (wsp WeightedSpecialPair) canCombine(b WeightedSpecialPair) bool {
	return wsp.Collateral.Denom == b.Collateral.Denom && wsp.Borrow.Denom == b.Borrow.Denom
}

// canCombine returns true if the borrow and collateral denoms of two WeightedNormalPair are equal
func (wnp WeightedNormalPair) canCombine(b WeightedNormalPair) bool {
	return wnp.Collateral.Asset.Denom == b.Collateral.Asset.Denom && wnp.Borrow.Asset.Denom == b.Borrow.Asset.Denom
}
