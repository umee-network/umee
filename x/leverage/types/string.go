package types

import (
	"fmt"
	"strings"

	"github.com/umee-network/umee/v6/util/sdkutil"
)

func (ap *AccountPosition) String() string {
	special := []string{}
	normal := []string{}
	for _, wsp := range ap.specialPairs {
		special = append(special, wsp.String())
	}
	for _, wnp := range ap.normalPairs {
		normal = append(normal, wnp.String())
	}
	for _, c := range ap.unpairedCollateral {
		normal = append(normal, fmt.Sprintf("[%s, -]", c))
	}
	for _, b := range ap.unpairedBorrows {
		normal = append(normal, fmt.Sprintf("[-, %s]", b))
	}
	sep := "\n  "
	return fmt.Sprint(
		"special:", sep,
		strings.Join(special, sep),
		"\nnormal:", sep,
		strings.Join(normal, sep),
	)
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
