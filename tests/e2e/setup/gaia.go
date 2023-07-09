package setup

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

type gaiaValidator struct {
	index    int
	mnemonic string
	keyInfo  keyring.Record
}

func (g *gaiaValidator) instanceName() string {
	return fmt.Sprintf("gaiaval%d", g.index)
}

func (v *gaiaValidator) Address() (string, error) {
	// create master key and derive first key for keyring
	derivedPriv, err := hd.Secp256k1.Derive()(v.mnemonic, "", sdk.FullFundraiserPath)
	if err != nil {
		return "", err
	}

	privKey := hd.Secp256k1.Generate()(derivedPriv)
	bech32Addr, err := bech32.ConvertAndEncode("cosmos", privKey.PubKey().Address())
	if err != nil {
		panic(err)
	}

	return bech32Addr, nil
}
