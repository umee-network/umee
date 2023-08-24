package util

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

type gaiaValidator struct {
	index    int
	Mnemonic string
	keyInfo  keyring.Record
}

func (g *gaiaValidator) InstanceName() string {
	return fmt.Sprintf("gaiaval%d", g.index)
}
