package store

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Int converts bytes to sdk.Int, and panics on failure or negative value
// with a message which includes the name of the value being retrieved
func Int(bz []byte, errField string) sdkmath.Int {
	val := sdk.ZeroInt()
	if err := val.Unmarshal(bz); err != nil {
		panic(fmt.Sprintf("error unmarshaling %s into %T: %s", errField, val, err))
	}
	if val.IsNegative() {
		panic(fmt.Sprintf("%s: retrieved negative %s", val, errField))
	}
	return val
}
