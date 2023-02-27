package incentive

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

type (
	BondTier      uint8
	ProgramStatus uint8
)

const (
	BondTierLong BondTier = iota
	BondTierMiddle
	BondTierShort
)

const (
	ProgramStatusUpcoming ProgramStatus = iota
	ProgramStatusOngoing
	ProgramStatusCompleted
)
