package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	gravitytypes "github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/types"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	tmconfig "github.com/tendermint/tendermint/config"
	tmjson "github.com/tendermint/tendermint/libs/json"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/umee-network/umee/v4/app"
	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/client"
	"github.com/umee-network/umee/v4/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
	"github.com/umee-network/umee/v4/x/uibc"
)

const (
	photonDenom    = "photon"
	initBalanceStr = "510000000000" + appparams.BondDenom + ",100000000000" + photonDenom
	gaiaChainID    = "test-gaia-chain"

	ethChainID uint = 15
	ethMinerPK      = "0xb1bab011e03a9862664706fc3bbaa1b16651528e5f0e7fbfcbfdd8be302a13e7"

	priceFeederContainerRepo  = "ghcr.io/umee-network/price-feeder-umee"
	priceFeederServerPort     = "7171/tcp"
	priceFeederMaxStartupTime = 20 // seconds
)

var (
	minGasPrice     = appparams.ProtocolMinGasPrice.String()
	stakeAmount, _  = sdk.NewIntFromString("100000000000")
	stakeAmountCoin = sdk.NewCoin(appparams.BondDenom, stakeAmount)

	stakeAmount2, _  = sdk.NewIntFromString("500000000000")
	stakeAmountCoin2 = sdk.NewCoin(appparams.BondDenom, stakeAmount2)
)

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs             []string
	chain               *chain
	ethClient           *ethclient.Client
	gaiaRPC             *rpchttp.HTTP
	dkrPool             *dockertest.Pool
	dkrNet              *dockertest.Network
	ethResource         *dockertest.Resource
	gaiaResource        *dockertest.Resource
	hermesResource      *dockertest.Resource
	priceFeederResource *dockertest.Resource
	valResources        []*dockertest.Resource
	orchResources       []*dockertest.Resource
	gravityContractAddr string
	umee                client.Client
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")

	var err error
	s.chain, err = newChain()
	s.Require().NoError(err)

	s.T().Logf("starting e2e infrastructure; chain-id: %s; datadir: %s", s.chain.id, s.chain.dataDir)

	s.dkrPool, err = dockertest.NewPool("")
	s.Require().NoError(err)

	s.dkrNet, err = s.dkrPool.CreateNetwork(fmt.Sprintf("%s-testnet", s.chain.id))
	s.Require().NoError(err)

	s.T().Logf("Ethereum and peggo are disable due to Ethereum PoS migration and PoW fork")
	// var useGanache bool
	// if str := os.Getenv("UMEE_E2E_USE_GANACHE"); len(str) > 0 {
	// 	useGanache, err = strconv.ParseBool(str)
	// 	s.Require().NoError(err)
	// }

	// The bootstrapping phase is as follows:
	//
	// 1. Initialize Umee validator nodes.
	// 2. Launch an Ethereum container that mines.
	// 3. Create and initialize Umee validator genesis files (setting delegate keys for validators).
	// 4. Start Umee network.
	// 5. Run an Oracle price feeder.
	// 6. Create and run Gaia container(s).
	// 7. Create and run IBC relayer (Hermes) containers.
	// 8. Deploy the Gravity Bridge contract.
	// 9. Create and start Peggo (orchestrator) containers.
	s.initNodes()
	// if useGanache {
	// 	s.runGanacheContainer()
	// } else {
	// 	s.initEthereum()
	// 	s.runEthContainer()
	// }
	s.initGenesis()
	s.initValidatorConfigs()
	s.runValidators()
	s.runPriceFeeder()
	s.runGaiaNetwork()
	s.runIBCRelayer()
	// s.runContractDeployment()
	// s.runOrchestrators()
	s.initUmeeClient()
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if str := os.Getenv("UMEE_E2E_SKIP_CLEANUP"); len(str) > 0 {
		skipCleanup, err := strconv.ParseBool(str)
		s.Require().NoError(err)

		if skipCleanup {
			return
		}
	}

	s.T().Log("tearing down e2e integration test suite...")

	// s.Require().NoError(s.dkrPool.Purge(s.ethResource))
	s.Require().NoError(s.dkrPool.Purge(s.gaiaResource))
	s.Require().NoError(s.dkrPool.Purge(s.hermesResource))
	s.Require().NoError(s.dkrPool.Purge(s.priceFeederResource))

	for _, vc := range s.valResources {
		s.Require().NoError(s.dkrPool.Purge(vc))
	}

	// for _, oc := range s.orchResources {
	// 	s.Require().NoError(s.dkrPool.Purge(oc))
	// }

	s.Require().NoError(s.dkrPool.RemoveNetwork(s.dkrNet))

	os.RemoveAll(s.chain.dataDir)
	for _, td := range s.tmpDirs {
		os.RemoveAll(td)
	}
}

func (s *IntegrationTestSuite) initNodes() {
	s.Require().NoError(s.chain.createAndInitValidators(3))
	s.Require().NoError(s.chain.createAndInitOrchestrators(3))
	s.Require().NoError(s.chain.createAndInitGaiaValidator())

	// initialize a genesis file for the first validator
	val0ConfigDir := s.chain.validators[0].configDir()
	for _, val := range s.chain.validators {
		valAddr, err := val.keyInfo.GetAddress()
		s.Require().NoError(err)
		s.Require().NoError(
			addGenesisAccount(val0ConfigDir, "", initBalanceStr, valAddr),
		)
	}

	// add orchestrator accounts to genesis file
	for _, orch := range s.chain.orchestrators {
		orchAddr, err := orch.keyInfo.GetAddress()
		s.Require().NoError(err)

		s.Require().NoError(
			addGenesisAccount(val0ConfigDir, "", initBalanceStr, orchAddr),
		)
	}

	// copy the genesis file to the remaining validators
	for _, val := range s.chain.validators[1:] {
		_, err := copyFile(
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
	for _, orch := range s.chain.orchestrators {
		s.Require().NoError(orch.generateEthereumKey())
		ethGenesis.Alloc[orch.ethereumKey.address] = alloc
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
	s.T().Log("starting e2e infrastructure; validator_0 config:", genFilePath)
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	s.Require().NoError(err)

	// Gravity Bridge
	var gravityGenState gravitytypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[gravitytypes.ModuleName], &gravityGenState))

	gravityGenState.Params.BridgeChainId = uint64(ethChainID)

	bz, err := cdc.MarshalJSON(&gravityGenState)
	s.Require().NoError(err)
	appGenState[gravitytypes.ModuleName] = bz

	var bech32GenState bech32ibctypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[bech32ibctypes.ModuleName], &bech32GenState))

	// bech32
	bech32GenState.NativeHRP = sdk.GetConfig().GetBech32AccountAddrPrefix()

	bz, err = cdc.MarshalJSON(&bech32GenState)
	s.Require().NoError(err)
	appGenState[bech32ibctypes.ModuleName] = bz

	// Leverage
	var leverageGenState leveragetypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[leveragetypes.ModuleName], &leverageGenState))

	leverageGenState.Registry = append(leverageGenState.Registry,
		fixtures.Token(appparams.BondDenom, appparams.DisplayDenom, 6),
		fixtures.Token(ATOMBaseDenom, ATOM, uint32(ATOMExponent)),
	)

	bz, err = cdc.MarshalJSON(&leverageGenState)
	s.Require().NoError(err)
	appGenState[leveragetypes.ModuleName] = bz

	// Oracle
	var oracleGenState oracletypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[oracletypes.ModuleName], &oracleGenState))

	oracleGenState.Params.HistoricStampPeriod = 5
	oracleGenState.Params.MaximumPriceStamps = 4
	oracleGenState.Params.MedianStampPeriod = 20
	oracleGenState.Params.MaximumMedianStamps = 2

	oracleGenState.Params.AcceptList = append(oracleGenState.Params.AcceptList, oracletypes.Denom{
		BaseDenom:   ATOMBaseDenom,
		SymbolDenom: ATOM,
		Exponent:    uint32(ATOMExponent),
	})

	bz, err = cdc.MarshalJSON(&oracleGenState)
	s.Require().NoError(err)
	appGenState[oracletypes.ModuleName] = bz

	// Gov
	var govGenState govtypesv1.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[govtypes.ModuleName], &govGenState))

	votingPeroid := 5 * time.Second
	govGenState.VotingParams.VotingPeriod = &votingPeroid
	govGenState.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, sdk.NewInt(100)))

	bz, err = cdc.MarshalJSON(&govGenState)
	s.Require().NoError(err)
	appGenState[govtypes.ModuleName] = bz

	// Bank
	var bankGenState banktypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState))

	bankGenState.DenomMetadata = append(bankGenState.DenomMetadata, banktypes.Metadata{
		Description: "An example stable token",
		Display:     photonDenom,
		Base:        photonDenom,
		Symbol:      photonDenom,
		Name:        photonDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    photonDenom,
				Exponent: 0,
			},
		},
	})

	bz, err = cdc.MarshalJSON(&bankGenState)
	s.Require().NoError(err)
	appGenState[banktypes.ModuleName] = bz

	// uibc (ibc quota)
	var uibcGenState uibc.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[uibc.ModuleName], &uibcGenState))

	// 100$ for each token
	uibcGenState.Params.TokenQuota = sdk.NewDec(100)
	// 120$ for all tokens on quota duration
	uibcGenState.Params.TotalQuota = sdk.NewDec(120)
	// quotas will reset every 300 seconds
	uibcGenState.Params.QuotaDuration = time.Second * 300

	bz, err = cdc.MarshalJSON(&uibcGenState)
	s.Require().NoError(err)
	appGenState[uibc.ModuleName] = bz

	var genUtilGenState genutiltypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState))

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(s.chain.validators))
	for i, val := range s.chain.validators {
		var createValmsg sdk.Msg
		if i == 2 {
			createValmsg, err = val.buildCreateValidatorMsg(stakeAmountCoin2)
		} else {
			createValmsg, err = val.buildCreateValidatorMsg(stakeAmountCoin)
		}
		s.Require().NoError(err)

		orchAddr, err := s.chain.orchestrators[i].keyInfo.GetAddress()
		s.Require().NoError(err)

		delKeysMsg, err := val.buildDelegateKeysMsg(orchAddr, s.chain.orchestrators[i].ethereumKey.address)
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

	bz, err = tmjson.MarshalIndent(genDoc, "", "  ")
	s.Require().NoError(err)

	// write the updated genesis file to each validator
	for _, val := range s.chain.validators {
		writeFile(filepath.Join(val.configDir(), "config", "genesis.json"), bz)
	}
}

func (s *IntegrationTestSuite) initValidatorConfigs() {
	for i, val := range s.chain.validators {
		tmCfgPath := filepath.Join(val.configDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(tmCfgPath)
		s.Require().NoError(vpr.ReadInConfig())

		valConfig := tmconfig.DefaultConfig()
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

		tmconfig.WriteConfigFile(tmCfgPath, valConfig)

		// set application configuration
		appCfgPath := filepath.Join(val.configDir(), "config", "app.toml")

		appConfig := srvconfig.DefaultConfig()
		appConfig.API.Enable = true
		appConfig.MinGasPrices = minGasPrice

		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}
}

func (s *IntegrationTestSuite) runGanacheContainer() {
	s.T().Log("starting Ganache container...")

	tmpDir, err := os.MkdirTemp("", "umee-e2e-testnet-eth-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	_, err = copyFile(
		filepath.Join("./docker/", "ganache.Dockerfile"),
		filepath.Join(tmpDir, "ganache.Dockerfile"),
	)
	s.Require().NoError(err)

	entrypoint := []string{
		"ganache-cli",
		"-h",
		"0.0.0.0",
		"--networkId",
		"15",
	}

	entrypoint = append(entrypoint, "--account", "0xb1bab011e03a9862664706fc3bbaa1b16651528e5f0e7fbfcbfdd8be302a13e7,0x3635C9ADC5DEA00000")
	for _, orch := range s.chain.orchestrators {
		s.Require().NoError(orch.generateEthereumKey())
		entrypoint = append(entrypoint, "--account", orch.ethereumKey.privateKey+",0x3635C9ADC5DEA00000")
	}

	s.ethResource, err = s.dkrPool.BuildAndRunWithBuildOptions(
		&dockertest.BuildOptions{
			Dockerfile: "ganache.Dockerfile",
			ContextDir: tmpDir,
		},
		&dockertest.RunOptions{
			Name:         "ganache",
			NetworkID:    s.dkrNet.Network.ID,
			ExposedPorts: []string{"8545"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"8545/tcp": {{HostIP: "", HostPort: "8545"}},
			},
			Env:        []string{},
			Entrypoint: entrypoint,
		},
		noRestart,
	)
	s.Require().NoError(err)

	s.ethClient, err = ethclient.Dial(fmt.Sprintf("http://%s", s.ethResource.GetHostPort("8545/tcp")))
	s.Require().NoError(err)

	match := "Listening on 0.0.0.0:8545"

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	// Wait for Ganache to start running.
	s.Require().Eventually(
		func() bool {
			err := s.dkrPool.Client.Logs(
				docker.LogsOptions{
					Container:    s.ethResource.Container.ID,
					OutputStream: &outBuf,
					ErrorStream:  &errBuf,
					Stdout:       true,
					Stderr:       true,
				},
			)
			if err != nil {
				return false
			}

			return strings.Contains(outBuf.String(), match)
		},
		1*time.Minute,
		5*time.Second,
		"ganache node failed to start",
	)
	s.T().Logf("started Ganache container: %s", s.ethResource.Container.ID)
}

func (s *IntegrationTestSuite) runEthContainer() {
	s.T().Log("starting Ethereum container...")
	tmpDir, err := os.MkdirTemp("", "umee-e2e-testnet-eth-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	_, err = copyFile(
		filepath.Join(s.chain.configDir(), "eth_genesis.json"),
		filepath.Join(tmpDir, "eth_genesis.json"),
	)
	s.Require().NoError(err)

	_, err = copyFile(
		filepath.Join("./docker/", "eth.Dockerfile"),
		filepath.Join(tmpDir, "eth.Dockerfile"),
	)
	s.Require().NoError(err)

	s.ethResource, err = s.dkrPool.BuildAndRunWithBuildOptions(
		&dockertest.BuildOptions{
			Dockerfile: "eth.Dockerfile",
			ContextDir: tmpDir,
		},
		&dockertest.RunOptions{
			Name:      "ethereum",
			NetworkID: s.dkrNet.Network.ID,
			PortBindings: map[docker.Port][]docker.PortBinding{
				"8545/tcp": {{HostIP: "", HostPort: "8545"}},
			},
			Env: []string{},
		},
		noRestart,
	)
	s.Require().NoError(err)

	s.ethClient, err = ethclient.Dial(fmt.Sprintf("http://%s", s.ethResource.GetHostPort("8545/tcp")))
	s.Require().NoError(err)

	// Wait for the Ethereum node to start producing blocks; DAG completion takes
	// about two minutes.
	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			height, err := s.ethClient.BlockNumber(ctx)
			if err != nil {
				return false
			}

			return height > 1
		},
		5*time.Minute,
		10*time.Second,
		"geth node failed to produce a block",
	)

	s.T().Logf("started Ethereum container: %s", s.ethResource.Container.ID)
}

func (s *IntegrationTestSuite) runValidators() {
	s.T().Log("starting Umee validator containers...")

	s.valResources = make([]*dockertest.Resource, len(s.chain.validators))
	for i, val := range s.chain.validators {
		runOpts := &dockertest.RunOptions{
			Name:      val.instanceName(),
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/root/.umee", val.configDir()),
			},
			Repository: "umee-network/umeed-e2e",
		}

		// expose the first validator for debugging and communication
		if val.index == 0 {
			runOpts.PortBindings = map[docker.Port][]docker.PortBinding{
				"1317/tcp":  {{HostIP: "", HostPort: "1317"}},
				"6060/tcp":  {{HostIP: "", HostPort: "6060"}},
				"6061/tcp":  {{HostIP: "", HostPort: "6061"}},
				"6062/tcp":  {{HostIP: "", HostPort: "6062"}},
				"6063/tcp":  {{HostIP: "", HostPort: "6063"}},
				"6064/tcp":  {{HostIP: "", HostPort: "6064"}},
				"6065/tcp":  {{HostIP: "", HostPort: "6065"}},
				"9090/tcp":  {{HostIP: "", HostPort: "9090"}},
				"26656/tcp": {{HostIP: "", HostPort: "26656"}},
				"26657/tcp": {{HostIP: "", HostPort: "26657"}},
			}
		}

		resource, err := s.dkrPool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		s.valResources[i] = resource
		s.T().Logf("started Umee validator container: %s", resource.Container.ID)
	}

	rpcClient, err := rpchttp.New("tcp://localhost:26657", "/websocket")
	s.Require().NoError(err)

	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			status, err := rpcClient.Status(ctx)
			if err != nil {
				return false
			}

			// let the node produce a few blocks
			if status.SyncInfo.CatchingUp || status.SyncInfo.LatestBlockHeight < 3 {
				return false
			}

			return true
		},
		5*time.Minute,
		time.Second,
		"umee node failed to produce blocks",
	)
}

func (s *IntegrationTestSuite) runGaiaNetwork() {
	s.T().Log("starting Gaia network container...")

	tmpDir, err := os.MkdirTemp("", "umee-e2e-testnet-gaia-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	gaiaVal := s.chain.gaiaValidators[0]

	gaiaCfgPath := path.Join(tmpDir, "cfg")
	s.Require().NoError(os.MkdirAll(gaiaCfgPath, 0o755))

	_, err = copyFile(
		filepath.Join("./scripts/", "gaia_bootstrap.sh"),
		filepath.Join(gaiaCfgPath, "gaia_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.gaiaResource, err = s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       gaiaVal.instanceName(),
			Repository: "ghcr.io/umee-network/gaia-e2e",
			Tag:        "latest",
			NetworkID:  s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/root/.gaia", tmpDir),
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"1317/tcp":  {{HostIP: "", HostPort: "1417"}},
				"9090/tcp":  {{HostIP: "", HostPort: "9190"}},
				"26656/tcp": {{HostIP: "", HostPort: "27656"}},
				"26657/tcp": {{HostIP: "", HostPort: "27657"}},
			},
			Env: []string{
				fmt.Sprintf("UMEE_E2E_GAIA_CHAIN_ID=%s", gaiaChainID),
				fmt.Sprintf("UMEE_E2E_GAIA_VAL_MNEMONIC=%s", gaiaVal.mnemonic),
			},
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x /root/.gaia/cfg/gaia_bootstrap.sh && /root/.gaia/cfg/gaia_bootstrap.sh",
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	endpoint := fmt.Sprintf("tcp://%s", s.gaiaResource.GetHostPort("26657/tcp"))
	s.gaiaRPC, err = rpchttp.New(endpoint, "/websocket")
	s.Require().NoError(err)

	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			status, err := s.gaiaRPC.Status(ctx)
			if err != nil {
				return false
			}

			// let the node produce a few blocks
			if status.SyncInfo.CatchingUp || status.SyncInfo.LatestBlockHeight < 3 {
				return false
			}

			return true
		},
		5*time.Minute,
		time.Second,
		"gaia node failed to produce blocks",
	)

	s.T().Logf("started Gaia network container: %s", s.gaiaResource.Container.ID)
}

func (s *IntegrationTestSuite) runIBCRelayer() {
	s.T().Log("starting Hermes relayer container...")

	tmpDir, err := os.MkdirTemp("", "umee-e2e-testnet-hermes-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	gaiaVal := s.chain.gaiaValidators[0]
	// umeeVal for the relayer needs to be a different account
	// than what we use for runPriceFeeder.
	umeeVal := s.chain.validators[1]
	hermesCfgPath := path.Join(tmpDir, "hermes")

	s.Require().NoError(os.MkdirAll(hermesCfgPath, 0o755))
	_, err = copyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.hermesResource, err = s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       "umee-gaia-relayer",
			Repository: "ghcr.io/umee-network/hermes-e2e",
			Tag:        "latest",
			NetworkID:  s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/home/hermes", hermesCfgPath),
			},
			ExposedPorts: []string{"3031"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"3031/tcp": {{HostIP: "", HostPort: "3031"}},
			},
			Env: []string{
				fmt.Sprintf("UMEE_E2E_GAIA_CHAIN_ID=%s", gaiaChainID),
				fmt.Sprintf("UMEE_E2E_UMEE_CHAIN_ID=%s", s.chain.id),
				fmt.Sprintf("UMEE_E2E_GAIA_VAL_MNEMONIC=%s", gaiaVal.mnemonic),
				fmt.Sprintf("UMEE_E2E_UMEE_VAL_MNEMONIC=%s", umeeVal.mnemonic),
				fmt.Sprintf("UMEE_E2E_GAIA_VAL_HOST=%s", s.gaiaResource.Container.Name[1:]),
				fmt.Sprintf("UMEE_E2E_UMEE_VAL_HOST=%s", s.valResources[1].Container.Name[1:]),
			},
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x /home/hermes/hermes_bootstrap.sh && /home/hermes/hermes_bootstrap.sh",
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	endpoint := fmt.Sprintf("http://%s/state", s.hermesResource.GetHostPort("3031/tcp"))
	s.Require().Eventually(
		func() bool {
			resp, err := http.Get(endpoint)
			if err != nil {
				return false
			}

			defer resp.Body.Close()

			bz, err := io.ReadAll(resp.Body)
			if err != nil {
				return false
			}

			var respBody map[string]interface{}
			if err := json.Unmarshal(bz, &respBody); err != nil {
				return false
			}

			status := respBody["status"].(string)
			result := respBody["result"].(map[string]interface{})

			return status == "success" && len(result["chains"].([]interface{})) == 2
		},
		5*time.Minute,
		time.Second,
		"hermes relayer not healthy",
	)

	s.T().Logf("started Hermes relayer container: %s", s.hermesResource.Container.ID)

	// create the client, connection and channel between the Umee and Gaia chains
	s.connectIBCChains()
}

func (s *IntegrationTestSuite) runContractDeployment() {
	s.T().Log("starting Gravity Bridge contract deployer container...")

	resource, err := s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       "gravity-contract-deployer",
			NetworkID:  s.dkrNet.Network.ID,
			Repository: "umee-network/umeed-e2e",
			// NOTE: container names are prefixed with '/'
			Env: []string{"PEGGO_ETH_PK=" + ethMinerPK},
			Entrypoint: []string{
				"peggo",
				"bridge",
				"deploy-gravity",
				"--eth-rpc",
				fmt.Sprintf("http://%s:8545", s.ethResource.Container.Name[1:]),
				"--cosmos-grpc",
				fmt.Sprintf("tcp://%s:9090", s.valResources[0].Container.Name[1:]),
				"--tendermint-rpc",
				fmt.Sprintf("http://%s:26657", s.valResources[0].Container.Name[1:]),
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	s.T().Logf("started contract deployer: %s", resource.Container.ID)

	// wait for the container to finish executing
	container := resource.Container
	for container.State.Running {
		time.Sleep(10 * time.Second)

		container, err = s.dkrPool.Client.InspectContainer(resource.Container.ID)
		s.Require().NoError(err)
	}

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	s.Require().NoErrorf(s.dkrPool.Client.Logs(
		docker.LogsOptions{
			Container:    resource.Container.ID,
			OutputStream: &outBuf,
			ErrorStream:  &errBuf,
			Stdout:       true,
			Stderr:       true,
		},
	),
		"failed to start contract deployer; stdout: %s, stderr: %s",
		outBuf.String(), errBuf.String(),
	)

	re := regexp.MustCompile(`Address: (0x.+)`)
	tokens := re.FindStringSubmatch(errBuf.String())
	s.Require().Len(tokens, 2)

	gravityContractAddr := tokens[1]
	s.Require().NotEmpty(gravityContractAddr)

	re = regexp.MustCompile(`Transaction: (0x.+)`)
	tokens = re.FindStringSubmatch(errBuf.String())
	s.Require().Len(tokens, 2)

	txHash := tokens[1]
	s.Require().NotEmpty(txHash)

	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := queryEthTx(ctx, s.ethClient, txHash); err != nil {
				return false
			}

			return true
		},
		time.Minute,
		time.Second,
		"failed to confirm Peggy contract deployment transaction",
	)

	s.Require().NoError(s.dkrPool.RemoveContainerByName(container.Name))

	s.T().Logf("deployed Gravity Bridge contract: %s", gravityContractAddr)
	s.gravityContractAddr = gravityContractAddr
}

func (s *IntegrationTestSuite) runOrchestrators() {
	s.T().Log("starting orchestrator containers...")

	s.orchResources = make([]*dockertest.Resource, len(s.chain.validators))
	for i, orch := range s.chain.orchestrators {
		resource, err := s.dkrPool.RunWithOptions(
			&dockertest.RunOptions{
				Name:       s.chain.orchestrators[i].instanceName(),
				NetworkID:  s.dkrNet.Network.ID,
				Repository: "umee-network/umeed-e2e",
				Env:        []string{"PEGGO_ETH_PK=" + orch.ethereumKey.privateKey},
				// NOTE: container names are prefixed with '/'
				Entrypoint: []string{
					"peggo",
					"orchestrator",
					s.gravityContractAddr,
					"--eth-rpc",
					fmt.Sprintf("http://%s:8545", s.ethResource.Container.Name[1:]),
					"--cosmos-chain-id",
					s.chain.id,
					"--cosmos-grpc",
					fmt.Sprintf("tcp://%s:9090", s.valResources[i].Container.Name[1:]),
					"--tendermint-rpc",
					fmt.Sprintf("http://%s:26657", s.valResources[i].Container.Name[1:]),
					"--cosmos-gas-prices",
					minGasPrice,
					"--cosmos-from",
					orch.keyInfo.Name,
					"--relay-batches=true",
					"--valset-relay-mode=minimum",
					"--profit-multiplier=0.0",
					"--oracle-providers=mock",
					"--relayer-loop-multiplier=1.0",
					"--requester-loop-multiplier=1.0",
					"--cosmos-pk",
					hexutil.Encode(s.chain.orchestrators[i].privateKey.Bytes()),
				},
			},
			noRestart,
		)
		s.Require().NoError(err)

		s.orchResources[i] = resource
		s.T().Logf("started orchestrator container: %s", resource.Container.ID)
	}

	match := "oracle sent set of claims successfully"
	for _, resource := range s.orchResources {
		s.T().Logf("waiting for orchestrator to be healthy: %s", resource.Container.ID)

		var (
			outBuf bytes.Buffer
			errBuf bytes.Buffer
		)

		s.Require().Eventuallyf(
			func() bool {
				err := s.dkrPool.Client.Logs(
					docker.LogsOptions{
						Container:    resource.Container.ID,
						OutputStream: &outBuf,
						ErrorStream:  &errBuf,
						Stdout:       true,
						Stderr:       true,
					},
				)
				if err != nil {
					return false
				}

				return strings.Contains(errBuf.String(), match)
			},
			5*time.Minute,
			time.Second,
			"orchestrator %s not healthy",
			resource.Container.ID,
		)
	}
}

func (s *IntegrationTestSuite) runPriceFeeder() {
	s.T().Log("starting price-feeder container...")

	umeeVal := s.chain.validators[2]
	umeeValAddr, err := umeeVal.keyInfo.GetAddress()
	s.Require().NoError(err)

	grpcEndpoint := fmt.Sprintf("tcp://%s:%s", umeeVal.instanceName(), "9090")
	tmrpcEndpoint := fmt.Sprintf("http://%s:%s", umeeVal.instanceName(), "26657")

	s.priceFeederResource, err = s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       "umee-price-feeder",
			NetworkID:  s.dkrNet.Network.ID,
			Repository: priceFeederContainerRepo,
			Mounts: []string{
				fmt.Sprintf("%s/:/root/.umee", umeeVal.configDir()),
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				priceFeederServerPort: {{HostIP: "", HostPort: "7171"}},
			},
			Env: []string{
				fmt.Sprintf("PRICE_FEEDER_PASS=%s", keyringPassphrase),
				fmt.Sprintf("ACCOUNT_ADDRESS=%s", umeeValAddr),
				fmt.Sprintf("ACCOUNT_VALIDATOR=%s", sdk.ValAddress(umeeValAddr)),
				fmt.Sprintf("KEYRING_DIR=%s", "/root/.umee"),
				fmt.Sprintf("ACCOUNT_CHAIN_ID=%s", s.chain.id),
				fmt.Sprintf("RPC_GRPC_ENDPOINT=%s", grpcEndpoint),
				fmt.Sprintf("RPC_TMRPC_ENDPOINT=%s", tmrpcEndpoint),
			},
			Cmd: []string{
				"--skip-provider-check",
				"--log-level=debug",
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	endpoint := fmt.Sprintf("http://%s/api/v1/prices", s.priceFeederResource.GetHostPort(priceFeederServerPort))

	checkHealth := func() bool {
		resp, err := http.Get(endpoint)
		if err != nil {
			s.T().Log("Price feeder endpoint not available", err)
			return false
		}

		defer resp.Body.Close()

		bz, err := io.ReadAll(resp.Body)
		if err != nil {
			s.T().Log("Can't get price feeder response", err)
			return false
		}

		var respBody map[string]interface{}
		if err := json.Unmarshal(bz, &respBody); err != nil {
			s.T().Log("Can't unmarshal price feed", err)
			return false
		}

		prices, ok := respBody["prices"].(map[string]interface{})
		if !ok {
			s.T().Log("price feeder: no prices")
			return false
		}

		return len(prices) > 0
	}

	isHealthy := false
	for i := 0; i < priceFeederMaxStartupTime; i++ {
		isHealthy = checkHealth()
		if isHealthy {
			break
		}
		time.Sleep(time.Second)
	}

	if !isHealthy {
		err := s.dkrPool.Client.Logs(docker.LogsOptions{
			Container:    s.priceFeederResource.Container.ID,
			OutputStream: os.Stdout,
			ErrorStream:  os.Stderr,
			Stdout:       true,
			Stderr:       true,
			Tail:         "false",
		})
		if err != nil {
			s.T().Log("Error retrieving price feeder logs", err)
		}

		s.T().Fatal("price-feeder not healthy")
	}

	s.T().Logf("started price-feeder container: %s", s.priceFeederResource.Container.ID)
}

func (s *IntegrationTestSuite) initUmeeClient() {
	var err error
	mnemonics := make([]string, 0)
	for _, v := range s.chain.validators {
		mnemonics = append(mnemonics, v.mnemonic)
	}
	ecfg := app.MakeEncodingConfig()

	s.umee, err = client.NewClient(
		s.chain.id,
		"tcp://localhost:26657",
		"tcp://localhost:9090",
		mnemonics,
		1,
		ecfg,
	)
	s.Require().NoError(err)
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}
