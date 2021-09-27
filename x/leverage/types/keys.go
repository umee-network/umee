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

/*

// Comment back in when no longer hard-coding associations between tokens

// returns store key used to store a specific asset coin's associated utoken denom
func assetAssociatedUtokenKey(coin sdk.Coin) []byte {
	return prefixDenomStoreKey(AssetAssociatedUtokenPrefix, coin)
	// intent: store[0x01+"uatom"] = "u/uatom"
}

// returns store key used to store a specific utoken coin's associated asset denom
func utokenAssociatedAssetKey(coin sdk.Coin) []byte {
	return prefixDenomStoreKey(UtokenAssociatedAssetPrefix, coin)
	// intent: store[0x02+"u/uatom"] = "uatom"
}

// prefixDenomStoreKey turns a coin to the key used to store specific info by appending a prefix to the denom.
// Returns empty bytes on invalid denom. Modeled after x/auth/types.AccountStoreKey
func prefixDenomStoreKey(prefix byte, coin sdk.Coin) []byte {
	if sdk.ValidateDenom(coin.Denom) != nil {
		// Denom did not match ^[a-z][a-z0-9/]{2,63}$
		return []byte{}
	}
	// example: byte(0x01) + []byte("uatom")
	key := []byte{prefix}
	key = append(key, []byte(coin.Denom)...)
	return key
	// Note: After IBC enable, want a reliable way to convert token denominations to bytes, such that
	//	a) Each token type (e.g. Atom) has only one prefix/key, regardless of the path it took via IBC
	//		to get to umee
	//	b) Tokens cannot be spoofed (e.g. EvilChain cannot name a token 'atom', mint their own, and deposit)
}

*/
