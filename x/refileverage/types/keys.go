package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/umee-network/umee/v5/util"
)

const (
	// ModuleName defines the module name
	ModuleName = "refileverage"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
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

// KeyRegisteredToken returns a KVStore key for getting and setting a Token.
func KeyRegisteredToken(baseTokenDenom string) []byte {
	// assetprefix | denom | 0x00 for null-termination
	return util.ConcatBytes(1, KeyPrefixRegisteredToken, []byte(baseTokenDenom))
}

// KeyAdjustedBorrow returns a KVStore key for getting and setting an
// adjusted borrow for a denom and borrower address.
func KeyAdjustedBorrow(borrowerAddr sdk.AccAddress) []byte {
	// borrowprefix | lengthprefixed(borrowerAddr)
	return util.ConcatBytes(0, KeyPrefixAdjustedBorrow, address.MustLengthPrefix(borrowerAddr))
}

// KeyCollateralAmount returns a KVStore key for getting and setting the amount of
// collateral stored for a user in a given denom.
func KeyCollateralAmount(addr sdk.AccAddress, uTokenDenom string) []byte {
	// collateralPrefix | lengthprefixed(addr) | denom | 0x00 for null-termination
	return util.ConcatBytes(1, KeyCollateralAmountNoDenom(addr), []byte(uTokenDenom))
}

// KeyCollateralAmountNoDenom returns the common prefix used by all collateral associated
// with a given address.
func KeyCollateralAmountNoDenom(addr sdk.AccAddress) []byte {
	// collateralPrefix | lengthprefixed(addr)
	return util.ConcatBytes(0, KeyPrefixCollateralAmount, address.MustLengthPrefix(addr))
}

// KeyReserveAmount returns a KVStore key for getting and setting the amount reserved of a a given token.
func KeyReserveAmount(tokenDenom string) []byte {
	// reserveamountprefix | denom | 0x00 for null-termination
	return util.ConcatBytes(1, KeyPrefixReserveAmount, []byte(tokenDenom))
}

// KeyBadDebt returns a KVStore key for tracking an address with unpaid bad debt
func KeyBadDebt(denom string, borrower sdk.AccAddress) []byte {
	// badDebtAddrPrefix | lengthprefixed(borrowerAddr) | denom | 0x00 for null-termination
	return util.ConcatBytes(1, KeyPrefixBadDebt, address.MustLengthPrefix(borrower), []byte(denom))
}

// KeyInterestScalar returns a KVStore key for getting and setting the interest scalar for a
// given token.
func KeyInterestScalar(tokenDenom string) []byte {
	// interestScalarPrefix | denom | 0x00 for null-termination
	return util.ConcatBytes(1, KeyPrefixInterestScalar, []byte(tokenDenom))
}

// KeyAdjustedTotalBorrow returns a KVStore key for getting and setting the total ajdusted borrows for
// a given token.
func KeyAdjustedTotalBorrow() []byte {
	// totalBorrowedPrefix | denom | 0x00 for null-termination
	return util.ConcatBytes(1, KeyPrefixAdjustedTotalBorrow, []byte(Gho))
}

// KeyUTokenSupply returns a KVStore key for getting and setting a utoken's total supply.
func KeyUTokenSupply(uTokenDenom string) []byte {
	// supplyprefix | denom | 0x00 for null-termination
	return util.ConcatBytes(1, KeyPrefixUtokenSupply, []byte(uTokenDenom))
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
