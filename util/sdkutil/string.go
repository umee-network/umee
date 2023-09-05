package sdkutil

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FormatDec formats a sdk.Dec as a string with no trailing zeroes after the decimal point,
// omitting the decimal point as well for whole numbers.
func FormatDec(d sdk.Dec) string {
	// split string before and after decimal
	dStr := d.String()
	parts := strings.Split(dStr, ".")
	if len(parts) != 2 {
		return dStr
	}
	// remove all trailing zeroes after decimal
	parts[1] = strings.TrimRight(parts[1], "0")
	// if input was a whole number, return without any decimal
	if parts[1] == "" {
		return parts[0]
	}
	// otherwise, return with decimal intact but trailing zeroes removed
	return fmt.Sprintf("%s.%s", parts[0], parts[1])
}

// FormatDecCoin formats a sdk.DecCoin with no trailing zeroes after the decimal point in its amount,
// omitting the decimal point as well for whole numbers.
func FormatDecCoin(c sdk.DecCoin) string {
	return fmt.Sprintf("%s %s", FormatDec(c.Amount), c.Denom)
}
