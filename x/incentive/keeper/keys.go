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
	keyPrefixRewardTracker             = []byte{0x09}
	keyPrefixRewardAccumulator         = []byte{0x0A}
	keyPrefixUnbondings                = []byte{0x0B}
	keyPrefixTotalUnbonding            = []byte{0x0C}
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

// keyTotalBonded returns a KVStore key for total bonded uTokens for a single tier.
func keyTotalBonded(denom string, tier incentive.BondTier) []byte {
	// totalBondedPrefix | denom | 0x00 | tier
	return util.ConcatBytes(0, keyTotalBondedNoTier(denom), []byte{byte(tier)})
}

// keyTotalBondedNoTier returns the common prefix used by all TotalBonds for a uToken denom.
func keyTotalBondedNoTier(denom string) []byte {
	// totalBondedPrefix | denom | 0x00
	return util.ConcatBytes(1, keyPrefixTotalBonded, []byte(denom))
}

// keyTotalUnbonding returns a KVStore key for total unbonding uTokens for a single tier.
func keyTotalUnbonding(denom string, tier incentive.BondTier) []byte {
	// totalUnbondingPrefix | denom | 0x00 | tier
	return util.ConcatBytes(0, keyTotalUnbondingNoTier(denom), []byte{byte(tier)})
}

// keyTotalUnbondingNoTier returns the common prefix used by all total unbondings for a uToken denom.
func keyTotalUnbondingNoTier(denom string) []byte {
	// totalUnbondingPrefix | denom | 0x00
	return util.ConcatBytes(1, keyPrefixTotalUnbonding, []byte(denom))
}

// keyBondAmount returns a KVStore key for bonded amounts for a uToken denom, account, and tier.
func keyBondAmount(addr sdk.AccAddress, denom string, tier incentive.BondTier) []byte {
	// bondPrefix | lengthprefixed(addr) | denom | 0x00 | tier
	return util.ConcatBytes(0, keyBondAmountNoTier(addr, denom), []byte{byte(tier)})
}

// keyBondAmountNoTier returns the common prefix used by all uTokens bonded to a given account and denom.
func keyBondAmountNoTier(addr sdk.AccAddress, denom string) []byte {
	// bondPrefix | lengthprefixed(addr) | denom | 0x00
	return util.ConcatBytes(1, keyBondAmountNoDenom(addr), []byte(denom))
}

// keyBondAmountNoDenom returns the common prefix used by all uTokens bonded to a given account.
func keyBondAmountNoDenom(addr sdk.AccAddress) []byte {
	// bondPrefix | lengthprefixed(addr)
	return util.ConcatBytes(0, keyPrefixBondAmount, address.MustLengthPrefix(addr))
}

// keyRewardAccumulator returns a KVStore key for a single RewardAccumulator denom for a bonded uToken
// denom and tier.
func keyRewardAccumulator(bondedDenom, rewardDenom string, tier incentive.BondTier) []byte {
	// rewardAccumulatorPrefix | bondedDenom | 0x00 | tier | rewardDenom | 0x00
	return util.ConcatBytes(1, keyRewardAccumulatorNoReward(bondedDenom, tier), []byte(rewardDenom))
}

// keyRewardAccumulatorNoReward returns the common prefix used by all RewardAccumulators for a bonded uToken
// denom and tier.
func keyRewardAccumulatorNoReward(bondedDenom string, tier incentive.BondTier) []byte {
	// rewardAccumulatorPrefix | bondedDenom | 0x00 | tier
	return util.ConcatBytes(0, keyRewardAccumulatorNoTier(bondedDenom), []byte{byte(tier)})
}

// keyRewardAccumulatorNoTier returns the common prefix used by all RewardAccumulators for a bonded uToken
// denom.
func keyRewardAccumulatorNoTier(bondedDenom string) []byte {
	// rewardAccumulatorPrefix | bondedDenom | 0x00
	return util.ConcatBytes(1, keyPrefixRewardAccumulator, []byte(bondedDenom))
}

// keyRewardTracker returns a KVStore key for a single reward tracker denom for an account and bonded uToken
// denom and tier.
func keyRewardTracker(addr sdk.AccAddress, bondedDenom, rewardDenom string, tier incentive.BondTier) []byte {
	// rewardTrackerPrefix | lengthprefixed(addr) | bondedDenom | 0x00 | tier | rewardDenom | 0x00
	return util.ConcatBytes(1, keyRewardTrackerNoReward(addr, bondedDenom, tier), []byte(rewardDenom))
}

// keyRewardTrackerNoReward returns a KVStore key for a single reward tracker denom for an account and bonded uToken
// denom and tier.
func keyRewardTrackerNoReward(addr sdk.AccAddress, bondedDenom string, tier incentive.BondTier) []byte {
	// rewardTrackerPrefix | lengthprefixed(addr) | bondedDenom | 0x00 | tier
	return util.ConcatBytes(0, keyRewardTrackerNoTier(addr, bondedDenom), []byte{byte(tier)})
}

// keyRewardTrackerNoTier returns the common prefix used by all reward trackers for an account and bonded uToken
// denom across all unbonding tiers.
func keyRewardTrackerNoTier(addr sdk.AccAddress, bondedDenom string) []byte {
	// rewardTrackerPrefix | lengthprefixed(addr) | bondedDenom | 0x00
	return util.ConcatBytes(1, keyRewardTrackerNoDenom(addr), []byte(bondedDenom))
}

// keyRewardTrackerNoDenom returns the common prefix used by all reward trackers on a given account.
func keyRewardTrackerNoDenom(addr sdk.AccAddress) []byte {
	// rewardTrackerPrefix | lengthprefixed(addr)
	return util.ConcatBytes(0, keyPrefixRewardTracker, address.MustLengthPrefix(addr))
}

// keyUnbondings returns a key to store all active unbondings on an account for a given denom and tier
func keyUnbondings(addr sdk.AccAddress, denom string, tier incentive.BondTier) []byte {
	// unbondingPrefix | lengthprefixed(addr) | denom | 0x00 | tier
	return util.ConcatBytes(0, keyUnbondingsNoTier(addr, denom), []byte{byte(tier)})
}

// keyUnbondingsNoTier returns the common prefix used by all unbondings from a given account and denom.
func keyUnbondingsNoTier(addr sdk.AccAddress, denom string) []byte {
	// unbondingPrefix | lengthprefixed(addr) | denom | 0x00
	return util.ConcatBytes(1, keyUnbondingsNoDenom(addr), []byte(denom))
}

// keyUnbondingsNoDenom returns the common prefix used by all unbondings from a given account.
func keyUnbondingsNoDenom(addr sdk.AccAddress) []byte {
	// unbondingPrefix | lengthprefixed(addr)
	return util.ConcatBytes(0, keyPrefixUnbondings, address.MustLengthPrefix(addr))
}
