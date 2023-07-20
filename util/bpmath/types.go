package bpmath

import (
	"cosmossdk.io/math"
)

type Rounding uint

const (
	DOWN Rounding = iota
	UP
)

const One = 10000
const Half = One / 2

var (
	oneBigInt = math.NewIntFromUint64(One)
)
