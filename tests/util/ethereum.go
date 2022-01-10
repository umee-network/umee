package util

import (
	"crypto/ecdsa"
	"fmt"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// GenerateRandomEthKey generates a random Ethereum keypair.
func GenerateRandomEthKey() (*ecdsa.PrivateKey, *ecdsa.PublicKey, ethcmn.Address, error) {
	privKey, err := ethcrypto.GenerateKey()
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
