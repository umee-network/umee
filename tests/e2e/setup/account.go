package setup

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type account struct {
	chain *chain
	name  string
	// fields below set during createKey
	rootDir    string
	mnemonic   string
	KeyInfo    keyring.Record
	privateKey cryptotypes.PrivKey
}

func (a *account) createKey(cdc codec.Codec, rootDir string) error {
	mnemonic, err := createMnemonic()
	if err != nil {
		return err
	}

	kb, err := keyring.New(keyringAppName, keyring.BackendTest, a.rootDir, nil, cdc)
	if err != nil {
		return err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return err
	}

	info, err := kb.NewAccount(a.name, mnemonic, "", sdk.FullFundraiserPath, algo)
	if err != nil {
		return err
	}

	privKeyArmor, err := kb.ExportPrivKeyArmor(a.name, keyringPassphrase)
	if err != nil {
		return err
	}

	privKey, _, err := sdkcrypto.UnarmorDecryptPrivKey(privKeyArmor, keyringPassphrase)
	if err != nil {
		return err
	}

	a.KeyInfo = *info
	a.mnemonic = mnemonic
	a.privateKey = privKey
	return nil
}
