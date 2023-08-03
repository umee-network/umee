package coin

import (
	"strings"
)

const (
	// UTokenPrefix defines the uToken denomination prefix for all uToken types.
	UTokenPrefix = "u/"
)

// HasUTokenPrefix detects the uToken prefix on a denom.
func HasUTokenPrefix(denom string) bool {
	return strings.HasPrefix(denom, UTokenPrefix)
}

// ToUTokenDenom adds the uToken prefix to a denom. Returns an empty string
// instead if the prefix was already present.
func ToUTokenDenom(denom string) string {
	if HasUTokenPrefix(denom) {
		return ""
	}
	return UTokenPrefix + denom
}

// StripUTokenDenom strips the uToken prefix from a denom, or returns an empty
// string if it was not present. Also returns an empty string if the prefix
// was repeated multiple times.
func StripUTokenDenom(denom string) string {
	if !HasUTokenPrefix(denom) {
		return ""
	}
	s := strings.TrimPrefix(denom, UTokenPrefix)
	if HasUTokenPrefix(s) {
		// denom started with "u/u/"
		return ""
	}
	return s
}
