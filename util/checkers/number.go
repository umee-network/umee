package checkers

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"
)

var (
	undefinedDec sdkmath.LegacyDec
	one          = sdkmath.LegacyOneDec()
)

func IntegerMaxDiff[T constraints.Integer](a, b, maxDiff T, note string) error {
	var diff T
	if a > b {
		diff = a - b
	} else {
		diff = b - a
	}
	if diff > maxDiff {
		return fmt.Errorf("%s: diff (=%v) is too big", note, diff)
	}
	return nil
}

func NumberMin[T constraints.Integer](a, minVal T, note string) error {
	if a < minVal {
		return fmt.Errorf("%s: must be at least %v", note, minVal)
	}
	return nil
}

func NumberPositive[T constraints.Integer](a T, note string) error {
	if a <= 0 {
		return fmt.Errorf("%s: must be defined and must be positive", note)
	}
	return nil
}

func BigNumPositive[T interface{ IsPositive() bool }](a T, note string) error {
	if !a.IsPositive() {
		return fmt.Errorf("%s: must be positive", note)
	}
	return nil
}

func DecMaxDiff(a, b, maxDiff sdkmath.LegacyDec, note string) error {
	diff := a.Sub(b).Abs()
	if diff.GT(maxDiff) {
		return fmt.Errorf("%s: diff (=%v) is too big", note, diff)
	}
	return nil
}

func RequireDecMaxDiff(t *testing.T, a, b, maxDiff sdkmath.LegacyDec, note string) {
	err := DecMaxDiff(a, b, maxDiff, note)
	require.NoError(t, err)
}

// DecInZeroOne asserts that 0 <= a <= 1 when oneInclusive=True, otherwise asserts
// 0 <= a < 1
func DecInZeroOne(a sdkmath.LegacyDec, name string, oneInclusive bool) error {
	maxCheck := a.GTE
	if oneInclusive {
		maxCheck = a.GT
	}
	if a == undefinedDec || a.IsNegative() || maxCheck(one) {
		return fmt.Errorf("invalid %s: %v", name, a)
	}
	return nil
}

// DecNotNegative checks if a is defined and a >= 0
func DecNotNegative(a sdkmath.LegacyDec, paramName string) error {
	if a.IsNil() || a.IsNegative() {
		return fmt.Errorf("%s can't be negative", paramName)
	}
	return nil
}

// DecPositive checks if a is defined and a > 0
func DecPositive(a sdkmath.LegacyDec, paramName string) error {
	if a.IsNil() || !a.IsPositive() {
		return fmt.Errorf("%s must be positive", paramName)
	}
	return nil
}
