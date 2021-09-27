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

// CreateTokenDenomPrefix returns a KVStore prefix for getting and storing a
// token's uToken denomination.
func CreateTokenDenomPrefix(tokenDenom string) []byte {
	key := append(KeyPrefixTokenDenom, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateUTokenDenomPrefix returns a KVStore prefix for getting and storing a
// uToken's token denomination.
func CreateUTokenDenomPrefix(uTokenDenom string) []byte {
	key := append(KeyPrefixUTokenDenom, []byte(uTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}
