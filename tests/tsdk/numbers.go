package tsdk

import (
	"strconv"

	sdkmath "cosmossdk.io/math"
)

// DecF creates a Dec based on a float number.
// MUST not be used in production code. Can only be used in tests.
func DecF(amount float64) sdkmath.LegacyDec {
	return sdkmath.LegacyMustNewDecFromStr(strconv.FormatFloat(amount, 'f', -1, 64))
}
