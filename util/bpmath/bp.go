package bpmath

import (
	"math"

	sdkmath "cosmossdk.io/math"
)

// BP represents values in basis points. Maximum value is 2^32-1.
// Note: BP operations should not be chained - this causes precision losses.
type BP uint32

func (bp BP) ToDec() sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecWithPrec(int64(bp), 4)
}

// Mul return a*bp rounding towards zero.
func (bp BP) Mul(a sdkmath.Int) sdkmath.Int {
	return Mul(a, bp)
}

// MulDec return a*bp rounding towards zero.
func (bp BP) MulDec(a sdkmath.LegacyDec) sdkmath.LegacyDec {
	return MulDec(a, bp)
}

// FromQuo returns a/b in basis points.
// Contract: a>=0 and b > 0.
// Panics if a/b >= MaxUint32/10'000 or if b==0.
func FromQuo(dividend, divisor sdkmath.Int, rounding Rounding) BP {
	return BP(quo(dividend, divisor, rounding, math.MaxUint32))
}

func quo(a, b sdkmath.Int, rounding Rounding, max uint64) uint64 {
	if b.IsZero() {
		panic("divider can't be zero")
	}
	bp := a.MulRaw(One)
	if rounding == UP {
		bp = bp.Add(b.SubRaw(1))
	}
	x := bp.Quo(b).Uint64()
	if x > max {
		panic("basis points out of band")
	}
	return x
}

// Mul returns a * b_basis_points rounding towards zero.
// Contract: b in [0, MaxUint32]
func Mul[T BP | FixedBP | int](a sdkmath.Int, b T) sdkmath.Int {
	if b == 0 {
		return sdkmath.ZeroInt()
	}
	if b == One {
		return a
	}
	return a.MulRaw(int64(b)).Quo(oneBigInt)
}

// MulDec returns a * b_basis_points rounding towards zero.
// Contract: b in [0, MaxUint32]
func MulDec[T BP | FixedBP | int](a sdkmath.LegacyDec, b T) sdkmath.LegacyDec {
	if b == 0 {
		return sdkmath.LegacyZeroDec()
	}
	if b == One {
		return a
	}
	return a.MulInt64(int64(b)).Quo(oneDec)
}
