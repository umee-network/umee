package e2e

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

type gaiaValidator struct {
	chain    *chain
	index    int
	mnemonic string
	keyInfo  keyring.Info
}

func (g *gaiaValidator) instanceName() string {
	return fmt.Sprintf("gaiaval%d", g.index)
}

func (g *gaiaValidator) configDir() string {
	return fmt.Sprintf("%s/%s", g.chain.configDir(), g.instanceName())
}
