package e2e

import (
	"crypto/ecdsa"
	"fmt"

	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type orchestrator struct {
	index       int
	mnemonic    string
	keyInfo     keyring.Record
	privateKey  cryptotypes.PrivKey
	ethereumKey ethereumKey
}

type ethereumKey struct {
	publicKey  string
	privateKey string
	address    string
}

func (o *orchestrator) instanceName() string {
	return fmt.Sprintf("orchestrator%d", o.index)
}

func (o *orchestrator) generateEthereumKey() error {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return err
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("unexpected public key type; expected: %T, got: %T", &ecdsa.PublicKey{}, publicKey)
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	o.ethereumKey = ethereumKey{
		privateKey: hexutil.Encode(privateKeyBytes),
		publicKey:  hexutil.Encode(publicKeyBytes),
		address:    crypto.PubkeyToAddress(*publicKeyECDSA).Hex(),
	}

	return nil
}

func (o *orchestrator) createKey(name string) error {
	mnemonic, err := createMnemonic()
	if err != nil {
		return err
	}

	return o.createKeyFromMnemonic(name, mnemonic)
}

func (o *orchestrator) createKeyFromMnemonic(name, mnemonic string) error {
	kb, err := keyring.New(keyringAppName, keyring.BackendMemory, "", nil, cdc)
	if err != nil {
		return err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return err
	}

	info, err := kb.NewAccount(name, mnemonic, "", sdk.FullFundraiserPath, algo)
	if err != nil {
		return err
	}

	privKeyArmor, err := kb.ExportPrivKeyArmor(name, keyringPassphrase)
	if err != nil {
		return err
	}

	privKey, _, err := sdkcrypto.UnarmorDecryptPrivKey(privKeyArmor, keyringPassphrase)
	if err != nil {
		return err
	}

	o.keyInfo = *info
	o.mnemonic = mnemonic
	o.privateKey = privKey

	return nil
}
