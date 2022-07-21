package bpmath

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FixedBP assures that all oprations are in 0-10'000 range
// Note: FixedBP operations should not be chained - this causes precision loses.
type FixedBP uint32

// FixedQuo returns a/b in basis points. Returns 10'000 if a >= b;
// Contract: a>=0 and b > 0.
// Panics if b==0.
func FixedQuo(a, b sdk.Int, rounding Rounding) FixedBP {
	return FixedBP(quo(a, b, rounding, ONE))
}

func (bp FixedBP) ToDec() sdk.Dec {
	return sdk.NewDecWithPrec(int64(bp), 4)
}
