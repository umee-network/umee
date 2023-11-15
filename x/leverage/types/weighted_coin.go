package types

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// A list of WeightedSpecialPair sorted by collateral weight (descending), and denom (alphabetical) to break ties.
type WeightedSpecialPairs []WeightedSpecialPair

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

func (wsp WeightedSpecialPair) Validate() error {
	if wsp.SpecialWeight.IsNil() || !wsp.SpecialWeight.IsPositive() {
		return fmt.Errorf("invalid special weight: %s", wsp.SpecialWeight)
	}
	if err := wsp.Collateral.Validate(); err != nil {
		return err
	}
	return wsp.Borrow.Validate()
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

// Add returns the sum of a WeightedSpecialPairs and an additional WeightedSpecialPair.
// The result is sorted. Zero amount pairs are not omitted.
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

// canCombine returns true if the borrow and collateral denoms of two WeightedSpecialPair are equal
func (wsp WeightedSpecialPair) canCombine(b WeightedSpecialPair) bool {
	return wsp.Collateral.Denom == b.Collateral.Denom && wsp.Borrow.Denom == b.Borrow.Denom
}
