package e2e

import (
	"fmt"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gravitytypes "github.com/peggyjv/gravity-bridge/module/x/gravity/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"

	"github.com/umee-network/umee/app"
	"github.com/umee-network/umee/app/params"
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
	encodingConfig = app.MakeEncodingConfig()

	encodingConfig.InterfaceRegistry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&stakingtypes.MsgCreateValidator{},
		&gravitytypes.MsgDelegateKeys{},
	)
	encodingConfig.InterfaceRegistry.RegisterImplementations(
		(*cryptotypes.PubKey)(nil),
		&secp256k1.PubKey{},
		&ed25519.PubKey{},
	)

	cdc = encodingConfig.Marshaler
}

type chain struct {
	dataDir        string
	id             string
	validators     []*validator
	orchestrators  []*orchestrator
	gaiaValidators []*gaiaValidator
}

func newChain() (*chain, error) {
	tmpDir, err := ioutil.TempDir("", "umee-e2e-testnet")
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

func (c *chain) createAndInitValidatorsWithMnemonics(count int, mnemonics []string) error {
	for i := 0; i < count; i++ {
		// create node
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(); err != nil {
			return err
		}

		c.validators = append(c.validators, node)

		// create keys
		if err := node.createKeyFromMnemonic("val", mnemonics[i]); err != nil {
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

		// create keys
		mnemonic, info, err := createMemoryKey()
		if err != nil {
			return err
		}

		orchestrator.keyInfo = *info
		orchestrator.mnemonic = mnemonic

		c.orchestrators = append(c.orchestrators, orchestrator)
	}

	return nil
}

func (c *chain) createAndInitOrchestratorsWithMnemonics(count int, mnemonics []string) error {
	for i := 0; i < count; i++ {
		// create orchestrator
		orchestrator := c.createOrchestrator(i)

		// create keys
		info, err := createMemoryKeyFromMnemonic(mnemonics[i])
		if err != nil {
			return err
		}

		orchestrator.keyInfo = *info
		orchestrator.mnemonic = mnemonics[i]

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
		chain: c,
		index: index,
	}
}
