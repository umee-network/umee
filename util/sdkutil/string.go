package sdkutil

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FormatDec formats a sdk.Dec as a string with no trailing zeroes after the decimal point,
// omitting the decimal point as well for whole numbers.
func FormatDec(d sdk.Dec) string {
	dStr := d.String()
	parts := strings.Split(dStr, ".")
	if len(parts) != 2 {
		return dStr
	}
	integer, decimal := parts[0], parts[1]
	decimal = strings.TrimRight(decimal, "0")  // no need for trailing zeros after the "."
	if decimal == "" {
		return integer
	}
	return fmt.Sprint(integer, ".", decimal)
}

// FormatDecCoin formats a sdk.DecCoin with no trailing zeroes after the decimal point in its amount,
// omitting the decimal point as well for whole numbers.
func FormatDecCoin(c sdk.DecCoin) string {
	return fmt.Sprintf("%s %s", FormatDec(c.Amount), c.Denom)
}
