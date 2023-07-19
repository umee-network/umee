package bpmath

import (
	"cosmossdk.io/math"
)

// FixedBP assures that all operations are in 0-10'000 range
// Note: FixedBP operations should not be chained - this causes precision loses.
type FixedBP uint32

// FixedFromQuo returns a/b in basis points. Returns 10'000 if a >= b;
// Contract: a>=0 and b > 0.
// Panics if b==0.
func FixedFromQuo(dividend, divisor math.Int, rounding Rounding) FixedBP {
	return FixedBP(quo(dividend, divisor, rounding, ONE))
}

func (bp FixedBP) ToDec() math.LegacyDec {
	return math.LegacyNewDecWithPrec(int64(bp), 4)
}
