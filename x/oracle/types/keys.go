package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the oracle module
	ModuleName = "oracle"

	// StoreKey is the string store representation
	StoreKey = ModuleName

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
	KeyPrefixMedian                       = []byte{0x06} // prefix for each key to a price median
	KeyPrefixMedianDeviation              = []byte{0x07} // prefix for each key to a price median standard deviation
	KeyPrefixHistoricPrice                = []byte{0x08} // prefix for each key to a historic price
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

// GetMedianKey - stored by *denom* and *block*
func GetMedianKey(denom string, blockNum uint64) (key []byte) {
	key = append(key, KeyPrefixMedian...)
	key = appendDenomAndBlockToKey(key, denom, blockNum)
	return append(key, 0) // append 0 for null-termination
}

// GetMedianDeviationKey - stored by *denom* and *block*
func GetMedianDeviationKey(denom string, blockNum uint64) (key []byte) {
	key = append(key, KeyPrefixMedianDeviation...)
	key = appendDenomAndBlockToKey(key, denom, blockNum)
	return append(key, 0) // append 0 for null-termination
}

// GetHistoricPriceKey - stored by *denom* and *block*
func GetHistoricPriceKey(denom string, blockNum uint64) (key []byte) {
	key = append(key, KeyPrefixHistoricPrice...)
	key = appendDenomAndBlockToKey(key, denom, blockNum)
	return append(key, 0) // append 0 for null-termination
}

func appendDenomAndBlockToKey(key []byte, denom string, blockNum uint64) []byte {
	key = append(key, []byte(denom)...)
	block := make([]byte, 8)
	binary.LittleEndian.PutUint64(block, blockNum)
	key = append(key, block...)
	return key
}

func ParseDemonFromHistoricPriceKey(key []byte) string {
	return string(key[len(KeyPrefixExchangeRate) : len(key)-9])
}

func ParseBlockFromHistoricPriceKey(key []byte) uint64 {
	return binary.LittleEndian.Uint64(key[len(key)-9 : len(key)-1])
}
