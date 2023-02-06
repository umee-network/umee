package store

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"
)

// GetInt retrieves an sdkmath.Int from a KVStore, or a provided minimum value if no value is stored.
// It panics if a stored value fails to unmarshal or is less than or equal to the minimum value.
func GetInt(store sdk.KVStore, key []byte, min sdkmath.Int) sdkmath.Int {
	if bz := store.Get(key); bz != nil {
		val := sdk.ZeroInt()
		if err := val.Unmarshal(bz); err != nil {
			panic(fmt.Sprintf("error unmarshaling into %T: %s", val, err))
		}
		if val.IsNil() || val.LTE(min) {
			panic(fmt.Sprintf("%s is not above the minimum of %s", val, min))
		}
		return val
	}
	// No stored bytes at key: return minimum value
	return min
}

// SetInt stores an sdkmath.Int in a KVStore, or clears if setting to a provided minimum value or nil.
// Returns an error on attempting to store value lower than the minimum or on failure to encode.
func SetInt(store sdk.KVStore, key []byte, val, min sdkmath.Int) error {
	if val.IsNil() || val.Equal(min) {
		store.Delete(key)
		return nil
	}
	if val.LT(min) {
		return fmt.Errorf("%s is below the minimum of %s", val, min)
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
func GetDec(store sdk.KVStore, key []byte, min sdk.Dec) sdk.Dec {
	if bz := store.Get(key); bz != nil {
		val := sdk.ZeroDec()
		if err := val.Unmarshal(bz); err != nil {
			panic(fmt.Sprintf("error unmarshaling into %T: %s", val, err))
		}
		if val.IsNil() || val.LTE(min) {
			panic(fmt.Sprintf("%s is not above the minimum of %s", val, min))
		}
		return val
	}
	// No stored bytes at key: return minimum value
	return min
}

// SetDec stores an sdk.Dec in a KVStore, or clears if setting to a provided minimum value or nil.
// Returns an error on attempting to store value lower than the minimum or on failure to encode.
func SetDec(store sdk.KVStore, key []byte, val, min sdk.Dec) error {
	if val.IsNil() || val.Equal(min) {
		store.Delete(key)
		return nil
	}
	if val.LT(min) {
		return fmt.Errorf("%s is below the minimum of %s", val, min)
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
func GetUint32(store sdk.KVStore, key []byte, min uint32) uint32 {
	if bz := store.Get(key); bz != nil {
		val := gogotypes.UInt32Value{}
		if err := val.Unmarshal(bz); err != nil {
			panic(fmt.Sprintf("error unmarshaling into %T: %s", val, err))
		}
		if val.Value <= min {
			panic(fmt.Sprintf("%d is not above the minimum of %d", val.Value, min))
		}
		return val.Value
	}
	// No stored bytes at key: return minimum value
	return min
}

// SetUint32 stores a uint32 in a KVStore, or clears if setting to a provided minimum value.
// Uses gogoproto Uint32Value for marshaling.
// Returns an error on attempting to store value lower than the minimum or on failure to encode.
func SetUint32(store sdk.KVStore, key []byte, val, min uint32) error {
	if val < min {
		return fmt.Errorf("%d is below the minimum of %d", val, min)
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
func GetUint64(store sdk.KVStore, key []byte, min uint64) uint64 {
	if bz := store.Get(key); bz != nil {
		val := gogotypes.UInt64Value{}
		if err := val.Unmarshal(bz); err != nil {
			panic(fmt.Sprintf("error unmarshaling into %T: %s", val, err))
		}
		if val.Value <= min {
			panic(fmt.Sprintf("%d is not above the minimum of %d", val.Value, min))
		}
		return val.Value
	}
	// No stored bytes at key: return minimum value
	return min
}

// SetUint64 stores a uint32 in a KVStore, or clears if setting to a provided minimum value.
// Uses gogoproto Uint64Value for marshaling.
// Returns an error on attempting to store value lower than the minimum or on failure to encode.
func SetUint64(store sdk.KVStore, key []byte, val, min uint64) error {
	if val < min {
		return fmt.Errorf("%d is below the minimum of %d", val, min)
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
func GetAddress(store sdk.KVStore, key []byte) sdk.AccAddress {
	if bz := store.Get(key); bz != nil {
		addr := sdk.AccAddress(bz)
		if err := sdk.VerifyAddressFormat(addr); err != nil {
			panic(err)
		}
		// Returns valid address
		return sdk.AccAddress(bz)
	}
	// No stored bytes at key: return empty address
	return sdk.AccAddress{}
}

// SetAddress stores an sdk.AccAddress in a KVStore, or clears if setting to an empty or nil address.
// Returns an error on attempting to store a non-empty address that fails sdk.VerifyAddressFormat.
func SetAddress(store sdk.KVStore, key []byte, val sdk.AccAddress) error {
	if val == nil || val.Empty() {
		store.Delete(key)
		return nil
	}
	if err := sdk.VerifyAddressFormat(val); err != nil {
		return err
	}
	store.Set(key, val)
	return nil
}
