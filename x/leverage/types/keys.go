package types

const (
	// ModuleName defines the module name
	ModuleName = "leverage"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore key prefixes
var (
	KeyPrefixTokenDenom  = []byte{0x01}
	KeyPrefixUTokenDenom = []byte{0x02}
)

// CreateTokenDenomKey returns a KVStore key for getting and storing a token's
// associated uToken denomination.
func CreateTokenDenomKey(tokenDenom string) []byte {
	var key []byte
	key = append(key, KeyPrefixTokenDenom...)
	key = append(key, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateUTokenDenomKey returns a KVStore key for getting and storing a uToken's
// associated token denomination.
func CreateUTokenDenomKey(uTokenDenom string) []byte {
	var key []byte
	key = append(key, KeyPrefixUTokenDenom...)
	key = append(key, []byte(uTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}
