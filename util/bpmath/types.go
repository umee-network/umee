package bpmath

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

var oneBigInt = math.NewIntFromUint64(One)
var oneDec = sdk.NewDec(One)
