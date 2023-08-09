package checkers

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/constraints"
)

var (
	undefinedDec sdk.Dec
	one          = sdk.OneDec()
)

func IntegerMaxDiff[T constraints.Integer](a, b, maxDiff T, note string) error {
	var diff T
	if a > b {
		diff = a - b
	} else {
		diff = b - a
	}
	if diff > maxDiff {
		return fmt.Errorf("%s, diff (=%v) is too big", note, diff)
	}
	return nil
}

func DecMaxDiff(a, b, maxDiff sdk.Dec, note string) error {
	diff := a.Sub(b).Abs()
	if diff.GT(maxDiff) {
		return fmt.Errorf("%s, diff (=%v) is too big", note, diff)
	}
	return nil
}

// DecInZeroOne asserts that 0 <= a <= 1 when oneInclusive=True, otherwise asserts
// 0 <= a < 1
func DecInZeroOne(a sdk.Dec, name string, oneInclusive bool) error {
	maxCheck := a.GTE
	if oneInclusive {
		maxCheck = a.GT
	}
	if a == undefinedDec || a.IsNegative() || maxCheck(one) {
		return fmt.Errorf("invalid %s: %v", name, a)
	}
	return nil
}
