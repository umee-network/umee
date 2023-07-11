package setup

import (
	"fmt"
	"os"

	tmrand "github.com/cometbft/cometbft/libs/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
)

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"
)

var encodingConfig testutil.TestEncodingConfig

type chain struct {
	dataDir        string
	ID             string
	Validators     []*validator
	GaiaValidators []*gaiaValidator
}

func newChain() (*chain, error) {
	tmpDir, err := os.MkdirTemp("", "umee-e2e-testnet-")
	if err != nil {
		return nil, err
	}

	return &chain{
		ID:      "chain-" + tmrand.NewRand().Str(6),
		dataDir: tmpDir,
	}, nil
}

func (c *chain) configDir() string {
	return fmt.Sprintf("%s/%s", c.dataDir, c.ID)
}

func (c *chain) createAndInitValidators(cdc codec.Codec, count int) error {
	for i := 0; i < count; i++ {
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(cdc); err != nil {
			return err
		}

		c.Validators = append(c.Validators, node)

		// create keys
		if err := node.createKey(cdc, "val"); err != nil {
			return err
		}
		if err := node.createNodeKey(); err != nil {
			return err
		}
		if err := node.createConsensusKey(); err != nil {
			return err
		}
	}

	return nil
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

func (c *chain) createValidator(index int) *validator {
	return &validator{
		chain:   c,
		index:   index,
		moniker: "umee",
	}
}

func (c *chain) createGaiaValidator(index int) *gaiaValidator {
	return &gaiaValidator{
		index: index,
	}
}
