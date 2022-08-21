package e2e

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/app/params"
)

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"
)

var (
	encodingConfig params.EncodingConfig
	cdc            codec.Codec
)

func init() {
	encodingConfig = umeeapp.MakeEncodingConfig()

	encodingConfig.InterfaceRegistry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&stakingtypes.MsgCreateValidator{},
	)
	encodingConfig.InterfaceRegistry.RegisterImplementations(
		(*cryptotypes.PubKey)(nil),
		&secp256k1.PubKey{},
		&ed25519.PubKey{},
	)

	cdc = encodingConfig.Codec
}

type chain struct {
	dataDir        string
	id             string
	validators     []*validator
	orchestrators  []*orchestrator
	gaiaValidators []*gaiaValidator
}

func newChain() (*chain, error) {
	tmpDir, err := os.MkdirTemp("", "umee-e2e-testnet-")
	if err != nil {
		return nil, err
	}

	return &chain{
		id:      "chain-" + tmrand.NewRand().Str(6),
		dataDir: tmpDir,
	}, nil
}

func (c *chain) configDir() string {
	return fmt.Sprintf("%s/%s", c.dataDir, c.id)
}

func (c *chain) createAndInitValidators(count int) error {
	for i := 0; i < count; i++ {
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(); err != nil {
			return err
		}

		c.validators = append(c.validators, node)

		// create keys
		if err := node.createKey("val"); err != nil {
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

func (c *chain) createAndInitGaiaValidator() error {
	// create gaia validator
	gaiaVal := c.createGaiaValidator(0)

	// create keys
	mnemonic, info, err := createMemoryKey()
	if err != nil {
		return err
	}

	gaiaVal.keyInfo = *info
	gaiaVal.mnemonic = mnemonic

	c.gaiaValidators = append(c.gaiaValidators, gaiaVal)

	return nil
}

func (c *chain) createAndInitOrchestrators(count int) error {
	for i := 0; i < count; i++ {
		// create orchestrator
		orchestrator := c.createOrchestrator(i)

		err := orchestrator.createKey("orch")
		if err != nil {
			return err
		}

		err = orchestrator.generateEthereumKey()
		if err != nil {
			return err
		}

		c.orchestrators = append(c.orchestrators, orchestrator)
	}

	return nil
}

func (c *chain) createValidator(index int) *validator {
	return &validator{
		chain:   c,
		index:   index,
		moniker: "umee",
	}
}

func (c *chain) createOrchestrator(index int) *orchestrator {
	return &orchestrator{
		index: index,
	}
}

func (c *chain) createGaiaValidator(index int) *gaiaValidator {
	return &gaiaValidator{
		index: index,
	}
}
