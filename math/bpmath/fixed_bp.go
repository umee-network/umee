package bpmath

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FixedBP assures that all oprations are in 0-10'000 range
type FixedBP uint32

// FixedQuo returns a/b in basis points. Returns 10'000 if a >= b or b==0;
// Contract: a>=0 and b >= 0.
func FixedQuo(a, b sdk.Int, rounding Rounding) FixedBP {
	if a.GTE(b) {
		return ONE
	}
	bp := a.MulRaw(ONE)
	if rounding == UP {
		bp = bp.Add(b.SubRaw(1))
	}
	return FixedBP(bp.Quo(b).Uint64())
}
