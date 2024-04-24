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
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetValue loads value from the store using default Unmarshaler. Panics on failure to decode.
// Returns nil if the key is not found in the store.
// If the value contains codec.Any field, then SetObject MUST be used instead.
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

// SetValue saves value in the store using default Marshaler. Returns error in case
// of marshaling failure.
// If the value contains codec.Any field, then SetObject MUST be used instead.
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

// GetValueCdc is similar to GetValue, but uses codec for marshaling. For Protobuf objects the
// result is the same, unless codec.Any is used. In the latter case this function MUST be used,
// instead of GetValue.
// Returns true when the data was found and deserialized into the object. Otherwise returns
// false without modifying the object.
func GetValueCdc(store sdk.KVStore, cdc codec.BinaryCodec, key []byte, object codec.ProtoMarshaler,
	errField string) bool {

	if bz := store.Get(key); len(bz) > 0 {
		err := cdc.Unmarshal(bz, object)
		if err != nil {
			panic(errField + " could not be unmarshaled: " + err.Error())
		}
		return true
	}
	return false
}

// SetValueCdc is similar to the SetValue, but uses codec for marshaling. For Protobuf objects the
// result is the same, unless codec.Any is used. In the latter case this function MUST be used,
// instead of SetValue.
func SetValueCdc(store sdk.KVStore, cdc codec.BinaryCodec, key []byte, object codec.ProtoMarshaler,
	errField string) error {

	bz, err := cdc.Marshal(object)
	if err != nil {
		return fmt.Errorf("failed to encode %s, %s", errField, err.Error())
	}
	store.Set(key, bz)
	return nil
}

// GetInt retrieves an sdkmath.Int from a KVStore, or returns (0, false) if no value is stored.
// It panics if a stored value fails to unmarshal or is negative.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetInt(store sdk.KVStore, key []byte, errField string) (sdkmath.Int, bool) {
	val := GetValue[*sdkmath.Int](store, key, errField)
	if val == nil { // Not found
		return sdk.ZeroInt(), false
	}
	return *val, true
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

// GetDec retrieves an sdk.Dec from a KVStore, or returns (0, false) if no value is stored.
// Accepts an additional string which should describe the field being retrieved in custom error messages.
func GetDec(store sdk.KVStore, key []byte, errField string) (sdk.Dec, bool) {
	val := GetValue[*sdk.Dec](store, key, errField)
	if val == nil { // Not found
		return sdk.ZeroDec(), false
	}
	return *val, true
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
		return bz
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

// GetTimeMs retrieves time saved as Unix time in Miliseconds.
// If the value is not in the store, returns (0 unix time, false).
func GetTimeMs(store sdk.KVStore, key []byte) (time.Time, bool) {
	t, ok := GetInteger[int64](store, key)
	return time.UnixMilli(t), ok
}

// SetTimeMs saves time as Unix time in Miliseconds.
func SetTimeMs(store sdk.KVStore, key []byte, t time.Time) {
	SetInteger(store, key, t.UnixMilli())
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

func GetInteger[T Integer](store sdk.KVStore, key []byte) (T, bool) {
	bz := store.Get(key)
	if bz == nil {
		return 0, false
	}
	var v T
	switch any(v).(type) {
	case int64, uint64:
		return T(binary.LittleEndian.Uint64(bz)), true
	case int32, uint32:
		return T(binary.LittleEndian.Uint32(bz)), true
	case byte:
		return T(bz[0]), true
	}
	panic("not possible: all types must be covered above")
}

// DeleteByPrefixStore will delete all keys stored in prefix store
func DeleteByPrefixStore(store sdk.KVStore) {
	iter := sdk.KVStorePrefixIterator(store, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}
