package types

import (
	"crypto/ecdsa"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	signaturePrefix = "\x19Ethereum Signed Message:\n32"
)

// NewEthereumSignature creates a new signuature over a given byte array
func NewEthereumSignature(hash common.Hash, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, sdkerrors.Wrap(ErrEmpty, "private key")
	}
	protectedHash := crypto.Keccak256Hash(append([]uint8(signaturePrefix), hash[:]...))
	return crypto.Sign(protectedHash.Bytes(), privateKey)
}

// ValidateEthereumSignature takes a message, an associated signature and public key and
// returns an error if the signature isn't valid
func EthAddressFromSignature(hash common.Hash, signature []byte) (common.Address, error) {
	if len(signature) < 65 {
		return common.Address{}, sdkerrors.Wrap(ErrInvalid, "signature too short")
	}
	// To verify signature
	// - use crypto.SigToPub to get the public key
	// - use crypto.PubkeyToAddress to get the address
	// - compare this to the address given.

	// for backwards compatibility reasons  the V value of an Ethereum sig is presented
	// as 27 or 28, internally though it should be a 0-3 value due to changed formats.
	// It seems that go-ethereum expects this to be done before sigs actually reach it's
	// internal validation functions. In order to comply with this requirement we check
	// the sig an dif it's in standard format we correct it. If it's in go-ethereum's expected
	// format already we make no changes.
	//
	// We could attempt to break or otherwise exit early on obviously invalid values for this
	// byte, but that's a task best left to go-ethereum
	if signature[64] == 27 || signature[64] == 28 {
		signature[64] -= 27
	}

	protectedHash := crypto.Keccak256Hash(append([]byte(signaturePrefix), hash[:]...))

	pubkey, err := crypto.SigToPub(protectedHash.Bytes(), signature)
	if err != nil {
		return common.Address{}, sdkerrors.Wrap(err, "signature to public key")
	}

	addr := crypto.PubkeyToAddress(*pubkey)

	return addr, nil
}

// ValidateEthereumSignature takes a message, an associated signature and public key and
// returns an error if the signature isn't valid
func ValidateEthereumSignature(hash common.Hash, signature []byte, ethAddress common.Address) error {
	addr, err := EthAddressFromSignature(hash, signature)

	if err != nil {
		return sdkerrors.Wrap(err, "")
	}

	if addr != ethAddress {
		return sdkerrors.Wrap(ErrInvalid, "signature not matching")
	}

	return nil
}
