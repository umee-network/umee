package simulation

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type EthAddressGenerator struct {
	fauxRandomizer int64
}

func NewEthAddressGenerator(fauxRandomizer int64) EthAddressGenerator {
	return EthAddressGenerator{
		fauxRandomizer: fauxRandomizer,
	}
}

func (e EthAddressGenerator) generateKey() *ecdsa.PrivateKey {
	c := secp256k1.S256()
	k := big.NewInt(e.fauxRandomizer)
	e.fauxRandomizer++
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = c
	priv.D = k
	priv.PublicKey.X, priv.PublicKey.Y = c.ScalarBaseMult(k.Bytes())
	return priv
}

// generateEthAddress generates a random valid eth address
func (e EthAddressGenerator) GenerateEthAddress() string {
	privateKey := e.generateKey()
	publicKey := privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return address
}
