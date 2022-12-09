package incentive

const (
	// ModuleName defines the module name
	ModuleName = "incentive"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// StoreKey defines the query route
	QuerierRoute = ModuleName
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
	KeyPrefixUpcomingIncentiveProgram  = []byte{0x02}
	KeyPrefixOngoingIncentiveProgram   = []byte{0x03}
	KeyPrefixCompletedIncentiveProgram = []byte{0x04}
	KeyPrefixNextProgramID             = []byte{0x05}
	KeyPrefixLastRewardsTime           = []byte{0x06}
	KeyPrefixTotalBonded               = []byte{0x07}
	KeyPrefixBondAmount                = []byte{0x08}
	KeyPrefixPendingReward             = []byte{0x09}
	KeyPrefixRewardBasis               = []byte{0x0A}
	KeyPrefixRewardAccumulator         = []byte{0x0B}
	KeyPrefixUnbonding                 = []byte{0x0C}
)
