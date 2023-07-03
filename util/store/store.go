// Package store implements KVStore getters and setters for various types.
//
// Numerical getters and setters have a minimum value of zero, which will be interpreted as follows:
//   - Setters clear a KVStore entry when setting to exactly zero.
//   - Getters panic on unmarshal error, and interpret empty data as zero.
//
// All functions which can fail require an errField parameter which is used when creating error messages.
// For example, the errField "balance" could appear in errors like "-12: cannot set negative balance".
// These errFields are cosmetic and do not affect the stored data (key or value) in any way.
package store

import (
	"encoding/binary"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetValue loads value from the store using default Unmarshaler
func GetValue[TPtr PtrMarshalable[T], T any](store sdk.KVStore, key []byte, errField string) TPtr {
	if bz := store.Get(key); len(bz) > 0 {
		var c TPtr = new(T)
		if err := c.Unmarshal(bz); err != nil {
			panic(fmt.Sprintf("error unmarshaling %s into %T: %s", errField, c, err))
		}
		return c
	}

	return nil
}

// SetValue saves value in the store using default Marshaler
func SetValue[T Marshalable](store sdk.KVStore, key []byte, value T, errField string) error {
	bz, err := value.Marshal()
	if err != nil {
		return fmt.Errorf("can't marshal %s: %s", errField, err)
	}
	store.Set(key, bz)
	return nil
}

// GetBinValue is similar to GetValue (loads value in the store),
// but uses UnmarshalBinary interface instead of protobuf
func GetBinValue[TPtr PtrBinMarshalable[T], T any](store sdk.KVStore, key []byte, errField string) (TPtr, error) {
	if bz := store.Get(key); len(bz) > 0 {
		var c TPtr = new(T)
		if err := c.UnmarshalBinary(bz); err != nil {
			return nil, fmt.Errorf("error unmarshaling %s into %T: %s", errField, c, err)
		}
		return c, nil
	}
	return nil, nil
}

// SetBinValue is similar to SetValue (stores value in the store),
// but uses UnmarshalBinary interface instead of protobuf
func SetBinValue[T BinMarshalable](store sdk.KVStore, key []byte, value T, errField string) error {
	bz, err := value.MarshalBinary()
	if err != nil {
		return fmt.Errorf("can't marshal %s: %s", errField, err)
	}
	store.Set(key, bz)
	return nil
}

// GetObject gets and unmarshals a structure from KVstore. Panics on failure to decode, and returns a boolean
// indicating whether any data was found. If the return is false, the object might not be initialized with
// valid zero values for its type.
func GetObject(store sdk.KVStore, cdc codec.Codec, key []byte, object codec.ProtoMarshaler, errField string) bool {
	if bz := store.Get(key); len(bz) > 0 {
		err := cdc.Unmarshal(bz, object)
		if err != nil {
			panic(errField + " could not be unmarshaled: " + err.Error())
		}
		return true
	}
	// No stored bytes at key: return false
	return false
}

// SetObject marshals and sets a structure in KVstore. Returns error on failure to encode.
func SetObject(store sdk.KVStore, cdc codec.Codec, key []byte, object codec.ProtoMarshaler, errField string) error {
	bz, err := cdc.Marshal(object)
	if err != nil {
		return fmt.Errorf("failed to encode %s, %s", errField, err.Error())
	}
	store.Set(key, bz)
	return nil
}

// GetInt retrieves an sdkmath.Int from a KVStore, or returns zero if no value is stored.
// It panics if a stored value fails to unmarshal or is negative.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetInt(store sdk.KVStore, key []byte, errField string) sdkmath.Int {
	val := GetValue[*sdkmath.Int](store, key, errField)
	if val == nil {
		// Not found, return zero
		return sdk.ZeroInt()
	}
	return *val
}

// SetInt stores an sdkmath.Int in a KVStore, or clears if setting to zero or nil.
// Returns an error on serialization error.
// Accepts an additional string which should describe the field being set in custom error messages.
func SetInt(store sdk.KVStore, key []byte, val sdkmath.Int, errField string) error {
	if val.IsNil() || val.IsZero() {
		store.Delete(key)
		return nil
	}
	return SetValue(store, key, &val, errField)
}

// GetDec retrieves an sdk.Dec from a KVStore, or returns zero if no value is stored.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetDec(store sdk.KVStore, key []byte, errField string) sdk.Dec {
	val := GetValue[*sdk.Dec](store, key, errField)
	if val == nil {
		// Not found: return zero
		return sdk.ZeroDec()
	}
	return *val
}

// SetDec stores an sdk.Dec in a KVStore, or clears if setting to zero or nil.
// Returns an error serialization failure.
// Accepts an additional string which should describe the field being set in custom error messages.
func SetDec(store sdk.KVStore, key []byte, val sdk.Dec, errField string) error {
	if val.IsNil() || val.IsZero() {
		store.Delete(key)
		return nil
	}
	return SetValue(store, key, &val, errField)
}

// GetAddress retrieves an sdk.AccAddress from a KVStore, or an empty address if no value is stored.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetAddress(store sdk.KVStore, key []byte) sdk.AccAddress {
	if bz := store.Get(key); len(bz) > 0 {
		addr := sdk.AccAddress(bz)
		return addr
	}
	// No stored bytes at key: return empty address
	return sdk.AccAddress{}
}

// SetAddress stores an sdk.AccAddress in a KVStore, or clears if setting to an empty or nil address.
// Accepts an additional string which should describe the field being set in custom error messages.
func SetAddress(store sdk.KVStore, key []byte, val sdk.AccAddress) {
	if val == nil || val.Empty() {
		store.Delete(key)
		return
	}
	store.Set(key, val)
}

func SetInteger[T Integer](store sdk.KVStore, key []byte, v T) {
	var bz []byte
	switch v := any(v).(type) {
	case int64:
		bz = make([]byte, 8)
		binary.LittleEndian.PutUint64(bz, uint64(v))
	case uint64:
		bz = make([]byte, 8)
		binary.LittleEndian.PutUint64(bz, v)
	case int32:
		bz = make([]byte, 4)
		binary.LittleEndian.PutUint32(bz, uint32(v))
	case uint32:
		bz = make([]byte, 4)
		binary.LittleEndian.PutUint32(bz, v)
	case byte:
		bz = []byte{v}
	}
	store.Set(key, bz)
}

func GetInteger[T Integer](store sdk.KVStore, key []byte) T {
	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	var v T
	switch any(v).(type) {
	case int64, uint64:
		v2 := binary.LittleEndian.Uint64(bz)
		return T(v2)
	case int32, uint32:
		return T(binary.LittleEndian.Uint32(bz))
	case byte:
		return T(bz[0])
	}
	panic("not possible: all types must be covered above")
}
