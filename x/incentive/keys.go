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
	ProgramStatus uint8
)

const (
	ProgramStatusNotExist ProgramStatus = iota
	ProgramStatusUpcoming
	ProgramStatusOngoing
	ProgramStatusCompleted
)

func (p ProgramStatus) Validate() error {
	if p == ProgramStatusNotExist || p > ProgramStatusCompleted {
		return ErrInvalidProgramStatus.Wrapf("%d", p)
	}
	return nil
}
