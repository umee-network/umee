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
	TestAccounts   []*testAccount
	GaiaValidators []*gaiaValidator
}

// newChain creates a chain with a random chain ID and makes a temporary directory to hold its data
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

// configDir is located at <chain.dataDir>/<chain.ID>
func (c *chain) configDir() string {
	return fmt.Sprintf("%s/%s", c.dataDir, c.ID)
}

// createAndInitValidators adds a number of validators to the chain and initializes their
// keys, config.toml and genesis.json in separate config directories created for each validator.
// Their monikers are all set to "umee" but their indexes are different.
func (c *chain) createAndInitValidators(cdc codec.Codec, count int) error {
	for i := 0; i < count; i++ {
		node := validator{
			chain:   c,
			index:   i,
			moniker: "umee",
		}

		// create config directory and initializes config.toml and genesis.json
		if err := node.init(cdc); err != nil {
			return err
		}

		// generate a random key and save to keyring-backend-test in the validator's config directory
		if err := node.createKey(cdc, "val"); err != nil {
			return err
		}
		// loads or generates a node key in the validator's config directory
		// TODO (comment): which are we doing? loading or generating?
		if err := node.createNodeKey(); err != nil {
			return err
		}
		// loads or generates a consensus key in the validator's config directory
		// TODO (comment): which are we doing? loading or generating?
		if err := node.createConsensusKey(); err != nil {
			return err
		}

		// create a client which contains only this validator's keys
		var err error
		node.Client, err = c.initDedicatedClient(fmt.Sprint("val", i), node.mnemonic)
		if err != nil {
			return err
		}

		c.Validators = append(c.Validators, &node)
	}

	return nil
}
