package bpmath

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Rounding uint

const (
	DOWN = iota
	UP
)

const ONE = 10000
const half = ONE / 2

var (
	oneBigInt sdk.Int
)

func init() {
	oneBigInt = sdk.NewIntFromUint64(ONE)
}
