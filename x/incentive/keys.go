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
	// BondTierUnspecified is used in functions which query unbondings, to indicate that all tiers should be counted
	BondTierUnspecified BondTier = iota
	BondTierShort
	BondTierMiddle
	BondTierLong
)

const (
	ProgramStatusUpcoming ProgramStatus = iota
	ProgramStatusOngoing
	ProgramStatusCompleted
)

func (p ProgramStatus) Validate() error {
	if p > ProgramStatusCompleted {
		return ErrInvalidProgramStatus.Wrapf("%d", p)
	}
	return nil
}

func (t BondTier) Validate(canBeUnspecified bool) error {
	if t > BondTierLong || (!canBeUnspecified && t == BondTierUnspecified) {
		return ErrInvalidTier.Wrapf("%d", t)
	}
	return nil
}
