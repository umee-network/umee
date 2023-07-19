package bpmath

import (
	"cosmossdk.io/math"
)

type Rounding uint

const (
	DOWN = iota
	UP
)

const ONE = 10000
const half = ONE / 2

var (
	oneBigInt = math.NewIntFromUint64(ONE)
)
