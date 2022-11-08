package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	gogotypes "github.com/gogo/protobuf/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "incentive"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

// KVStore key prefixes
var (
	// Individually store params from MsgGovSetParams
	KeyPrefixParamMaxUnbondings           = []byte{0x01, 0x01}
	KeyPrefixParamUnbondingDurationLong   = []byte{0x01, 0x02}
	KeyPrefixParamUnbondingDurationMiddle = []byte{0x01, 0x03}
	KeyPrefixParamUnbondingDurationShort  = []byte{0x01, 0x04}
	KeyPrefixParamTierWeightShort         = []byte{0x01, 0x05}
	KeyPrefixParamTierWeightMiddle        = []byte{0x01, 0x06}

	// Regular state
	KeyPrefixIncentiveProgram  = []byte{0x02}
	KeyPrefixNextProgramID     = []byte{0x03}
	KeyPrefixTotalBonded       = []byte{0x04}
	KeyPrefixBondAmount        = []byte{0x05}
	KeyPrefixPendingReward     = []byte{0x06}
	KeyPrefixRewardBasis       = []byte{0x07}
	KeyPrefixRewardAccumulator = []byte{0x08}
	KeyPrefixUnbonding         = []byte{0x09}
)

// CreateIncentiveProgramKey returns a KVStore key for getting and setting an IncentiveProgram.
func CreateIncentiveProgramKey(cdc codec.Codec, id uint32) []byte {
	// prefix | id
	var key []byte
	key = append(key, KeyPrefixIncentiveProgram...)

	// note: use of codec required by using a uint32 as part of a key
	bz, err := cdc.Marshal(&gogotypes.UInt32Value{Value: id})
	if err != nil {
		panic(err)
	}

	key = append(key, bz...)
	return key
}

// CreateTotalBondedKey returns a KVStore key for getting and setting the
// total bonded amount tracker for a single uToken.
func CreateTotalBondedKey(uTokenDenom string) []byte {
	// prefix | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixTotalBonded...)
	key = append(key, []byte(uTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateBondedAmountKey returns a KVStore key for getting and setting a
// bonded amount for a denom and address.
func CreateBondedAmountKey(addr sdk.AccAddress, uTokenDenom string) []byte {
	// prefix | lengthprefixed(addr) | denom | 0x00
	key := CreateBondedAmountKeyNoDenom(addr)
	key = append(key, []byte(uTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateBondedAmountKeyNoDenom returns the common prefix used by all bonded tokens
// associated with a given address.
func CreateBondedAmountKeyNoDenom(addr sdk.AccAddress) []byte {
	// prefix | lengthprefixed(addr)
	var key []byte
	key = append(key, KeyPrefixBondAmount...)
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// CreatePendingRewardKey returns a KVStore key for getting and setting the
// amount of rewards for a given address which have been calculated but not yet claimed.
func CreatePendingRewardKey(addr sdk.AccAddress) []byte {
	// prefix | lengthprefixed(addr)
	var key []byte
	key = append(key, KeyPrefixPendingReward...)
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// CreateRewardBasisKey returns a KVStore key for getting and setting the
// reward basis for a given bonded uToken denom and address.
func CreateRewardBasisKey(addr sdk.AccAddress, uTokenDenom string) []byte {
	// prefix | lengthprefixed(addr) | denom | 0x00
	key := CreateRewardBasisKeyNoDenom(addr)
	key = append(key, []byte(uTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateRewardBasisKeyNoDenom returns the common prefix used by all reward bases
// associated with a given address.
func CreateRewardBasisKeyNoDenom(addr sdk.AccAddress) []byte {
	// prefix | lengthprefixed(addr)
	var key []byte
	key = append(key, KeyPrefixRewardBasis...)
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// CreateRewardAccumulatorKey returns a KVStore key for getting and setting the
// reward basis tracker for a single uToken.
func CreateRewardAccumulatorKey(uTokenDenom string) []byte {
	// prefix | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixRewardAccumulator...)
	key = append(key, []byte(uTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateUnbondingKey returns the common prefix used by all reward bases
// associated with a given address.
func CreateUnbondingKey(cdc codec.Codec, addr sdk.AccAddress, height uint64) []byte {
	// prefix | lengthprefixed(addr) | height
	var key []byte
	key = append(key, KeyPrefixUnbonding...)
	key = append(key, address.MustLengthPrefix(addr)...)

	// note: use of codec required by using a uint64 as part of a key
	bz, err := cdc.Marshal(&gogotypes.UInt64Value{Value: height})
	if err != nil {
		panic(err)
	}
	key = append(key, bz...)

	return key
}
