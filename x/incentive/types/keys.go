package types

const (
	// ModuleName defines the module name
	ModuleName = "incentive"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore key prefixes
var (
	KeyPrefixIncentiveProgram = []byte{0x01}
	// TODO: more
)

// CreateIncentiveProgramKey returns a KVStore key for getting and setting an IncentiveProgram.
func CreateIncentiveProgramKey(id uint32) []byte {
	// programprefix | id
	var key []byte
	key = append(key, KeyPrefixIncentiveProgram...)

	// key = append(key, bz...) // TODO: marshal uint32 to bytes without k.cdc
	return key
}

// TODO: more
