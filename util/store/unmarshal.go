package store

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

// Int converts bytes to sdkmath.Int, and panics on failure or negative value
// with a message which includes the name of the value being retrieved
func Int(bz []byte, errField string) sdkmath.Int {
	val := sdkmath.ZeroInt()
	if err := val.Unmarshal(bz); err != nil {
		panic(fmt.Sprintf("error unmarshaling %s into %T: %s", errField, val, err))
	}
	if val.IsNegative() {
		panic(fmt.Sprintf("%s: retrieved negative %s", val, errField))
	}
	return val
}
