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
	KeyPrefixBadDebtDenom      = []byte{0x08}
	KeyPrefixBadDebtAddress    = []byte{0x09}
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
	var key []byte
	key = append(key, KeyPrefixLoanToken...)
	key = append(key, address.MustLengthPrefix(borrowerAddr)...)
	key = append(key, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateLoanKeyNoDenom returns the common prefix used by all loans associated with a given
// borrower address.
func CreateLoanKeyNoDenom(borrowerAddr sdk.AccAddress) []byte {
	// loanprefix | lengthprefixed(borrowerAddr)
	var key []byte
	key = append(key, KeyPrefixLoanToken...)
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
	// collateralprefix | lengthprefixed(lenderAddr) | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixCollateralAmount...)
	key = append(key, address.MustLengthPrefix(lenderAddr)...)
	key = append(key, []byte(uTokenDenom)...)
	return append(key, 0) // append 0 for null-termination
}

// CreateCollateralAmountKeyNoDenom returns the common prefix used by all collateral associated
// with a given lender address.
func CreateCollateralAmountKeyNoDenom(lenderAddr sdk.AccAddress) []byte {
	// collateralprefix | lengthprefixed(lenderAddr)
	var key []byte
	key = append(key, KeyPrefixCollateralAmount...)
	key = append(key, address.MustLengthPrefix(lenderAddr)...)
	return key
}

// CreateReserveAmountKey returns a KVStore key for getting and setting the amount reserved of a a given token.
func CreateReserveAmountKey(tokenDenom string) []byte {
	// reserveamountprefix | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixReserveAmount...)
	key = append(key, []byte(tokenDenom)...)
	return append(key, 0) // append 0 for null-termination
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
	var key []byte
	key = append(key, KeyPrefixExchangeRate...)
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

// CreateBadDebtDenomKey returns a KVStore key for tracking a denom with unpaid bad debt
func CreateBadDebtDenomKey(denom string) []byte {
	// badDebtDenomPrefix | denom | 0x00
	var key []byte
	key = append(key, KeyPrefixBadDebtDenom...)
	key = append(key, []byte(denom)...)
	return key
}

// CreateBadDebtDenomKeyNoDenom returns a safe copy of badDebtDenomPrefix
func CreateBadDebtDenomKeyNoDenom() []byte {
	// badDebtDenomPrefix
	var key []byte
	key = append(key, KeyPrefixBadDebtDenom...)
	return key
}

// CreateBadDebtAddressKey returns a KVStore key for tracking a denom with unpaid bad debt
func CreateBadDebtAddressKey(denom string, borrowerAddr sdk.AccAddress) []byte {
	// badDebtAddrPrefix | lengthprefixed(denom) | borrowerAddr
	var key []byte
	key = append(key, KeyPrefixBadDebtAddress...)
	key = append(key, LengthPrefixDenom(denom)...)
	key = append(key, borrowerAddr...)
	return key
}

// CreateBadDebtAddressKeyNoAddress returns a KVStore key for iterating over a debom's bad debt
func CreateBadDebtAddressKeyNoAddress(denom string) []byte {
	// badDebtAddrPrefix | lengthprefixed(denom)
	var key []byte
	key = append(key, KeyPrefixBadDebtAddress...)
	key = append(key, LengthPrefixDenom(denom)...)
	return key
}

// LengthPrefixDenom matches the functionality of address.MustLengthPrefix,
// but we are using it for denoms here, and also forbid len=0
func LengthPrefixDenom(denom string) []byte {
	bz := []byte(denom)
	bzLen := len(bz)
	// Note: Valid denoms in cosmos are 3-128 characters
	if bzLen == 0 {
		panic("length cannot be 0 bytes")
	}
	if bzLen > 255 {
		panic("length cannot exceed 255 bytes")
	}
	return append([]byte{byte(bzLen)}, bz...)
}
