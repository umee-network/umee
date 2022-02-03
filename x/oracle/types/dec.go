package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// constants & variables taken from the cosmos sdk's types package
const (
	decimalPrecisionBits = 60
	maxBitLen            = 256
	maxDecBitLen         = maxBitLen + decimalPrecisionBits
	precision            = 18
)

var (
	precisionReuse = new(big.Int).Exp(big.NewInt(10), big.NewInt(precision), nil)
	oneInt         = big.NewInt(1)
	fivePrecision  = new(big.Int).Quo(precisionReuse, big.NewInt(2))
)

// Remove a Precision amount of rightmost digits and perform bankers rounding
// on the remainder (gaussian rounding) on the digits which have been removed.
// Mutates the input.
// Fork of the method in the SDK.
// Ref: https://github.com/cosmos/cosmos-sdk/blob/e0543a3be0f6bffa2d49fb2911328a6d4a0072bd/types/decimal.go#L508
func chopPrecisionAndRound(d *big.Int) *big.Int {
	// remove the negative and add it back when returning
	if d.Sign() == -1 {
		// make d positive, compute chopped value, and then un-mutate d
		d = d.Neg(d)
		d = chopPrecisionAndRound(d)
		d = d.Neg(d)
		return d
	}

	// get the truncated quotient and remainder
	quo, rem := d, big.NewInt(0)
	quo, rem = quo.QuoRem(d, precisionReuse, rem)

	if rem.Sign() == 0 { // remainder is zero
		return quo
	}

	switch rem.Cmp(fivePrecision) {
	case -1:
		return quo
	case 1:
		return quo.Add(quo, oneInt)
	default: // bankers rounding must take place
		// always round to an even number
		if quo.Bit(0) == 0 {
			return quo
		}
		return quo.Add(quo, oneInt)
	}
}

// checkMulOverflow checks to see if two sdk.Decs will overflow
// when multiplied. Uses the same logic as sdk.Dec.Mul, but
// errors instead of panicing.
// Ref: https://github.com/cosmos/cosmos-sdk/blob/e0543a3be0f6bffa2d49fb2911328a6d4a0072bd/types/decimal.go#L249
func checkMulOverflow(d sdk.Dec, d2 sdk.Dec) error {
	mul := new(big.Int).Mul(d.BigInt(), d2.BigInt())
	chopped := chopPrecisionAndRound(mul)
	if chopped.BitLen() > maxDecBitLen {
		return ErrDecOverflow
	}
	return nil
}
