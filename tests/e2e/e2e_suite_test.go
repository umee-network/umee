package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	tmconfig "github.com/tendermint/tendermint/config"
)

const (
	initBalanceStr      = "100000000000uumee,100000000000wdai"
	ethChainID     uint = 15
)

var (
	stakeAmount, _  = sdk.NewIntFromString("100000000000")
	stakeAmountCoin = sdk.NewCoin("uumee", stakeAmount)
)

type IntegrationTestSuite struct {
	suite.Suite

	chain *chain
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	s.T().Log("setting up e2e integration test suite...")

	var err error
	s.chain, err = newChain()
	s.Require().NoError(err)

	s.T().Logf("starting e2e infrastructure; chain-id: %s; datadir: %s", s.chain.id, s.chain.dataDir)

	s.initNodes()
	s.initEthereum()
	s.initGenesis()
	s.initValidatorConfigs()
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e integration test suite...")
	os.RemoveAll(s.chain.dataDir)
}

func (s *IntegrationTestSuite) initNodes() {
	s.Require().NoError(s.chain.createAndInitValidators(4))
	s.Require().NoError(s.chain.createAndInitOrchestrators(4))

	// initialize a genesis file for the first validator
	val0ConfigDir := s.chain.validators[0].configDir()
	for _, val := range s.chain.validators {
		s.Require().NoError(
			addGenesisAccount(val0ConfigDir, "", initBalanceStr, val.keyInfo.GetAddress()),
		)
	}

	// add orchestrator accounts to genesis file
	for _, orch := range s.chain.orchestrators {
		s.Require().NoError(
			addGenesisAccount(val0ConfigDir, "", initBalanceStr, orch.keyInfo.GetAddress()),
		)
	}

	// copy the genesis file to the remaining validators
	for _, val := range s.chain.validators[1:] {
		_, err := fileCopy(
			filepath.Join(val0ConfigDir, "config", "genesis.json"),
			filepath.Join(val.configDir(), "config", "genesis.json"),
		)
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) initEthereum() {
	// generate ethereum keys for validators add them to the ethereum genesis
	ethGenesis := EthereumGenesis{
		Difficulty: "0x400",
		GasLimit:   "0xB71B00",
		Config:     EthereumConfig{ChainID: ethChainID},
		Alloc:      make(map[string]Allocation, len(s.chain.validators)+1),
	}

	alloc := Allocation{
		Balance: "0x1337000000000000000000",
	}

	ethGenesis.Alloc["0xBf660843528035a5A4921534E156a27e64B231fE"] = alloc
	for _, val := range s.chain.validators {
		s.Require().NoError(val.generateEthereumKey())
		ethGenesis.Alloc[val.ethereumKey.address] = alloc
	}

	ethGenBz, err := json.MarshalIndent(ethGenesis, "", "  ")
	s.Require().NoError(err)

	// write out the genesis file
	s.Require().NoError(writeFile(filepath.Join(s.chain.configDir(), "eth_genesis.json"), ethGenBz))
}

func (s *IntegrationTestSuite) initGenesis() {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(s.chain.validators[0].configDir())
	config.Moniker = s.chain.validators[0].moniker

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	s.Require().NoError(err)

	var bankGenState banktypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState))

	bankGenState.DenomMetadata = append(bankGenState.DenomMetadata, banktypes.Metadata{
		Description: "The native staking token of the Umee network",
		Display:     "umee",
		Base:        "uumee",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "uumee",
				Exponent: 0,
				Aliases: []string{
					"microumee",
				},
			},
			{
				Denom:    "umee",
				Exponent: 6,
			},
		},
	})
	bankGenState.DenomMetadata = append(bankGenState.DenomMetadata, banktypes.Metadata{
		Description: "An example stable token",
		Display:     "wdai",
		Base:        "wdai",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "wdai",
				Exponent: 0,
			},
		},
	})

	bz, err := cdc.MarshalJSON(&bankGenState)
	s.Require().NoError(err)
	appGenState[banktypes.ModuleName] = bz

	var genUtilGenState genutiltypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState))

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(s.chain.validators))
	for i, val := range s.chain.validators {
		createValmsg, err := val.buildCreateValidatorMsg(stakeAmountCoin)
		s.Require().NoError(err)

		delKeysMsg := val.buildDelegateKeysMsg()
		s.Require().NoError(err)

		signedTx, err := val.signMsg(createValmsg, delKeysMsg)
		s.Require().NoError(err)

		txRaw, err := cdc.MarshalJSON(signedTx)
		s.Require().NoError(err)

		genTxs[i] = txRaw
	}

	genUtilGenState.GenTxs = genTxs

	bz, err = cdc.MarshalJSON(&genUtilGenState)
	s.Require().NoError(err)
	appGenState[genutiltypes.ModuleName] = bz

	bz, err = json.MarshalIndent(appGenState, "", "  ")
	s.Require().NoError(err)

	genDoc.AppState = bz

	bz, err = json.MarshalIndent(genDoc, "", "  ")
	s.Require().NoError(err)

	// write the updated genesis file to each validator
	for _, val := range s.chain.validators {
		writeFile(filepath.Join(val.configDir(), "config", "genesis.json"), bz)
	}
}

func (s *IntegrationTestSuite) initValidatorConfigs() {
	for i, val := range s.chain.validators {
		cfgPath := filepath.Join(val.configDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(cfgPath)
		s.Require().NoError(vpr.ReadInConfig())

		valConfig := &tmconfig.Config{}
		s.Require().NoError(vpr.Unmarshal(valConfig))

		valConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
		valConfig.P2P.AddrBookStrict = false
		valConfig.P2P.ExternalAddress = fmt.Sprintf("%s:%d", val.instanceName(), 26656)
		valConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		valConfig.StateSync.Enable = false
		valConfig.LogLevel = "info"

		var peers []string

		for j := 0; j < len(s.chain.validators); j++ {
			if i == j {
				continue
			}

			peer := s.chain.validators[j]
			peerID := fmt.Sprintf("%s@%s%d:26656", peer.nodeKey.ID(), peer.moniker, j)
			peers = append(peers, peerID)
		}

		valConfig.P2P.PersistentPeers = strings.Join(peers, ",")

		tmconfig.WriteConfigFile(cfgPath, valConfig)
	}
}
