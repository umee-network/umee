package sdkutil

import (
	"fmt"
	"math/rand"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FormatDec formats a sdk.Dec as a string with no trailing zeroes after the decimal point,
// omitting the decimal point as well for whole numbers.
// e.g. 4.000 -> 4 and 3.500 -> 3.5
func FormatDec(d sdk.Dec) string {
	dStr := d.String()
	parts := strings.Split(dStr, ".")
	if len(parts) != 2 {
		return dStr
	}
	integer, decimal := parts[0], parts[1]
	decimal = strings.TrimRight(decimal, "0") // no need for trailing zeros after the "."
	if decimal == "" {
		return integer
	}
	return fmt.Sprint(integer, ".", decimal)
}

// FormatDecCoin formats a sdk.DecCoin with no trailing zeroes after the decimal point in its amount,
// omitting the decimal point as well for whole numbers. Also places a space between amount and denom.
// e.g. 4.000uumee -> 4 uumee and 3.500ibc/abcd -> 3.5 ibc/abcd
func FormatDecCoin(c sdk.DecCoin) string {
	return fmt.Sprintf("%s %s", FormatDec(c.Amount), c.Denom)
}

func GenerateString(length uint) string {
	// character set used for generating a random string in GenerateString
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = charset[rand.Intn(len(charset))]
	}
	return string(bytes)
}
