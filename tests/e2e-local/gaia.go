package e2e

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

type gaiaValidator struct {
	index    int
	mnemonic string
	keyInfo  keyring.Record
}

func (g *gaiaValidator) instanceName() string {
	return fmt.Sprintf("gaiaval%d", g.index)
}
