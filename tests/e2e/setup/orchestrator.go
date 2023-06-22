package setup

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type orchestrator struct {
	Index       int
	Mnemonic    string
	keyInfo     keyring.Record
	PrivateKey  cryptotypes.PrivKey
	EthereumKey ethereumKey
}

type ethereumKey struct {
	PublicKey  string
	PrivateKey string
	Address    string
}

func (o *orchestrator) instanceName() string {
	return fmt.Sprintf("orchestrator%d", o.Index)
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

	o.EthereumKey = ethereumKey{
		PrivateKey: hexutil.Encode(privateKeyBytes),
		PublicKey:  hexutil.Encode(publicKeyBytes),
		Address:    crypto.PubkeyToAddress(*publicKeyECDSA).Hex(),
	}

	return nil
}

func (o *orchestrator) createKey(cdc codec.Codec, name string) error {
	mnemonic, err := createMnemonic()
	if err != nil {
		return err
	}

	return o.createKeyFromMnemonic(cdc, name, mnemonic)
}

func (o *orchestrator) createKeyFromMnemonic(cdc codec.Codec, name, mnemonic string) error {
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
	o.Mnemonic = mnemonic
	o.PrivateKey = privKey

	return nil
}
