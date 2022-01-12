package util

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"fmt"
	"io"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// GenerateRandomEthKey generates a random Ethereum keypair.
func GenerateRandomEthKey() (*ecdsa.PrivateKey, *ecdsa.PublicKey, ethcmn.Address, error) {
	return GenerateRandomEthKeyFromRand(crand.Reader)
}

// GenerateRandomEthKeyFromRand generates a random Ethereum keypair from a
// reader.
func GenerateRandomEthKeyFromRand(r io.Reader) (*ecdsa.PrivateKey, *ecdsa.PublicKey, ethcmn.Address, error) {
	privKey, err := ecdsa.GenerateKey(ethcrypto.S256(), r)
	if err != nil {
		return nil, nil, ethcmn.Address{}, err
	}

	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		err := fmt.Errorf("unexpected public key type; expected: %T, got: %T", &ecdsa.PublicKey{}, publicKey)
		return nil, nil, ethcmn.Address{}, err
	}

	return privKey, publicKeyECDSA, ethcrypto.PubkeyToAddress(*publicKeyECDSA), nil
}
