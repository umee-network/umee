package setup

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"

	umeeapp "github.com/umee-network/umee/v6/app"
)

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"
)

var encodingConfig sdkparams.EncodingConfig

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
}

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
		val := validator{
			chain:   c,
			index:   i,
			moniker: "umee",
		}

		// create config directory and initializes config.toml and genesis.json
		if err := val.init(cdc); err != nil {
			return err
		}

		// generate a random key and save to keyring-backend-test in the validator's config directory
		if err := val.createKey(cdc, "val"); err != nil {
			return err
		}
		// loads or generates a node key in the validator's config directory
		// TODO (comment): which are we doing? loading or generating?
		if err := val.createNodeKey(); err != nil {
			return err
		}
		// loads or generates a consensus key in the validator's config directory
		// TODO (comment): which are we doing? loading or generating?
		if err := val.createConsensusKey(); err != nil {
			return err
		}

		c.Validators = append(c.Validators, &val)
	}

	return nil
}
