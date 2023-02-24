// Package store implements KVStore getters and setters for various types.
//
// The methods in this package are made to be resistant to bugs introduced through the following pathways:
//   - Manually edited genesis state (bad input will appear in SetX during ImportGenesis)
//   - Bug in code (bad input could appear in SetX at any time)
//   - Incomplete migration (bad data will be unmarshaled by GetX after the migration)
//   - Incorrect binary swap (same vectors as incomplete migration)
//
// Setters return errors on bad input. For getters (which should only see bad data in exotic cases)
// panics are used instead for simpler function signatures.
//
// All functions require an errField parameter which is used when creating error messages. For example,
// the errField "balance" could appear in errors like "-12 is not above the minimum balance of 0".
// These errFields are cosmetic and do not affect the stored data (key or value) in any way.
package store

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"
)

// GetInt retrieves an sdkmath.Int from a KVStore, or a provided minimum value if no value is stored.
// It panics if a stored value fails to unmarshal or is less than or equal to the minimum value.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetInt(store sdk.KVStore, key []byte, min sdkmath.Int, errField string) sdkmath.Int {
	if bz := store.Get(key); len(bz) > 0 {
		val := sdk.ZeroInt()
		if err := val.Unmarshal(bz); err != nil {
			panic(fmt.Sprintf("error unmarshaling %s into %T: %s", errField, val, err))
		}
		if val.IsNil() || val.LTE(min) {
			panic(fmt.Sprintf("%s is not above the minimum %s of %s", val, errField, min))
		}
		return val
	}
	// No stored bytes at key: return minimum value
	return min
}

// SetInt stores an sdkmath.Int in a KVStore, or clears if setting to a provided minimum value or nil.
// Returns an error on attempting to store value lower than the minimum or on failure to encode.
// Accepts an additional string which should describe the field being set in custom error messages.
func SetInt(store sdk.KVStore, key []byte, val, min sdkmath.Int, errField string) error {
	if val.IsNil() || val.Equal(min) {
		store.Delete(key)
		return nil
	}
	if val.LT(min) {
		return fmt.Errorf("%s is below the minimum %s of %s", val, errField, min)
	}
	bz, err := val.Marshal()
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetDec retrieves an sdk.Dec from a KVStore, or a provided minimum value if no value is stored.
// It panics if a stored value fails to unmarshal or is less than or equal to the minimum value.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetDec(store sdk.KVStore, key []byte, min sdk.Dec, errField string) sdk.Dec {
	if bz := store.Get(key); len(bz) > 0 {
		val := sdk.ZeroDec()
		if err := val.Unmarshal(bz); err != nil {
			panic(fmt.Sprintf("error unmarshaling %s into %T: %s", errField, val, err))
		}
		if val.IsNil() || val.LTE(min) {
			panic(fmt.Sprintf("%s is not above the minimum %s of %s", val, errField, min))
		}
		return val
	}
	// No stored bytes at key: return minimum value
	return min
}

// SetDec stores an sdk.Dec in a KVStore, or clears if setting to a provided minimum value or nil.
// Returns an error on attempting to store value lower than the minimum or on failure to encode.
// Accepts an additional string which should describe the field being set in custom error messages.
func SetDec(store sdk.KVStore, key []byte, val, min sdk.Dec, errField string) error {
	if val.IsNil() || val.Equal(min) {
		store.Delete(key)
		return nil
	}
	if val.LT(min) {
		return fmt.Errorf("%s is below the minimum %s of %s", val, errField, min)
	}
	bz, err := val.Marshal()
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetUint32 retrieves a uint32 from a KVStore, or a provided minimum value if no value is stored.
// Uses gogoproto Uint32Value for unmarshaling.
// It panics if a stored value fails to unmarshal or is less than or equal to the minimum value.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetUint32(store sdk.KVStore, key []byte, min uint32, errField string) uint32 {
	if bz := store.Get(key); len(bz) > 0 {
		val := gogotypes.UInt32Value{}
		if err := val.Unmarshal(bz); err != nil {
			panic(fmt.Sprintf("error unmarshaling %s into %T: %s", errField, val, err))
		}
		if val.Value <= min {
			panic(fmt.Sprintf("%d is not above the minimum %s of %d", val.Value, errField, min))
		}
		return val.Value
	}
	// No stored bytes at key: return minimum value
	return min
}

// SetUint32 stores a uint32 in a KVStore, or clears if setting to a provided minimum value.
// Uses gogoproto Uint32Value for marshaling.
// Returns an error on attempting to store value lower than the minimum or on failure to encode.
// Accepts an additional string which should describe the field being set in custom error messages.
func SetUint32(store sdk.KVStore, key []byte, val, min uint32, errField string) error {
	if val < min {
		return fmt.Errorf("%d is below the minimum %s of %d", val, errField, min)
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

// GetUint64 retrieves a uint32 from a KVStore, or a provided minimum value if no value is stored.
// Uses gogoproto Uint64Value for unmarshaling.
// It panics if a stored value fails to unmarshal or is less than or equal to the minimum value.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetUint64(store sdk.KVStore, key []byte, min uint64, errField string) uint64 {
	if bz := store.Get(key); len(bz) > 0 {
		val := gogotypes.UInt64Value{}
		if err := val.Unmarshal(bz); err != nil {
			panic(fmt.Sprintf("error unmarshaling %s into %T: %s", errField, val, err))
		}
		if val.Value <= min {
			panic(fmt.Sprintf("%d is not above the minimum %s of %d", val.Value, errField, min))
		}
		return val.Value
	}
	// No stored bytes at key: return minimum value
	return min
}

// SetUint64 stores a uint32 in a KVStore, or clears if setting to a provided minimum value.
// Uses gogoproto Uint64Value for marshaling.
// Returns an error on attempting to store value lower than the minimum or on failure to encode.
// Accepts an additional string which should describe the field being set in custom error messages.
func SetUint64(store sdk.KVStore, key []byte, val, min uint64, errField string) error {
	if val < min {
		return fmt.Errorf("%d is below the minimum %s of %d", val, errField, min)
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

// GetAddress retrieves an sdk.AccAddress from a KVStore, or an empty address if no value is stored.
// Panics if a non-empty address fails sdk.VerifyAddressFormat, so non-empty returns are always valid.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetAddress(store sdk.KVStore, key []byte, errField string) sdk.AccAddress {
	if bz := store.Get(key); len(bz) > 0 {
		addr := sdk.AccAddress(bz)
		if err := sdk.VerifyAddressFormat(addr); err != nil {
			panic(fmt.Sprintf("%s is not valid: %s", errField, err))
		}
		// Returns valid address
		return sdk.AccAddress(bz)
	}
	// No stored bytes at key: return empty address
	return sdk.AccAddress{}
}

// SetAddress stores an sdk.AccAddress in a KVStore, or clears if setting to an empty or nil address.
// Returns an error on attempting to store a non-empty address that fails sdk.VerifyAddressFormat.
// Accepts an additional string which should describe the field being set in custom error messages.
func SetAddress(store sdk.KVStore, key []byte, val sdk.AccAddress, errField string) error {
	if val == nil || val.Empty() {
		store.Delete(key)
		return nil
	}
	if err := sdk.VerifyAddressFormat(val); err != nil {
		return fmt.Errorf("%s is not valid: %s", errField, err)
	}
	store.Set(key, val)
	return nil
}
