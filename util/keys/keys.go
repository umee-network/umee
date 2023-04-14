package keys

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// LeadingAddress extracts addr from a key of the form of
// prefix | lengthPrefixed(addr) | ...
// prefix can be nil if it must be skipped
// also returns how many bytes were read
func LeadingAddress(prefix, key []byte) (addr sdk.AccAddress, read int, err error) {
	if len(key) <= len(prefix) {
		return sdk.AccAddress{}, 0, fmt.Errorf("key too short to contain leading address after prefix")
	}
	// interpret, as 0-255, the byte after the prefix
	addrLen := int(key[len(prefix)])
	if len(key) < len(prefix)+1+addrLen {
		return sdk.AccAddress{}, 0, fmt.Errorf("key too short to contain specified address length after prefix")
	}
	addr = key[len(prefix)+1 : len(prefix)+1+addrLen]
	return addr, len(prefix) + 1 + addrLen, nil
}

// LeadingDenom extracts denom from a key of the form of
// prefix | bytes(string) | 0x0 | ...
// prefix can be nil if it must be skipped
// also returns how many bytes were read
func LeadingDenom(prefix, key []byte) (denom string, read int, err error) {
	if len(key) <= len(prefix) {
		return "", 0, fmt.Errorf("key too short to contain null terminated denom after prefix")
	}
	// find the first 0x00 after prefix
	for i, b := range key[len(prefix):] {
		if b == 0 {
			denom = string(key[len(prefix) : len(prefix)+i])
			read = len(prefix) + i + 1
			return denom, read, nil
		}
	}
	return "", 0, fmt.Errorf("null terminator not found")
}

// LeadingAddressAndDenom extracts addr and denom from a key of the form of
// prefix | lengthPrefixed(addr) | denom | 0x0 |  ...
func LeadingAddressAndDenom(prefix, key []byte) (addr sdk.AccAddress, denom string, read int, err error) {
	// first parse leading address
	addr, read, err = LeadingAddress(prefix, key)
	if err != nil {
		return addr, "", read, err
	}

	// continue pasring key from where we left off (no more prefix)
	secondRead := 0
	denom, secondRead, err = LeadingDenom(nil, key[read:])

	// bytes read = amount in address+prefix, plus additional bytes for denom
	return addr, denom, read + secondRead, err
}
