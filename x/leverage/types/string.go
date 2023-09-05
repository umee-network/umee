package types

import (
	"fmt"

	"github.com/umee-network/umee/v6/util/sdkutil"
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
// e.g. [10 uumee (0.35), 3.5 uumee (0.35)]
func (wnp WeightedNormalPair) String() string {
	return fmt.Sprintf("[%s, %s]", wnp.Collateral, wnp.Borrow)
}

// String represents a WeightedSpecialPair in the form [weight, DecCoin, DecCoin]
// e.g. {0.35, 10 uumee, 3.5 uumee}
func (wsp WeightedSpecialPair) String() string {
	return fmt.Sprintf(
		"{%s, %s, %s}",
		sdkutil.FormatDec(wsp.SpecialWeight),
		sdkutil.FormatDecCoin(wsp.Collateral),
		sdkutil.FormatDecCoin(wsp.Borrow),
	)
}

// String represents a WeightedDecCoin in the form coin (weight)
// e.g. 10 uumee (0.35)
func (wdc WeightedDecCoin) String() string {
	return fmt.Sprintf(
		"%s (%s)",
		sdkutil.FormatDecCoin(wdc.Asset),
		sdkutil.FormatDec(wdc.Weight),
	)
}
