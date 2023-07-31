package ugov

const (
	// ModuleName defines the module name
	ModuleName = "ugov"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

// store key prefixes
var (
	KeyMinGasPrice       = []byte{0x01}
	KeyEmergencyGroup    = []byte{0x02}
	KeyInflationParams   = []byte{0x03}
	KeyInflationCycleEnd = []byte{0x04}
)
