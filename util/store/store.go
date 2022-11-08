package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"
)

// GetStoredInt retrieves an sdkmath.Int from a KVStore, or a provided minimum value if no value is stored.
// It panics if a stored value fails to unmarshal or is less than or equal to the minumum value.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetStoredInt(store sdk.KVStore, key []byte, min sdkmath.Int, desc string) sdkmath.Int {
	if bz := store.Get(key); bz != nil {
		val := sdk.ZeroInt()
		if err := val.Unmarshal(bz); err != nil {
			panic(fmt.Errorf("error unmarshaling %s into %T: %s", desc, val, err))
		}
		if val.LTE(min) {
			panic(fmt.Sprintf("%s is not above the minimum %s of %s", val, desc, min))
		}
		return val
	}
	// No stored bytes at key: return minimum value
	return min
}

// SetStoredInt stores an sdkmath.Int in a KVStore, or clears if setting to a provided minimum value.
// Returns an error on attempting to store value lower than the minimum or on failure to encode.
// Accepts an additional string which should describe the field being set in custom error messages.
func SetStoredInt(store sdk.KVStore, key []byte, val, min sdkmath.Int, desc string) error {
	if val.LT(min) {
		return fmt.Errorf("%s is below the minimum %s of %s", val, desc, min)
	}
	if val.Equal(min) {
		store.Delete(key)
	} else {
		bz, err := val.Marshal()
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}
	return nil
}

// GetStoredDec retrieves an sdk.Dec from a KVStore, or a provided minimum value if no value is stored.
// It panics if a stored value fails to unmarshal or is less than or equal to the minumum value.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetStoredDec(store sdk.KVStore, key []byte, min sdk.Dec, desc string) sdk.Dec {
	if bz := store.Get(key); bz != nil {
		val := sdk.ZeroDec()
		if err := val.Unmarshal(bz); err != nil {
			panic(fmt.Errorf("error unmarshaling %s into %T: %s", desc, val, err))
		}
		if val.LTE(min) {
			panic(fmt.Sprintf("%s is not above the minimum %s of %s", val, desc, min))
		}
		return val
	}
	// No stored bytes at key: return minimum value
	return min
}

// SetStoredDec stores an sdk.Dec in a KVStore, or clears if setting to a provided minimum value.
// Returns an error on attempting to store value lower than the minimum or on failure to encode.
// Accepts an additional string which should describe the field being set in custom error messages.
func SetStoredDec(store sdk.KVStore, key []byte, val, min sdk.Dec, desc string) error {
	if val.LT(min) {
		return fmt.Errorf("%s is below the minimum %s of %s", val, desc, min)
	}
	if val.Equal(min) {
		store.Delete(key)
	} else {
		bz, err := val.Marshal()
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}
	return nil
}

// GetStoredUint32 retrieves a uint32 from a KVStore, or a provided minimum value if no value is stored.
// Uses gogoproto Uint32Value for unmarshaling.
// It panics if a stored value fails to unmarshal or is less than or equal to the minumum value.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetStoredUint32(store sdk.KVStore, key []byte, min uint32, desc string) uint32 {
	if bz := store.Get(key); bz != nil {
		val := gogotypes.UInt32Value{}
		if err := val.Unmarshal(bz); err != nil {
			panic(fmt.Errorf("error unmarshaling %s into %T: %s", desc, val, err))
		}
		if val.Value <= min {
			panic(fmt.Sprintf("%d is not above the minimum %s of %d", val.Value, desc, min))
		}
		return val.Value
	}
	// No stored bytes at key: return minimum value
	return min
}

// SetStoredUint32 stores a uint32 in a KVStore, or clears if setting to a provided minimum value.
// Uses gogoproto Uint32Value for marshaling.
// Returns an error on attempting to store value lower than the minimum or on failure to encode.
// Accepts an additional string which should describe the field being set in custom error messages.
func SetStoredUint32(store sdk.KVStore, key []byte, val, min uint32, desc string) error {
	if val < min {
		return fmt.Errorf("%d is below the minimum %s of %d", val, desc, min)
	}
	if val == min {
		store.Delete(key)
	} else {
		v := &gogotypes.UInt32Value{Value: val}
		bz, err := v.Marshal()
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}
	return nil
}

// GetStoredUint64 retrieves a uint32 from a KVStore, or a provided minimum value if no value is stored.
// Uses gogoproto Uint64Value for unmarshaling.
// It panics if a stored value fails to unmarshal or is less than or equal to the minumum value.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetStoredUint64(store sdk.KVStore, key []byte, min uint64, desc string) uint64 {
	if bz := store.Get(key); bz != nil {
		val := gogotypes.UInt64Value{}
		if err := val.Unmarshal(bz); err != nil {
			panic(fmt.Errorf("error unmarshaling %s into %T: %s", desc, val, err))
		}
		if val.Value <= min {
			panic(fmt.Sprintf("%d is not above the minimum %s of %d", val.Value, desc, min))
		}
		return val.Value
	}
	// No stored bytes at key: return minimum value
	return min
}

// SetStoredUint64 stores a uint32 in a KVStore, or clears if setting to a provided minimum value.
// Uses gogoproto Uint64Value for marshaling.
// Returns an error on attempting to store value lower than the minimum or on failure to encode.
// Accepts an additional string which should describe the field being set in custom error messages.
func SetStoredUint64(store sdk.KVStore, key []byte, val, min uint64, desc string) error {
	if val < min {
		return fmt.Errorf("%d is below the minimum %s of %d", val, desc, min)
	}
	if val == min {
		store.Delete(key)
	} else {
		v := &gogotypes.UInt64Value{Value: val}
		bz, err := v.Marshal()
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}
	return nil
}
