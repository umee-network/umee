package setup

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
)

// createMnemonic generates a random mnemonic to be used in key generation
func createMnemonic() (string, error) {
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropySeed)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
}

// createMemoryKey generates a random key which will be stored only in memory
func createMemoryKey(cdc codec.Codec) (mnemonic string, info *keyring.Record, err error) {
	mnemonic, err = createMnemonic()
	if err != nil {
		return "", nil, err
	}

	kb, err := keyring.New("testnet", keyring.BackendMemory, "", nil, cdc)
	if err != nil {
		return "", nil, err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return "", nil, err
	}

	account, err := kb.NewAccount("", mnemonic, "", sdk.FullFundraiserPath, algo)
	if err != nil {
		return "", nil, err
	}

	return mnemonic, account, nil
}
