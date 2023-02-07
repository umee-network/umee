package keeper

import (
	"encoding/binary"

	"github.com/umee-network/umee/v4/util"
)

// KVStore key prefixes
var (
	// Individually store params from MsgGovSetParams
	keyPrefixParamMaxUnbondings           = []byte{0x01, 0x01}
	keyPrefixParamUnbondingDurationLong   = []byte{0x01, 0x02}
	keyPrefixParamUnbondingDurationMiddle = []byte{0x01, 0x03}
	keyPrefixParamUnbondingDurationShort  = []byte{0x01, 0x04}
	keyPrefixParamTierWeightShort         = []byte{0x01, 0x05}
	keyPrefixParamTierWeightMiddle        = []byte{0x01, 0x06}
	keyPrefixParamCommunityFundAddress    = []byte{0x01, 0x07}

	// Regular state
	keyPrefixUpcomingIncentiveProgram  = []byte{0x02}
	keyPrefixOngoingIncentiveProgram   = []byte{0x03}
	keyPrefixCompletedIncentiveProgram = []byte{0x04}
	keyPrefixNextProgramID             = []byte{0x05}
	keyPrefixLastRewardsTime           = []byte{0x06}
	keyPrefixTotalBonded               = []byte{0x07}
	keyPrefixBondAmount                = []byte{0x08}
	keyPrefixRewardBasis               = []byte{0x09}
	keyPrefixRewardAccumulator         = []byte{0x0A}
	keyPrefixUnbonding                 = []byte{0x0B}
)

// keyUpcomingIncentiveProgram returns a KVStore key for getting and setting an incentive program.
func keyUpcomingIncentiveProgram(id uint32) []byte {
	// programPrefix | id
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, id)
	return util.ConcatBytes(0, keyPrefixUpcomingIncentiveProgram, bz)
}

// keyOngoingIncentiveProgram returns a KVStore key for getting and setting an incentive program.
func keyOngoingIncentiveProgram(id uint32) []byte {
	// programPrefix | id
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, id)
	return util.ConcatBytes(0, keyPrefixOngoingIncentiveProgram, bz)
}

// keyCompletedIncentiveProgram returns a KVStore key for getting and setting an incentive program.
func keyCompletedIncentiveProgram(id uint32) []byte {
	// programPrefix | id
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, id)
	return util.ConcatBytes(0, keyPrefixCompletedIncentiveProgram, bz)
}

// keyTotalBonded returns a KVStore key for getting and setting total bonded amounts for a uToken.
func keyTotalBonded(denom string) []byte {
	// totalBondedPrefix | denom | 0x00 for null-termination
	return util.ConcatBytes(1, keyPrefixTotalBonded, []byte(denom))
}

/*
// keyBondAmount returns a KVStore key for getting and setting bonded amounts for a uToken on a single account.
func keyBondAmount(addr sdk.AccAddress, uTokenDenom string) []byte {
	// bondPrefix | lengthprefixed(addr) | denom | 0x00 for null-termination
	return util.ConcatBytes(1, keyBondAmountAmountNoDenom(addr), []byte(uTokenDenom))
}

// keyBondAmountAmountNoDenom returns the common prefix used by all uTokens bonded to a given account.
func keyBondAmountAmountNoDenom(addr sdk.AccAddress) []byte {
	// bondPrefix | lengthprefixed(addr)
	return util.ConcatBytes(0, keyPrefixBondAmount, address.MustLengthPrefix(addr))
}
*/
