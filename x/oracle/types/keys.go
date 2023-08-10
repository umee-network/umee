package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/umee-network/umee/v6/util"
)

const (
	// ModuleName is the name of the oracle module
	ModuleName = "oracle"

	// StoreKey is the string store representation
	StoreKey = ModuleName
)

// KVStore key prefixes
var (
	KeyPrefixExchangeRate                 = []byte{0x01} // prefix for each key to a rate
	KeyPrefixFeederDelegation             = []byte{0x02} // prefix for each key to a feeder delegation
	KeyPrefixMissCounter                  = []byte{0x03} // prefix for each key to a miss counter
	KeyPrefixAggregateExchangeRatePrevote = []byte{0x04} // prefix for each key to a aggregate prevote
	KeyPrefixAggregateExchangeRateVote    = []byte{0x05} // prefix for each key to a aggregate vote
	KeyPrefixMedian                       = []byte{0x06} // prefix for each key to a price median
	KeyPrefixMedianDeviation              = []byte{0x07} // prefix for each key to a price median standard deviation
	KeyPrefixHistoricPrice                = []byte{0x08} // prefix for each key to a historic price
	KeyPrefixAvgCounter                   = []byte{0x09} // prefix for each key to a historic avg price counter
	KeyLatestAvgCounter                   = []byte{0x10} // key where we store the latest avg price counter
)

// KeyExchangeRate - stored by *denom*
func KeyExchangeRate(denom string) []byte {
	// append 0 for null-termination
	return util.ConcatBytes(1, KeyPrefixExchangeRate, []byte(denom))
}

// KeyFeederDelegation - stored by *Validator* address
func KeyFeederDelegation(v sdk.ValAddress) []byte {
	return util.ConcatBytes(0, KeyPrefixFeederDelegation, address.MustLengthPrefix(v))
}

// KeyMissCounter - stored by *Validator* address
func KeyMissCounter(v sdk.ValAddress) []byte {
	return util.ConcatBytes(0, KeyPrefixMissCounter, address.MustLengthPrefix(v))
}

// KeyAggregateExchangeRatePrevote - stored by *Validator* address
func KeyAggregateExchangeRatePrevote(v sdk.ValAddress) []byte {
	return util.ConcatBytes(0, KeyPrefixAggregateExchangeRatePrevote, address.MustLengthPrefix(v))
}

// KeyAggregateExchangeRateVote - stored by *Validator* address
func KeyAggregateExchangeRateVote(v sdk.ValAddress) []byte {
	return util.ConcatBytes(0, KeyPrefixAggregateExchangeRateVote, address.MustLengthPrefix(v))
}

// KeyMedian - stored by *denom*
func KeyMedian(denom string, blockNum uint64) (key []byte) {
	return util.ConcatBytes(0, KeyPrefixMedian, []byte(denom), util.UintWithNullPrefix(blockNum))
}

// KeyMedianDeviation - stored by *denom*
func KeyMedianDeviation(denom string, blockNum uint64) (key []byte) {
	return util.ConcatBytes(0, KeyPrefixMedianDeviation, []byte(denom), util.UintWithNullPrefix(blockNum))
}

// KeyHistoricPrice - stored by *denom* and *block*
func KeyHistoricPrice(denom string, blockNum uint64) (key []byte) {
	return util.ConcatBytes(0, KeyPrefixHistoricPrice, []byte(denom), util.UintWithNullPrefix(blockNum))
}

// KeyHistoricPrice - stored by *denom* and *block*
func KeyAvgCounter(denom string, counterID byte) (key []byte) {
	return util.ConcatBytes(0, KeyPrefixAvgCounter, []byte(denom), []byte{counterID})
}

// ParseDenomAndBlockFromKey returns the denom and block contained in the *key*
// that has a uint64 at the end with a null prefix (length 9).
func ParseDenomAndBlockFromKey(key []byte, prefix []byte) (string, uint64) {
	return string(key[len(prefix) : len(key)-9]), binary.LittleEndian.Uint64(key[len(key)-8:])
}
