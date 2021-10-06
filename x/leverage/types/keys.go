package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
	KeyPrefixTokenDenom        = []byte{0x01}
	KeyPrefixUTokenDenom       = []byte{0x02}
	KeyPrefixRegisteredToken   = []byte{0x03}
	KeyPrefixLoanToken         = []byte{0x04}
	KeyPrefixCollateralSetting = []byte{0x05}
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

// CreateRegisteredTokenKey returns a KVStore key for getting and setting an Asset.
func CreateRegisteredTokenKey(baseTokenDenom string) []byte {
	var key []byte
	key = append(key, KeyPrefixRegisteredToken...)
	key = append(key, []byte(baseTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateLoanKey returns a KVStore key for getting and setting a Loan in a single denom and borrower address
func CreateLoanKey(borrowerAddr sdk.AccAddress, tokenDenom string) []byte {
	// loanprefix | lengthprefixed(borrowerAddr) | denom
	var key []byte
	key = append(key, KeyPrefixLoanToken...)
	addr := []byte(borrowerAddr.String())
	key = append(key, byte(len(addr))) // simple length prefix since len(addr.String()) is always < 255
	key = append(key, addr...)
	key = append(key, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// LoanKeyRange returns start/end keys for creating an iterator over all of an account's open loans.
func LoanKeyRange(borrowerAddr sdk.AccAddress) ([]byte, []byte) {
	//	Question: Is this the right way to derive a range for an sdk.Iterator?
	//
	//	e.g. if KeyPrefixLoanToken | lengthPrefixed(borrowerAddr.String()) were resolved to
	//		0x04 | 0x03 0x41 0x42 0x43 (example address string simplified to "ABC")
	//	then the iterator start/end would be
	//		0x04 | 0x03 0x41 0x42 0x43 (inclusive start)
	//		0x04 | 0x03 0x41 0x42 0x44 (exclusive end)
	//	and keys like the following would fall within the range
	//		0x04 | 0x03 0x41 0x42 0x43 | ... (any key that has prefix)
	//
	//	I couldn't find documentation on this behavior but it seems like
	//		how it would reasonably work if sdk.Iterators are prefix-friendly.
	startkey := CreateLoanKey(borrowerAddr, "") // loanprefix | lengthprefixed(borrowerAddr)
	endkey := CreateLoanKey(borrowerAddr, "")   // loanprefix | lengthprefixed(borrowerAddr)
	endkey[len(endkey)-1]++                     // last byte of borrowerAddr.String() shouldn't ever be 255
	return startkey, endkey
}

// CreateCollateralSettingKey returns a KVStore key for getting and setting a borrower's
// collateral setting for a single uToken
func CreateCollateralSettingKey(borrowerAddr sdk.AccAddress, uTokenDenom string) []byte {
	var key []byte
	key = append(key, KeyPrefixCollateralSetting...)
	addr := []byte(borrowerAddr.String())
	key = append(key, byte(len(addr))) // simple length prefix since len(addr.String()) is always < 255
	key = append(key, addr...)
	key = append(key, []byte(uTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}
