package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/umee-network/umee/v3/util"
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

// KeyMedian - stored by *denom* and *block*
func KeyMedian(denom string, blockNum uint64) (key []byte) {
	return util.ConcatBytes(0, KeyPrefixMedian, []byte(denom), uintWithNullPrefix(blockNum))
}

// KeyMedianDeviation - stored by *denom* and *block*
func KeyMedianDeviation(denom string, blockNum uint64) (key []byte) {
	return util.ConcatBytes(0, KeyPrefixMedianDeviation, []byte(denom), uintWithNullPrefix(blockNum))
}

// KeyHistoricPrice - stored by *denom* and *block*
func KeyHistoricPrice(denom string, blockNum uint64) (key []byte) {
	return util.ConcatBytes(0, KeyPrefixHistoricPrice, []byte(denom), uintWithNullPrefix(blockNum))
}

func uintWithNullPrefix(n uint64) []byte {
	bz := make([]byte, 9)
	binary.LittleEndian.PutUint64(bz[1:], n)
	return bz
}

func ParseDemonFromHistoricPriceKey(key []byte) string {
	return string(key[len(KeyPrefixExchangeRate) : len(key)-9])
}

func ParseBlockFromHistoricPriceKey(key []byte) uint64 {
	return binary.LittleEndian.Uint64(key[len(key)-9 : len(key)-1])
}
