package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the oracle module
	ModuleName = "oracle"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the oracle module
	RouterKey = ModuleName

	// QuerierRoute is the query router key for the oracle module
	QuerierRoute = ModuleName
)

// KVStore key prefixes
var (
	KeyPrefixExchangeRate                 = []byte{0x01} // prefix for each key to a rate
	KeyPrefixFeederDelegation             = []byte{0x02} // prefix for each key to a feeder delegation
	KeyPrefixMissCounter                  = []byte{0x03} // prefix for each key to a miss counter
	KeyPrefixAggregateExchangeRatePrevote = []byte{0x04} // prefix for each key to a aggregate prevote
	KeyPrefixAggregateExchangeRateVote    = []byte{0x05} // prefix for each key to a aggregate vote
	KeyPrefixTobinTax                     = []byte{0x06} // prefix for each key to a tobin tax
)

// GetExchangeRateKey - stored by *denom*
func GetExchangeRateKey(denom string) (key []byte) {
	key = append(key, KeyPrefixExchangeRate...)
	key = append(key, []byte(denom)...)
	return append(key, 0) // append 0 for null-termination
}

// GetFeederDelegationKey - stored by *Validator* address
func GetFeederDelegationKey(v sdk.ValAddress) (key []byte) {
	key = append(key, KeyPrefixFeederDelegation...)
	return append(key, address.MustLengthPrefix(v)...)
}

// GetMissCounterKey - stored by *Validator* address
func GetMissCounterKey(v sdk.ValAddress) (key []byte) {
	key = append(key, KeyPrefixMissCounter...)
	return append(key, address.MustLengthPrefix(v)...)
}

// GetAggregateExchangeRatePrevoteKey - stored by *Validator* address
func GetAggregateExchangeRatePrevoteKey(v sdk.ValAddress) (key []byte) {
	key = append(key, KeyPrefixAggregateExchangeRatePrevote...)
	return append(key, address.MustLengthPrefix(v)...)
}

// GetAggregateExchangeRateVoteKey - stored by *Validator* address
func GetAggregateExchangeRateVoteKey(v sdk.ValAddress) (key []byte) {
	key = append(key, KeyPrefixAggregateExchangeRateVote...)
	return append(key, address.MustLengthPrefix(v)...)
}

// GetTobinTaxKey - stored by *denom* bytes
func GetTobinTaxKey(d string) (key []byte) {
	key = append(key, KeyPrefixTobinTax...)
	key = append(key, []byte(d)...)
	return append(key, 0) // append 0 for null-termination
}

// ExtractDenomFromTobinTaxKey - split denom from the tobin tax key
func ExtractDenomFromTobinTaxKey(key []byte) (denom string) {
	denom = string(key[1:])
	return
}
