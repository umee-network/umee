package incentive

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/umee-network/umee/v4/util"
)

const (
	// ModuleName defines the module name
	ModuleName = "incentive"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// StoreKey defines the query route
	QuerierRoute = ModuleName

	// RouterKey is the message route
	RouterKey = ModuleName
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
	KeyPrefixParamCommunityFundAddress    = []byte{0x01, 0x07}

	// Regular state
	KeyPrefixUpcomingIncentiveProgram  = []byte{0x02}
	KeyPrefixOngoingIncentiveProgram   = []byte{0x03}
	KeyPrefixCompletedIncentiveProgram = []byte{0x04}
	KeyPrefixNextProgramID             = []byte{0x05}
	KeyPrefixLastRewardsTime           = []byte{0x06}
	KeyPrefixTotalBonded               = []byte{0x07}
	KeyPrefixBondAmount                = []byte{0x08}
	KeyPrefixRewardBasis               = []byte{0x09}
	KeyPrefixRewardAccumulator         = []byte{0x0A}
	KeyPrefixUnbonding                 = []byte{0x0B}
)

// KeyUpcomingIncentiveProgram returns a KVStore key for getting and setting an incentive program.
func KeyUpcomingIncentiveProgram(id uint32) []byte {
	// programPrefix | id
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, id)
	return util.ConcatBytes(0, KeyPrefixUpcomingIncentiveProgram, bz)
}

// KeyOngoingIncentiveProgram returns a KVStore key for getting and setting an incentive program.
func KeyOngoingIncentiveProgram(id uint32) []byte {
	// programPrefix | id
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, id)
	return util.ConcatBytes(0, KeyPrefixOngoingIncentiveProgram, bz)
}

// KeyCompletedIncentiveProgram returns a KVStore key for getting and setting an incentive program.
func KeyCompletedIncentiveProgram(id uint32) []byte {
	// programPrefix | id
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, id)
	return util.ConcatBytes(0, KeyPrefixCompletedIncentiveProgram, bz)
}

// KeyTotalBonded returns a KVStore key for getting and setting total bonded amounts for a uToken.
func KeyTotalBonded(denom string) []byte {
	// totalBondedPrefix | denom | 0x00 for null-termination
	return util.ConcatBytes(1, KeyPrefixCompletedIncentiveProgram, []byte(denom))
}

// KeyBondAmount returns a KVStore key for getting and setting bonded amounts for a uToken on a single account.
func KeyBondAmount(addr sdk.AccAddress, uTokenDenom string) []byte {
	// bondPrefix | lengthprefixed(addr) | denom | 0x00 for null-termination
	return util.ConcatBytes(1, KeyBondAmountAmountNoDenom(addr), []byte(uTokenDenom))
}

// KeyBondAmountAmountNoDenom returns the common prefix used by all uTokens bonded to a given account.
func KeyBondAmountAmountNoDenom(addr sdk.AccAddress) []byte {
	// bondPrefix | lengthprefixed(addr)
	return util.ConcatBytes(0, KeyPrefixBondAmount, address.MustLengthPrefix(addr))
}
