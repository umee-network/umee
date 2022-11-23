package ibctransfer

const (
	// ModuleName defines the module name
	ModuleName = "ibcratelimit"

	// StoreKey defines the primary module store key
	StoreKey = "ratelimit"

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	KeyPrefixForIBCDenom = []byte{0x01}
	KeyTotalOutflowSum   = []byte("TotalOutflowSum")
)

func CreateKeyForRateLimitOfIBCDenom(ibcDenom string) []byte {
	// interestScalarPrefix | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixForIBCDenom...)
	key = append(key, []byte(ibcDenom)...)
	return append(key, 0) // append 0 for null-termination
}
