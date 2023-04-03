package store

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"
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

// Dec converts bytes to sdk.Dec, and panics on failure or negative value
// with a message which includes the name of the value being retrieved
func Dec(bz []byte, errField string) sdk.Dec {
	val := sdk.ZeroDec()
	if err := val.Unmarshal(bz); err != nil {
		panic(fmt.Sprintf("error unmarshaling %s into %T: %s", errField, val, err))
	}
	if val.IsNegative() {
		panic(fmt.Sprintf("%s: retrieved negative %s", val, errField))
	}
	return val
}

// Uint32 converts bytes to a uint64 using gogotypes.Uint32Value, and panics on failure
// with a message which includes the name of the value being retrieved
func Uint32(bz []byte, errField string) uint32 {
	val := gogotypes.UInt32Value{}
	if err := val.Unmarshal(bz); err != nil {
		panic(fmt.Sprintf("error unmarshaling %s into %T: %s", errField, val, err))
	}
	return val.Value
}

// Int64 converts bytes to an int64 using gogotypes.Int64Value, and panics on failure
// with a message which includes the name of the value being retrieved
func Int64(bz []byte, errField string) int64 {
	val := gogotypes.Int64Value{}
	if err := val.Unmarshal(bz); err != nil {
		panic(fmt.Sprintf("error unmarshaling %s into %T: %s", errField, val, err))
	}
	return val.Value
}

// Uint64 converts bytes to a uint64 using gogotypes.Uint64Value, and panics on failure
// with a message which includes the name of the value being retrieved
func Uint64(bz []byte, errField string) uint64 {
	val := gogotypes.UInt64Value{}
	if err := val.Unmarshal(bz); err != nil {
		panic(fmt.Sprintf("error unmarshaling %s into %T: %s", errField, val, err))
	}
	return val.Value
}
