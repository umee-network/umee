package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/umee-network/umee/v4/util"
	"github.com/umee-network/umee/v4/x/incentive"
)

// KVStore key prefixes
var (
	keyPrefixParams                    = []byte{0x01}
	keyPrefixUpcomingIncentiveProgram  = []byte{0x02}
	keyPrefixOngoingIncentiveProgram   = []byte{0x03}
	keyPrefixCompletedIncentiveProgram = []byte{0x04}
	keyNextProgramID                   = []byte{0x05}
	keyLastRewardsTime                 = []byte{0x06}
	keyPrefixRewardTracker             = []byte{0x07}
	keyPrefixRewardAccumulator         = []byte{0x08}
	keyPrefixUnbondings                = []byte{0x09}
	keyPrefixBondAmount                = []byte{0x0A}
	keyPrefixUnbondAmount              = []byte{0x0B}
	keyPrefixTotalBonded               = []byte{0x0C}
	keyPrefixTotalUnbonding            = []byte{0x0D}
)

// keyIncentiveProgram returns a KVStore key for an incentive program.
func keyIncentiveProgram(id uint32, status incentive.ProgramStatus) []byte {
	// programPrefix (one of three) | id
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, id)
	var prefix []byte
	switch status {
	case incentive.ProgramStatusUpcoming:
		prefix = keyPrefixUpcomingIncentiveProgram
	case incentive.ProgramStatusOngoing:
		prefix = keyPrefixOngoingIncentiveProgram
	case incentive.ProgramStatusCompleted:
		prefix = keyPrefixCompletedIncentiveProgram
	default:
		panic("invalid incentive program status in key")
	}
	return util.ConcatBytes(0, prefix, bz)
}

// keyTotalBonded returns a KVStore key for total bonded uTokens of a given denom.
func keyTotalBonded(denom string) []byte {
	// totalBondedPrefix | denom | 0x00
	return util.ConcatBytes(1, keyPrefixTotalBonded, []byte(denom))
}

// keyTotalUnbonding returns a KVStore key for total unbonding uTokens of a given denom.
func keyTotalUnbonding(denom string) []byte {
	// totalUnbondingPrefix | denom | 0x00
	return util.ConcatBytes(1, keyPrefixTotalUnbonding, []byte(denom))
}

// keyBondAmount returns a KVStore key for bonded amounts for a uToken denom and account.
func keyBondAmount(addr sdk.AccAddress, denom string) []byte {
	// bondPrefix | lengthprefixed(addr) | denom | 0x00
	return util.ConcatBytes(1, keyBondAmountNoDenom(addr), []byte(denom))
}

// keyBondAmountNoDenom returns the common prefix used by all uTokens bonded to a given account.
func keyBondAmountNoDenom(addr sdk.AccAddress) []byte {
	// bondPrefix | lengthprefixed(addr)
	return util.ConcatBytes(0, keyPrefixBondAmount, address.MustLengthPrefix(addr))
}

// keyUnbondAmount returns a KVStore key for unbonding amounts for a uToken denom and account.
func keyUnbondAmount(addr sdk.AccAddress, denom string) []byte {
	// unbondPrefix | lengthprefixed(addr) | denom | 0x00
	return util.ConcatBytes(1, keyUnbondAmountNoDenom(addr), []byte(denom))
}

// keyUnbondAmountNoDenom returns the common prefix used by all uTokens unbonding from to a given account.
func keyUnbondAmountNoDenom(addr sdk.AccAddress) []byte {
	// unbondPrefix | lengthprefixed(addr)
	return util.ConcatBytes(0, keyPrefixUnbondAmount, address.MustLengthPrefix(addr))
}

// keyRewardAccumulator returns a KVStore key for a RewardAccumulator denom for a bonded uToken.
func keyRewardAccumulator(bondedDenom string) []byte {
	// rewardAccumulatorPrefix | bondedDenom | 0x00
	return util.ConcatBytes(1, keyPrefixRewardAccumulator, []byte(bondedDenom))
}

// keyRewardTracker returns a KVStore key for a single reward tracker denom for an account and bonded uToken.
func keyRewardTracker(addr sdk.AccAddress, bondedDenom string) []byte {
	// rewardTrackerPrefix | lengthprefixed(addr) | bondedDenom | 0x00
	return util.ConcatBytes(1, keyRewardTrackerNoDenom(addr), []byte(bondedDenom))
}

// keyRewardTrackerNoDenom returns the common prefix used by all reward trackers on a given account.
func keyRewardTrackerNoDenom(addr sdk.AccAddress) []byte {
	// rewardTrackerPrefix | lengthprefixed(addr)
	return util.ConcatBytes(0, keyPrefixRewardTracker, address.MustLengthPrefix(addr))
}

// keyUnbondings returns a key to store all active unbondings on an account for a given uToken.
func keyUnbondings(addr sdk.AccAddress, denom string) []byte {
	// unbondingPrefix | lengthprefixed(addr) | denom | 0x00
	return util.ConcatBytes(1, keyUnbondingsNoDenom(addr), []byte(denom))
}

// keyUnbondingsNoDenom returns the common prefix used by all unbondings from a given account.
func keyUnbondingsNoDenom(addr sdk.AccAddress) []byte {
	// unbondingPrefix | lengthprefixed(addr)
	return util.ConcatBytes(0, keyPrefixUnbondings, address.MustLengthPrefix(addr))
}
