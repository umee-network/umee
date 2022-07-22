package bpmath

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BP assures represents values in basis points. Maximum value is 2^32-1.
// Note: BP operations should not be chained - this causes precision loses.
type BP uint32

// Quo returns a/b in basis points.
// Contract: a>=0 and b > 0.
// Panics if a/b >= MaxUint32/10'000 or if b==0.
func Quo(a, b sdk.Int, rounding Rounding) BP {
	return BP(quo(a, b, rounding, math.MaxUint32))
}

func quo(a, b sdk.Int, rounding Rounding, max uint64) uint64 {
	if b.IsZero() {
		panic("divider can't be zero")
	}
	bp := a.MulRaw(ONE)
	if rounding == UP {
		bp = bp.Add(b.SubRaw(1))
	}
	x := bp.Quo(b).Uint64()
	if x > max {
		panic("basis points out of band")
	}
	return x
}

// Mul returns a * b_in_basis_points
// Contract: b \in [0; MatxUint32]
func Mul[T BP | FixedBP](a sdk.Int, b T) sdk.Int {
	if b == 0 {
		return sdk.ZeroInt()
	}
	if b == ONE {
		return a
	}
	return a.MulRaw(int64(b)).Quo(oneBigInt)
}

func (bp BP) ToDec() sdk.Dec {
	return sdk.NewDecWithPrec(int64(bp), 4)
}
