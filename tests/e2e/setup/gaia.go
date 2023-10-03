package setup

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
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

func (c *chain) createAndInitGaiaValidator(cdc codec.Codec) error {
	// create gaia validator
	gaiaVal := c.createGaiaValidator(0)

	// create keys
	mnemonic, info, err := createMemoryKey(cdc)
	if err != nil {
		return err
	}

	gaiaVal.keyInfo = *info
	gaiaVal.mnemonic = mnemonic

	c.GaiaValidators = append(c.GaiaValidators, gaiaVal)

	return nil
}

func (c *chain) createGaiaValidator(index int) *gaiaValidator {
	return &gaiaValidator{
		index: index,
	}
}
