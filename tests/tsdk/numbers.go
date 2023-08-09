package tsdk

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DecF creates a Dec based on a float number.
// MUST not be used in production code. Can only be used in tests.
func DecF(amount float64) sdk.Dec {
	return sdk.MustNewDecFromStr(strconv.FormatFloat(amount, 'f', -1, 64))
}
