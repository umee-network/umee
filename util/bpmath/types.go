package bpmath

import (
	"cosmossdk.io/math"
)

type Rounding uint

const (
	DOWN Rounding = iota
	UP
)

const (
	One  = 10000
	Half = One / 2
	Zero = 0
)

var (
	oneBigInt = math.NewIntFromUint64(One)
	oneDec    = math.LegacyNewDec(One)
)
