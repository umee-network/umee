package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/spf13/viper"
	tmconfig "github.com/tendermint/tendermint/config"

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
	KeyringPassphrase = "testpassphrase"
	KeyringAppName    = "testnet"
)

var EncodingConfig sdkparams.EncodingConfig

func init() {
	EncodingConfig = umeeapp.MakeEncodingConfig()

	EncodingConfig.InterfaceRegistry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&stakingtypes.MsgCreateValidator{},
	)
	EncodingConfig.InterfaceRegistry.RegisterImplementations(
		(*cryptotypes.PubKey)(nil),
		&secp256k1.PubKey{},
		&ed25519.PubKey{},
	)
}

type Chain struct {
	DataDir        string
	ID             string
	Validators     []*validator
	GaiaValidators []*gaiaValidator
}

func NewChain() (*Chain, error) {
	tmpDir, err := os.MkdirTemp("", "umee-e2e-testnet-")
	if err != nil {
		return nil, err
	}

	return &Chain{
		ID:      "Chain-" + tmrand.NewRand().Str(6),
		DataDir: tmpDir,
	}, nil
}

func (c *Chain) configDir() string {
	return fmt.Sprintf("%s/%s", c.DataDir, c.ID)
}

func (c *Chain) InitNodes(cdc codec.Codec, numValidators int) error {
	if err := c.CreateAndInitValidators(cdc, numValidators); err != nil {
		return err
	}

	// initialize a genesis file for the first validator
	val0ConfigDir := c.Validators[0].ConfigDir()
	for _, val := range c.Validators {
		valAddr, err := val.KeyInfo.GetAddress()
		if err != nil {
			return err
		}
		if err := AddGenesisAccount(cdc, val0ConfigDir, "", InitBalanceStr, valAddr); err != nil {
			return err
		}

	}

	// copy the genesis file to the remaining validators
	for _, val := range c.Validators[1:] {
		if _, err := CopyFile(
			filepath.Join(val0ConfigDir, "config", "genesis.json"),
			filepath.Join(val.ConfigDir(), "config", "genesis.json"),
		); err != nil {
			return err
		}
	}

	return nil
}

func (c *Chain) InitValidatorConfigs() error {
	for i, val := range c.Validators {
		tmCfgPath := filepath.Join(val.ConfigDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(tmCfgPath)
		if err := vpr.ReadInConfig(); err != nil {
			return err
		}

		valConfig := tmconfig.DefaultConfig()
		valConfig.Consensus.SkipTimeoutCommit = true
		if err := vpr.Unmarshal(valConfig); err != nil {
			return err
		}

		valConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
		valConfig.P2P.AddrBookStrict = false
		valConfig.P2P.ExternalAddress = fmt.Sprintf("%s:%d", val.InstanceName(), 26656)
		valConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		valConfig.StateSync.Enable = false
		valConfig.LogLevel = "info"

		var peers []string

		for j := 0; j < len(c.Validators); j++ {
			if i == j {
				continue
			}

			peer := c.Validators[j]
			peerID := fmt.Sprintf("%s@%s%d:26656", peer.NodeKey.ID(), peer.Moniker, j)
			peers = append(peers, peerID)
		}

		valConfig.P2P.PersistentPeers = strings.Join(peers, ",")

		tmconfig.WriteConfigFile(tmCfgPath, valConfig)

		// set application configuration
		appCfgPath := filepath.Join(val.ConfigDir(), "config", "app.toml")

		appConfig := srvconfig.DefaultConfig()
		appConfig.API.Enable = true
		appConfig.MinGasPrices = MinGasPrice
		appConfig.Pruning = "nothing"

		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}

	return nil
}

func (c *Chain) CreateAndInitValidators(cdc codec.Codec, count int) error {
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

func (c *Chain) CreateAndInitGaiaValidator(cdc codec.Codec) error {
	// create gaia validator
	gaiaVal := c.createGaiaValidator(0)

	// create keys
	mnemonic, info, err := createMemoryKey(cdc)
	if err != nil {
		return err
	}

	gaiaVal.keyInfo = *info
	gaiaVal.Mnemonic = mnemonic

	c.GaiaValidators = append(c.GaiaValidators, gaiaVal)

	return nil
}

func (c *Chain) createValidator(index int) *validator {
	return &validator{
		chain:   c,
		Index:   index,
		Moniker: "umee",
	}
}

func (c *Chain) createGaiaValidator(index int) *gaiaValidator {
	return &gaiaValidator{
		index: index,
	}
}
