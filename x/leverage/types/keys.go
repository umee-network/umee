package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
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
	KeyPrefixRegisteredToken     = []byte{0x01}
	KeyPrefixAdjustedBorrow      = []byte{0x02}
	KeyPrefixCollateralAmount    = []byte{0x04}
	KeyPrefixReserveAmount       = []byte{0x05}
	KeyPrefixLastInterestTime    = []byte{0x06}
	KeyPrefixBadDebt             = []byte{0x07}
	KeyPrefixInterestScalar      = []byte{0x08}
	KeyPrefixAdjustedTotalBorrow = []byte{0x09}
	KeyPrefixUtokenSupply        = []byte{0x0A}
)

// CreateRegisteredTokenKey returns a KVStore key for getting and setting a Token.
func CreateRegisteredTokenKey(baseTokenDenom string) []byte {
	// assetprefix | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixRegisteredToken...)
	key = append(key, []byte(baseTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateAdjustedBorrowKey returns a KVStore key for getting and setting an
// adjusted borrow for a denom and borrower address.
func CreateAdjustedBorrowKey(borrowerAddr sdk.AccAddress, tokenDenom string) []byte {
	// borrowprefix | lengthprefixed(borrowerAddr) | denom | 0x00
	key := CreateAdjustedBorrowKeyNoDenom(borrowerAddr)
	key = append(key, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateAdjustedBorrowKeyNoDenom returns the common prefix used by all borrows
// associated with a given borrower address.
func CreateAdjustedBorrowKeyNoDenom(borrowerAddr sdk.AccAddress) []byte {
	// borrowprefix | lengthprefixed(borrowerAddr)
	var key []byte
	key = append(key, KeyPrefixAdjustedBorrow...)
	key = append(key, address.MustLengthPrefix(borrowerAddr)...)
	return key
}

// CreateCollateralAmountKey returns a KVStore key for getting and setting the amount of
// collateral stored for a lender in a given denom.
func CreateCollateralAmountKey(lenderAddr sdk.AccAddress, uTokenDenom string) []byte {
	// collateralPrefix | lengthprefixed(lenderAddr) | denom | 0x00
	key := CreateCollateralAmountKeyNoDenom(lenderAddr)
	key = append(key, []byte(uTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateCollateralAmountKeyNoDenom returns the common prefix used by all collateral associated
// with a given lender address.
func CreateCollateralAmountKeyNoDenom(lenderAddr sdk.AccAddress) []byte {
	// collateralPrefix | lengthprefixed(lenderAddr)
	key := CreateCollateralAmountKeyNoAddress()
	key = append(key, address.MustLengthPrefix(lenderAddr)...)
	return key
}

// CreateCollateralAmountKeyNoAddress returns a safe copy of collateralPrefix
func CreateCollateralAmountKeyNoAddress() []byte {
	// collateralPrefix
	var key []byte
	key = append(key, KeyPrefixCollateralAmount...)
	return key
}

// CreateReserveAmountKey returns a KVStore key for getting and setting the amount reserved of a a given token.
func CreateReserveAmountKey(tokenDenom string) []byte {
	// reserveamountprefix | denom | 0x00
	key := CreateReserveAmountKeyNoDenom()
	key = append(key, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateReserveAmountKeyNoDenom returns a safe copy of reserveAmountPrefix
func CreateReserveAmountKeyNoDenom() []byte {
	// reserveAmountPrefix
	var key []byte
	key = append(key, KeyPrefixReserveAmount...)
	return key
}

// CreateLastInterestTimeKey returns a KVStore key for getting and setting the amount reserved of a a given token.
func CreateLastInterestTimeKey() []byte {
	// lastinterestprefix
	var key []byte
	key = append(key, KeyPrefixLastInterestTime...)
	return key
}

// CreateBadDebtKey returns a KVStore key for tracking an address with unpaid bad debt
func CreateBadDebtKey(denom string, borrowerAddr sdk.AccAddress) []byte {
	// badDebtAddrPrefix | lengthprefixed(borrowerAddr) | denom | 0x00
	key := CreateBadDebtKeyNoAddress()
	key = append(key, address.MustLengthPrefix(borrowerAddr)...)
	key = append(key, []byte(denom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateBadDebtKeyNoAddress returns a safe copy of bad debt prefix
func CreateBadDebtKeyNoAddress() []byte {
	// badDebtPrefix
	var key []byte
	key = append(key, KeyPrefixBadDebt...)
	return key
}

// CreateInterestScalarKey returns a KVStore key for getting and setting the interest scalar for a
// given token.
func CreateInterestScalarKey(tokenDenom string) []byte {
	// interestScalarPrefix | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixInterestScalar...)
	key = append(key, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateAdjustedTotalBorrowKey returns a KVStore key for getting and setting the total ajdusted borrows for
// a given token.
func CreateAdjustedTotalBorrowKey(tokenDenom string) []byte {
	// totalBorrowedPrefix | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixAdjustedTotalBorrow...)
	key = append(key, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateUTokenSupplyKey returns a KVStore key for getting and setting a utoken's total supply.
func CreateUTokenSupplyKey(uTokenDenom string) []byte {
	// supplyprefix | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixUtokenSupply...)
	key = append(key, []byte(uTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// AddressFromKey extracts address from a key with the form
// prefix | lengthPrefixed(addr) | ...
func AddressFromKey(key, prefix []byte) sdk.AccAddress {
	addrLength := int(key[len(prefix)])
	return key[len(prefix)+1 : len(prefix)+1+addrLength]
}

// DenomFromKeyWithAddress extracts denom from a key with the form
// prefix | lengthPrefixed(addr) | denom | 0x00
func DenomFromKeyWithAddress(key, prefix []byte) string {
	addrLength := int(key[len(prefix)])
	return string(key[len(prefix)+addrLength+1 : len(key)-1])
}

// DenomFromKey extracts denom from a key with the form
// prefix | denom | 0x00
func DenomFromKey(key, prefix []byte) string {
	return string(key[len(prefix) : len(key)-1])
}
