package tx

import (
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"
)

func CreateAccountFromMnemonic(kb keyring.Keyring, name, mnemonic string) (*keyring.Record, error) {
	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return nil, err
	}

	account, err := kb.NewAccount(name, mnemonic, "", sdk.FullFundraiserPath, algo)
	if err != nil {
		return nil, err
	}

	return account, nil
}
