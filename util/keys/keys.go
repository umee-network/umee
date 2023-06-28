package keys

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExtractAddress extracts addr from a key of the form of
// ... | lengthPrefixed(addr) | ...
// starting at a given index. Also returns the index of the first byte following the address.
func ExtractAddress(startIndex int, key []byte) (addr sdk.AccAddress, nextIndex int, err error) {
	if len(key) <= startIndex {
		return sdk.AccAddress{}, 0, fmt.Errorf("key too short to contain address after startIndex")
	}
	// interpret, as 0-255, the byte after the prefix
	addrLen := int(key[startIndex])
	if len(key) < startIndex+1+addrLen {
		return sdk.AccAddress{}, 0, fmt.Errorf("key too short to contain specified address length after prefix")
	}
	addr = key[startIndex+1 : startIndex+1+addrLen]
	return addr, startIndex + 1 + addrLen, nil
}

// ExtractString extracts a null-terminated string from a key of the form of
// ... | bytes(string) | 0x0 | ...
// starting at a given index. Also returns the index of the first byte following the null terminator.
func ExtractString(startIndex int, key []byte) (s string, nextIndex int, err error) {
	if len(key) <= startIndex+1 {
		return "", 0, fmt.Errorf("key too short to contain non-empty null terminated string after startIndex")
	}
	if key[startIndex] == 0 {
		return "", 0, fmt.Errorf("empty null-terminated string not allowed in key")
	}
	// find the first 0x00 after prefix
	for i, b := range key[startIndex:] {
		if b == 0 {
			s = string(key[startIndex : startIndex+i])
			nextIndex = startIndex + i + 1
			return s, nextIndex, nil
		}
	}
	return "", 0, fmt.Errorf("null terminator not found")
}

// ExtractAddressAndString extracts addr and a null-terminated string from a key of the form of
// ... | lengthPrefixed(addr) | bytes(string) | 0x0 |  ...
// starting at a given index. Also returns the index of the first byte following the null terminator.
func ExtractAddressAndString(startIndex int, key []byte) (addr sdk.AccAddress, s string, nextIndex int, err error) {
	// first parse leading address
	addr, nextIndex, err = ExtractAddress(startIndex, key)
	if err != nil {
		return addr, "", 0, err
	}

	// continue parsing key from where we left off
	s, nextIndex, err = ExtractString(nextIndex, key)
	return addr, s, nextIndex, err
}

// ToStr takes the full key and converts it to a string
func ToStr(key []byte) string {
	return string(key)
}

// NoLastByte returns sub-slice of the key without the last byte.
// Panics if length of key is zero.
func NoLastByte(key []byte) string {
	return string(key[:len(key)-1])
}
