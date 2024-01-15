package bpmath

import (
	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
)

// FixedBP assures that all operations are in 0-10'000 range
// Note: FixedBP operations should not be chained - this causes precision loses.
type FixedBP uint32

// FixedFromQuo returns a/b in basis points. Returns 10'000 if a >= b;
// Contract: a>=0 and b > 0.
// Panics if b==0.
func FixedFromQuo(dividend, divisor math.Int, rounding Rounding) FixedBP {
	return FixedBP(quo(dividend, divisor, rounding, One))
}

func (bp FixedBP) ToDec() sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecWithPrec(int64(bp), 4)
}

// Mul return a*bp rounding towards zero.
func (bp FixedBP) Mul(a math.Int) math.Int {
	return Mul(a, bp)
}

// MulDec return a*bp rounding towards zero.
func (bp FixedBP) MulDec(a sdkmath.LegacyDec) sdkmath.LegacyDec {
	return MulDec(a, bp)
}
