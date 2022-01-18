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
	KeyPrefixRegisteredToken   = []byte{0x01}
	KeyPrefixLoanToken         = []byte{0x02}
	KeyPrefixCollateralSetting = []byte{0x03}
	KeyPrefixCollateralAmount  = []byte{0x04}
	KeyPrefixReserveAmount     = []byte{0x05}
	KeyPrefixLastInterestTime  = []byte{0x06}
	KeyPrefixExchangeRate      = []byte{0x07}
	KeyPrefixBadDebt           = []byte{0x08}
	KeyPrefixBorrowAPY         = []byte{0x09}
	KeyPrefixLendAPY           = []byte{0x0A}
)

// CreateRegisteredTokenKey returns a KVStore key for getting and setting a Token.
func CreateRegisteredTokenKey(baseTokenDenom string) []byte {
	// assetprefix | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixRegisteredToken...)
	key = append(key, []byte(baseTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateLoanKey returns a KVStore key for getting and setting a Loan for a denom
// and borrower address.
func CreateLoanKey(borrowerAddr sdk.AccAddress, tokenDenom string) []byte {
	// loanprefix | lengthprefixed(borrowerAddr) | denom | 0x00
	key := CreateLoanKeyNoDenom(borrowerAddr)
	key = append(key, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateLoanKeyNoDenom returns the common prefix used by all loans associated with a given
// borrower address.
func CreateLoanKeyNoDenom(borrowerAddr sdk.AccAddress) []byte {
	// loanprefix | lengthprefixed(borrowerAddr)
	key := CreateLoanKeyNoAddress()
	key = append(key, address.MustLengthPrefix(borrowerAddr)...)
	return key
}

// CreateLoanKeyNoAddress returns a safe copy of loanprefix
func CreateLoanKeyNoAddress() []byte {
	// loanprefix
	var key []byte
	key = append(key, KeyPrefixLoanToken...)
	return key
}

// CreateCollateralSettingKey returns a KVStore key for getting and setting a borrower's
// collateral setting for a single uToken
func CreateCollateralSettingKey(borrowerAddr sdk.AccAddress, uTokenDenom string) []byte {
	// collatprefix | lengthprefixed(borrowerAddr) | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixCollateralSetting...)
	key = append(key, address.MustLengthPrefix(borrowerAddr)...)
	key = append(key, []byte(uTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
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

// CreateExchangeRateKey returns a KVStore key for getting and setting the token:uToken rate for a a given token.
func CreateExchangeRateKey(tokenDenom string) []byte {
	// exchangeRatePrefix | denom | 0x00
	key := CreateExchangeRateKeyNoDenom()
	key = append(key, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateExchangeRateKeyNoDenom returns a safe copy of exchangeRatePrefix
func CreateExchangeRateKeyNoDenom() []byte {
	// exchangeRatePrefix
	var key []byte
	key = append(key, KeyPrefixExchangeRate...)
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

// CreateBorrowAPYKey returns a KVStore key for getting and setting the borrow APY for a given token.
func CreateBorrowAPYKey(tokenDenom string) []byte {
	// borrowAPYPrefix | denom | 0x00
	key := CreateBorrowAPYKeyNoDenom()
	key = append(key, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateBorrowAPYKeyNoDenom returns a safe copy of borrow APY prefix
func CreateBorrowAPYKeyNoDenom() []byte {
	// borrowAPYPrefix
	var key []byte
	key = append(key, KeyPrefixBorrowAPY...)
	return key
}

// CreateLendAPYKey returns a KVStore key for getting and setting the lend APY for a given token.
func CreateLendAPYKey(tokenDenom string) []byte {
	// lendAPYPrefix | denom | 0x00
	key := CreateLendAPYKeyNoDenom()
	key = append(key, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateLendAPYKeyNoDenom returns a safe copy of lend APY prefix
func CreateLendAPYKeyNoDenom() []byte {
	// lendAPYPrefix
	var key []byte
	key = append(key, KeyPrefixLendAPY...)
	return key
}

// AddressFromKey extracts address from a key with the form
// prefix | lengthPrefixed(addr) | ...
func AddressFromKey(key []byte, prefix []byte) sdk.AccAddress {
	addrLength := int(key[len(prefix)])
	return key[len(prefix)+1 : len(prefix)+1+addrLength]
}

// DenomFromKeyWithAddress extracts denom from a key with the form
// prefix | lengthPrefixed(addr) | denom | 0x00
func DenomFromKeyWithAddress(key []byte, prefix []byte) string {
	addrLength := int(key[len(prefix)])
	return string(key[len(prefix)+addrLength+1 : len(key)-1])
}

// DenomFromKey extracts denom from a key with the form
// prefix | denom | 0x00
func DenomFromKey(key []byte, prefix []byte) string {
	return string(key[len(prefix) : len(key)-1])
}
