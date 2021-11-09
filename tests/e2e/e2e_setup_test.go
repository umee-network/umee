package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	tmconfig "github.com/tendermint/tendermint/config"
	tmjson "github.com/tendermint/tendermint/libs/json"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/umee-network/umee/app"
	peggytypes "github.com/umee-network/umee/x/peggy/types"
)

const (
	photonDenom    = "photon"
	initBalanceStr = "110000000000uumee,100000000000photon"
	minGasPrice    = "0.00001"
	gaiaChainID    = "test-gaia-chain"

	ethChainID uint = 15
	ethMinerPK      = "0xb1bab011e03a9862664706fc3bbaa1b16651528e5f0e7fbfcbfdd8be302a13e7"
)

var (
	stakeAmount, _  = sdk.NewIntFromString("100000000000")
	stakeAmountCoin = sdk.NewCoin(app.BondDenom, stakeAmount)
)

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs           []string
	chain             *chain
	ethClient         *ethclient.Client
	gaiaRPC           *rpchttp.HTTP
	dkrPool           *dockertest.Pool
	dkrNet            *dockertest.Network
	ethResource       *dockertest.Resource
	gaiaResource      *dockertest.Resource
	hermesResource    *dockertest.Resource
	valResources      []*dockertest.Resource
	orchResources     []*dockertest.Resource
	peggyContractAddr string
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

	// The boostrapping phase is as follows:
	//
	// 1. Initialize Umee validator nodes.
	// 2. Launch an Ethereum container that mines.
	// 3. Deploy the Peggy (Gravity Bridge) contract
	// 4. Create and initialize Umee validator genesis files.
	// 5. Start Umee network.
	// 6. Register each validator's Ethereum key.
	// 7. Invoke the initialize method on the Peggy contract.
	// 8. Create and start peggo (orchestrator) containers.
	// 9. Create and run Gaia container(s).
	// 10. Create and run IBC relayer (Hermes) containers.
	s.initNodes()
	s.initEthereum()
	s.runEthContainer()
	s.runContractDeployment()
	s.initGenesis()
	s.initValidatorConfigs()
	s.runValidators()
	s.runGaiaNetwork()
	s.runIBCRelayer()
	s.registerEthKeys()
	s.initPeggy()
	s.runOrchestrators()
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

	os.RemoveAll(s.chain.dataDir)
	for _, td := range s.tmpDirs {
		os.RemoveAll(td)
	}

	s.Require().NoError(s.dkrPool.Purge(s.ethResource))
	s.Require().NoError(s.dkrPool.Purge(s.gaiaResource))
	s.Require().NoError(s.dkrPool.Purge(s.hermesResource))

	for _, vc := range s.valResources {
		s.Require().NoError(s.dkrPool.Purge(vc))
	}

	for _, oc := range s.orchResources {
		s.Require().NoError(s.dkrPool.Purge(oc))
	}

	s.Require().NoError(s.dkrPool.RemoveNetwork(s.dkrNet))
}

func (s *IntegrationTestSuite) initNodes() {
	s.Require().NoError(s.chain.createAndInitValidators(2))
	s.Require().NoError(s.chain.createAndInitOrchestrators(2))
	s.Require().NoError(s.chain.createAndInitGaiaValidator())

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

	var peggyGenState peggytypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[peggytypes.ModuleName], &peggyGenState))

	peggyGenState.Params.BridgeEthereumAddress = s.peggyContractAddr
	peggyGenState.Params.BridgeContractStartHeight = 0
	peggyGenState.Params.BridgeChainId = uint64(ethChainID)

	bz, err := cdc.MarshalJSON(&peggyGenState)
	s.Require().NoError(err)
	appGenState[peggytypes.ModuleName] = bz

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

	var genUtilGenState genutiltypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState))

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(s.chain.validators))
	for i, val := range s.chain.validators {
		createValmsg, err := val.buildCreateValidatorMsg(stakeAmountCoin)
		s.Require().NoError(err)

		signedTx, err := val.signMsg(createValmsg)
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

		tmconfig.WriteConfigFile(tmCfgPath, valConfig)

		// set application configuration
		appCfgPath := filepath.Join(val.configDir(), "config", "app.toml")

		appConfig := srvconfig.DefaultConfig()
		appConfig.API.Enable = true
		appConfig.MinGasPrices = fmt.Sprintf("%s%s", minGasPrice, photonDenom)

		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}
}

func (s *IntegrationTestSuite) runEthContainer() {
	s.T().Log("starting Ethereum container...")

	tmpDir, err := ioutil.TempDir("", "umee-e2e-testnet-eth-")
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
			Repository: "umeenet/umeed",
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

	tmpDir, err := ioutil.TempDir("", "umee-e2e-testnet-gaia-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	gaiaVal := s.chain.gaiaValidators[0]

	gaiaCfgPath := path.Join(tmpDir, "cfg")
	s.Require().NoError(os.MkdirAll(gaiaCfgPath, 0755))

	_, err = copyFile(
		filepath.Join("./scripts/", "gaia_bootstrap.sh"),
		filepath.Join(gaiaCfgPath, "gaia_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.gaiaResource, err = s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       gaiaVal.instanceName(),
			Repository: "umeenet/gaia",
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

	tmpDir, err := ioutil.TempDir("", "umee-e2e-testnet-hermes-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	gaiaVal := s.chain.gaiaValidators[0]
	umeeVal := s.chain.validators[0]
	hermesCfgPath := path.Join(tmpDir, "hermes")

	s.Require().NoError(os.MkdirAll(hermesCfgPath, 0755))
	_, err = copyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	s.Require().NoError(err)

	_, err = copyFile(
		filepath.Join("./docker/", "hermes.Dockerfile"),
		filepath.Join(tmpDir, "hermes.Dockerfile"),
	)
	s.Require().NoError(err)

	s.hermesResource, err = s.dkrPool.BuildAndRunWithBuildOptions(
		&dockertest.BuildOptions{
			Dockerfile: "hermes.Dockerfile",
			ContextDir: tmpDir,
		},
		&dockertest.RunOptions{
			Name:      "umee-gaia-relayer",
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/root/hermes", hermesCfgPath),
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"3031/tcp": {{HostIP: "", HostPort: "3031"}},
			},
			Env: []string{
				fmt.Sprintf("UMEE_E2E_GAIA_CHAIN_ID=%s", gaiaChainID),
				fmt.Sprintf("UMEE_E2E_UMEE_CHAIN_ID=%s", s.chain.id),
				fmt.Sprintf("UMEE_E2E_GAIA_VAL_MNEMONIC=%s", gaiaVal.mnemonic),
				fmt.Sprintf("UMEE_E2E_UMEE_VAL_MNEMONIC=%s", umeeVal.mnemonic),
				fmt.Sprintf("UMEE_E2E_GAIA_VAL_HOST=%s", s.gaiaResource.Container.Name[1:]),
				fmt.Sprintf("UMEE_E2E_UMEE_VAL_HOST=%s", s.valResources[0].Container.Name[1:]),
			},
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x /root/hermes/hermes_bootstrap.sh && /root/hermes/hermes_bootstrap.sh",
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
	s.T().Log("starting contract deployer container...")

	resource, err := s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       "peggy-contract-deployer",
			NetworkID:  s.dkrNet.Network.ID,
			Repository: "umeenet/umeed",
			// NOTE: container names are prefixed with '/'
			Entrypoint: []string{
				"peggo",
				"bridge",
				"deploy-peggy",
				"--eth-pk",
				ethMinerPK[2:], // remove 0x prefix
				"--eth-rpc",
				fmt.Sprintf("http://%s:8545", s.ethResource.Container.Name[1:]),
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

	peggyContractAddr := tokens[1]
	s.Require().NotEmpty(peggyContractAddr)

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

	s.T().Logf("deployed Peggy (Gravity Bridge) contract: %s", peggyContractAddr)
	s.peggyContractAddr = peggyContractAddr
}

func (s *IntegrationTestSuite) registerEthKeys() {
	s.T().Log("registering Umee validator Ethereum keys...")

	resources := make([]*dockertest.Resource, len(s.chain.validators))
	for i, val := range s.chain.validators {
		resource, err := s.dkrPool.RunWithOptions(
			&dockertest.RunOptions{
				Name:       fmt.Sprintf("peggy-key-registration-%d", i),
				NetworkID:  s.dkrNet.Network.ID,
				Repository: "umeenet/umeed",
				Mounts: []string{
					fmt.Sprintf("%s/:/root/.umee", val.configDir()),
				},
				// NOTE: container names are prefixed with '/'
				Entrypoint: []string{
					"peggo",
					"tx",
					"register-eth-key",
					"--eth-pk",
					val.ethereumKey.privateKey[2:], // remove 0x prefix
					"--cosmos-chain-id",
					s.chain.id,
					"--cosmos-grpc",
					fmt.Sprintf("tcp://%s:9090", s.valResources[i].Container.Name[1:]),
					"--tendermint-rpc",
					fmt.Sprintf("http://%s:26657", s.valResources[i].Container.Name[1:]),
					"--cosmos-from",
					val.keyInfo.GetName(),
					"--cosmos-gas-prices",
					fmt.Sprintf("%s%s", minGasPrice, photonDenom),
					"--cosmos-keyring-dir=/root/.umee",
					"--cosmos-keyring=test",
					"-y",
				},
			},
			noRestart,
		)
		s.Require().NoError(err)

		resources[i] = resource
		s.T().Logf("started Umee validator Ethereum key registration: %s", resource.Container.ID)
	}

	for i, r := range resources {
		var err error

		// wait for the container to finish executing
		container := r.Container
		for container.State.Running {
			time.Sleep(10 * time.Second)

			container, err = s.dkrPool.Client.InspectContainer(r.Container.ID)
			s.Require().NoError(err)
		}

		s.Require().NoError(s.dkrPool.RemoveContainerByName(container.Name))
		s.T().Logf("registered Ethereum key for validator: %s", s.valResources[i].Container.Name[1:])
	}
}

func (s *IntegrationTestSuite) initPeggy() {
	s.T().Log("initializing Peggy contract...")

	resource, err := s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       "peggy-contract-init",
			NetworkID:  s.dkrNet.Network.ID,
			Repository: "umeenet/umeed",
			// NOTE: container names are prefixed with '/'
			Entrypoint: []string{
				"peggo",
				"bridge",
				"init-peggy",
				"--eth-pk",
				ethMinerPK[2:], // remove 0x prefix
				"--eth-rpc",
				fmt.Sprintf("http://%s:8545", s.ethResource.Container.Name[1:]),
				"--cosmos-chain-id",
				s.chain.id,
				"--cosmos-grpc",
				fmt.Sprintf("tcp://%s:9090", s.valResources[0].Container.Name[1:]),
				"--tendermint-rpc",
				fmt.Sprintf("http://%s:26657", s.valResources[0].Container.Name[1:]),
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	s.T().Logf("started Peggy contract initializer: %s", resource.Container.ID)

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
		"failed to get Peggy initializer logs; stdout: %s, stderr: %s",
		outBuf.String(), errBuf.String(),
	)

	re := regexp.MustCompile(`Transaction: (0x.+)`)
	tokens := re.FindStringSubmatch(errBuf.String())
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
		"failed to confirm Peggy initialization transaction",
	)

	s.Require().NoError(s.dkrPool.RemoveContainerByName(container.Name))

	s.T().Log("initialized Peggy (Gravity Bridge) contract")
}

func (s *IntegrationTestSuite) runOrchestrators() {
	s.T().Log("starting orchestrator containers...")

	s.orchResources = make([]*dockertest.Resource, len(s.chain.validators))
	for i, val := range s.chain.validators {
		resource, err := s.dkrPool.RunWithOptions(
			&dockertest.RunOptions{
				Name:       s.chain.orchestrators[i].instanceName(),
				NetworkID:  s.dkrNet.Network.ID,
				Repository: "umeenet/umeed",
				Mounts: []string{
					fmt.Sprintf("%s/:/root/.umee", val.configDir()),
				},
				// NOTE: container names are prefixed with '/'
				Entrypoint: []string{
					"peggo",
					"orchestrator",
					"--eth-pk",
					val.ethereumKey.privateKey[2:], // remove 0x prefix
					"--eth-rpc",
					fmt.Sprintf("http://%s:8545", s.ethResource.Container.Name[1:]),
					"--eth-chain-id",
					fmt.Sprintf("%d", ethChainID),
					"--cosmos-chain-id",
					s.chain.id,
					"--cosmos-grpc",
					fmt.Sprintf("tcp://%s:9090", s.valResources[i].Container.Name[1:]),
					"--tendermint-rpc",
					fmt.Sprintf("http://%s:26657", s.valResources[i].Container.Name[1:]),
					"--cosmos-gas-prices",
					fmt.Sprintf("%s%s", minGasPrice, photonDenom),
					"--cosmos-from",
					val.keyInfo.GetName(),
					"--cosmos-keyring-dir=/root/.umee",
					"--cosmos-keyring=test",
					"--relay-batches=true",
					"--relay-valsets=true",
					"--relayer-loop-duration=1m",
					"--orch-loop-duration=10s",
				},
			},
			noRestart,
		)
		s.Require().NoError(err)

		s.orchResources[i] = resource
		s.T().Logf("started orchestrator container: %s", resource.Container.ID)
	}

	match := "oracle sent ValsetUpdate event successfully"
	for _, resource := range s.orchResources {
		s.T().Logf("waiting for orchestrator to be healthy: %s", resource.Container.ID)

		s.Require().Eventuallyf(
			func() bool {
				var (
					outBuf bytes.Buffer
					errBuf bytes.Buffer
				)

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

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}
