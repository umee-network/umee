package types

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// A list of WeightedDecCoin sorted by collateral weight (descending), then liquidation threshold
// (descending), and finally denom (alphabetical) to break ties.
type WeightedDecCoins []WeightedDecCoin

// A list of WeightedSpecialPair sorted by collateral weight (descending), then liquidation threshold
// (descending), and finally denom (alphabetical) to break ties.
type WeightedSpecialPairs []WeightedSpecialPair

// WeightedDecCoin holds an sdk.DecCoin representing a USD value amount of a given token denom, with
// no information on the underlying token amount. It also holds the token's collateral weight and
// liquidation threshold.
type WeightedDecCoin struct {
	// the USD value of an Asset in a position
	Asset sdk.DecCoin
	// the collateral Weight
	Weight sdk.Dec
	// the Liquidation threshold
	Liquidation sdk.Dec
}

// WeightedSpecialPair contains borrowed and collateral value that has been matched
// together as part of a special asset pair in an account's position. The parameters
// of the special asset pair are also included.
type WeightedSpecialPair struct {
	// the Collateral asset and its value
	Collateral sdk.DecCoin
	// the borrowed asset and its value
	Borrow sdk.DecCoin
	// the collateral weight of the special pair
	SpecialWeight sdk.Dec
	// the liquidation threshold of the special pair
	LiquidationThreshold sdk.Dec
}

// higher returns true if a WeightedDecCoin wdc should be sorted after
// another WeightedDecCoin b
func (wdc WeightedDecCoin) higher(b WeightedDecCoin) bool {
	if wdc.Weight.GT(b.Weight) {
		return true // sort first by collateral weight
	}
	if wdc.Liquidation.GT(b.Liquidation) {
		return true // sort next by liquidation threshold
	}
	return wdc.Asset.Denom < b.Asset.Denom // break ties by denom
}

// higher returns true if a WeightedSpecialPair wdc should be sorted after
// another WeightedSpecialPair b
func (wsp WeightedSpecialPair) higher(b WeightedSpecialPair) bool {
	if wsp.SpecialWeight.GT(b.SpecialWeight) {
		return true // sort first by collateral weight
	}
	if wsp.LiquidationThreshold.GT(b.LiquidationThreshold) {
		return true // sort next by liquidation threshold
	}
	if wsp.Collateral.Denom < b.Collateral.Denom {
		return true // break ties by collateral denom first
	}
	return wsp.Borrow.Denom < b.Borrow.Denom // then by borrow denom
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
				Asset:       c.Asset.Add(add.Asset),
				Weight:      c.Weight,
				Liquidation: c.Liquidation,
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
				Asset:       c.Asset.Sub(sub), // sdk.DecCoin.Sub panics on negative result
				Weight:      c.Weight,
				Liquidation: c.Liquidation,
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
				Collateral:           wp.Collateral.Add(add.Collateral),
				Borrow:               wp.Borrow.Add(add.Borrow),
				SpecialWeight:        wp.SpecialWeight,
				LiquidationThreshold: wp.LiquidationThreshold,
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
