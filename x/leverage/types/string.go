package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (ap *AccountPosition) String() string {
	s := "special:\n"
	for _, wsp := range ap.specialPairs {
		s += fmt.Sprintf("  %s\n", wsp)
	}
	s += "normal:\n"
	for _, wnp := range ap.normalPairs {
		s += fmt.Sprintf("  %s\n", wnp)
	}
	for _, sc := range ap.unpairedCollateral {
		s += fmt.Sprintf("  [%s, -]\n", sc)
	}
	for _, sb := range ap.unpairedBorrows {
		s += fmt.Sprintf("  [-, %s]\n", sb)
	}
	return s
}

// String represents a WeightedNormalPair in the form [WeightedDecCoin, WeightedDecCoin]
// e.g. [10 uumee @ 0.35, 3.5 uumee @ 0.35]
func (wnp WeightedNormalPair) String() string {
	return fmt.Sprintf("[%s, %s]", wnp.Collateral, wnp.Borrow)
}

// String represents a WeightedSpecialPair in the form [DecCoin, DecCoin] @ weight
// e.g. [10 uumee, 3.5 uumee] @ 0.35
func (wsp WeightedSpecialPair) String() string {
	return "[" + sDecCoin(wsp.Collateral) + ", " + sDecCoin(wsp.Borrow) + "] @ " + sDec(wsp.SpecialWeight)
}

// String represents a WeightedDecCoin in the form coin @ weight
// e.g. 10 uumee @ 0.35
func (wdc WeightedDecCoin) String() string {
	return sDecCoin(wdc.Asset) + " @ " + sDec(wdc.Weight)
}

// sDec formats a sdk.Dec as a string with no trailing zeroes after the decimal point,
// omitting the decimal point as well for whole numbers
func sDec(d sdk.Dec) string {
	// split string before and after decimal
	parts := strings.Split(d.String(), ".")
	if len(parts) != 2 {
		return d.String()
	}
	// remove all trailing zeroes after decimal
	parts[1] = strings.TrimRight(parts[1], "0")
	// if input was a whole number, return without any decimal
	if parts[1] == "" {
		return parts[0]
	}
	// otherwise, return with decimal intact but trailing zeroes removed
	return parts[0] + "." + parts[1]
}

// sDecCoin formats a sdk.DecCoin with no trailing zeroes after the decimal point in its amount,
// omitting the decimal point as well for whole numbers
func sDecCoin(c sdk.DecCoin) string {
	return sDec(c.Amount) + " " + c.Denom
}
