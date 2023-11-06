package types

import (
	"fmt"
	"strings"

	"github.com/umee-network/umee/v6/util/sdkutil"
)

func (ap *AccountPosition) String() string {
	special := []string{}
	collateral := []string{}
	borrowed := []string{}
	for _, wsp := range ap.specialPairs {
		special = append(special, wsp.String())
	}
	for _, c := range ap.collateralValue {
		collateral = append(collateral, c.String())
	}
	for _, b := range ap.borrowedValue {
		borrowed = append(borrowed, b.String())
	}
	sep := "\n  "
	return fmt.Sprint(
		"special:", sep,
		strings.Join(special, sep),
		"\ncollateral:", sep,
		strings.Join(collateral, sep),
		"\nborrowed:", sep,
		strings.Join(borrowed, sep),
	)
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
