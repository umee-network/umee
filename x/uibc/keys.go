package uibc

const (
	// ModuleName defines the module name
	ModuleName = "uibc"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	KeyPrefixForIBCDenom = []byte{0x01}
	KeyTotalOutflowSum   = []byte("TotalOutflowSum")
	// ParamsKey is the key to query all gov params
	ParamsKey = []byte("params")
	// QuotaExpiresKey is the key to store the expire time of ibc-transfer quota
	QuotaExpiresKey = []byte("QuotaExpiresKey")
)

func CreateKeyForQuotaOfIBCDenom(ibcDenom string) []byte {
	// interestScalarPrefix | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixForIBCDenom...)
	key = append(key, []byte(ibcDenom)...)
	return append(key, 0) // append 0 for null-termination
}
