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
)

const (
	photonDenom         = "photon"
	initBalanceStr      = "110000000000uumee,100000000000photon"
	minGasPrice         = "0.00001"
	ethChainID     uint = 15
	gaiaChainID         = "test-gaia-chain"
)

var (
	stakeAmount, _  = sdk.NewIntFromString("100000000000")
	stakeAmountCoin = sdk.NewCoin(app.BondDenom, stakeAmount)
)

type IntegrationTestSuite struct {
	suite.Suite

	chain               *chain
	ethClient           *ethclient.Client
	gaiaRPC             *rpchttp.HTTP
	dkrPool             *dockertest.Pool
	dkrNet              *dockertest.Network
	ethResource         *dockertest.Resource
	gaiaResource        *dockertest.Resource
	hermesResource      *dockertest.Resource
	valResources        []*dockertest.Resource
	orchResources       []*dockertest.Resource
	gravityContractAddr string
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

	// initialization
	s.initNodes()
	s.initEthereum()
	s.initGenesis()
	s.initValidatorConfigs()

	s.dkrPool, err = dockertest.NewPool("")
	s.Require().NoError(err)

	s.dkrNet, err = s.dkrPool.CreateNetwork(fmt.Sprintf("%s-testnet", s.chain.id))
	s.Require().NoError(err)

	// container infrastructure
	s.runValidators()
	s.runGaiaNetwork()
	s.runIBCRelayer()
	s.runEthContainer()
	s.runContractDeployment()
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

	var bankGenState banktypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState))

	bankGenState.DenomMetadata = append(bankGenState.DenomMetadata, banktypes.Metadata{
		Description: "The native staking token of the Umee network",
		Display:     "umee",
		Base:        app.BondDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    app.BondDenom,
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
		Display:     photonDenom,
		Base:        photonDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    photonDenom,
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

	_, err := copyFile(
		filepath.Join("./docker/", "eth.Dockerfile"),
		filepath.Join(s.chain.configDir(), "eth.Dockerfile"),
	)
	s.Require().NoError(err)

	s.ethResource, err = s.dkrPool.BuildAndRunWithBuildOptions(
		&dockertest.BuildOptions{
			Dockerfile: "eth.Dockerfile",
			ContextDir: s.chain.configDir(),
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
			status, err := rpcClient.Status(context.Background())
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

func (s *IntegrationTestSuite) runContractDeployment() {
	s.T().Log("starting contract deployer container...")

	resource, err := s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       "gravity-contract-deployer",
			NetworkID:  s.dkrNet.Network.ID,
			Repository: "umeenet/umeed",
			// NOTE: container names are prefixed with '/'
			Entrypoint: []string{
				"contract-deployer",
				"--cosmos-node",
				fmt.Sprintf("http://%s:26657", s.valResources[0].Container.Name[1:]),
				"--eth-node",
				fmt.Sprintf("http://%s:8545", s.ethResource.Container.Name[1:]),
				"--eth-privkey",
				"0xb1bab011e03a9862664706fc3bbaa1b16651528e5f0e7fbfcbfdd8be302a13e7",
				"--contract",
				"/var/data/Gravity.json",
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	s.T().Logf("started contract deployer: %s", resource.Container.ID)

	// wait for the container to finish executing, i.e. deploys the gravity contract
	container := resource.Container
	for container.State.Running {
		time.Sleep(10 * time.Second)

		container, err = s.dkrPool.Client.InspectContainer(resource.Container.ID)
		s.Require().NoError(err)
	}

	var containerLogsBuf bytes.Buffer
	s.Require().NoError(s.dkrPool.Client.Logs(
		docker.LogsOptions{
			Container:    resource.Container.ID,
			OutputStream: &containerLogsBuf,
			Stdout:       true,
		},
	), containerLogsBuf.String())

	var gravityContractAddr string
	for _, s := range strings.Split(containerLogsBuf.String(), "\n") {
		if strings.HasPrefix(s, "Gravity deployed at Address") {
			tokens := strings.Split(s, "-")
			gravityContractAddr = strings.ReplaceAll(tokens[1], " ", "")
			break
		}
	}

	s.Require().NoError(s.dkrPool.RemoveContainerByName(container.Name))
	s.Require().NotEmpty(gravityContractAddr)

	s.T().Logf("deployed gravity contract: %s", gravityContractAddr)
	s.gravityContractAddr = gravityContractAddr
}

func (s *IntegrationTestSuite) runOrchestrators() {
	s.T().Log("starting orchestrator containers...")

	s.orchResources = make([]*dockertest.Resource, len(s.chain.orchestrators))
	for i, orch := range s.chain.orchestrators {
		gorcCfg := fmt.Sprintf(`keystore = "/root/gorc/keystore/"

[gravity]
contract = "%s"
fees_denom = "%s"

[ethereum]
key_derivation_path = "m/44'/60'/0'/0/0"
rpc = "http://%s:8545"

[cosmos]
key_derivation_path = "m/44'/118'/0'/0/0"
grpc = "http://%s:9090"
prefix = "umee"

[cosmos.gas_price]
amount = %s
denom = "%s"

[metrics]
listen_addr = "127.0.0.1:3000"
`,
			s.gravityContractAddr,
			photonDenom,
			// NOTE: container names are prefixed with '/'
			s.ethResource.Container.Name[1:],
			s.valResources[i].Container.Name[1:],
			minGasPrice,
			photonDenom,
		)

		val := s.chain.validators[i]

		gorcCfgPath := path.Join(val.configDir(), "gorc")
		s.Require().NoError(os.MkdirAll(gorcCfgPath, 0755))

		filePath := path.Join(gorcCfgPath, "config.toml")
		s.Require().NoError(writeFile(filePath, []byte(gorcCfg)))

		// We must first populate the orchestrator's keystore prior to starting
		// the orchestrator gorc process. The keystore must contain the Ethereum
		// and orchestrator keys. These keys will be used for relaying txs to
		// and from Umee and Ethereum. The gorc_bootstrap.sh scripts encapsulates
		// this entire process.
		//
		// NOTE: If the Docker build changes, the script might have to be modified
		// as it relies on busybox.
		_, err := copyFile(
			filepath.Join("./scripts/", "gorc_bootstrap.sh"),
			filepath.Join(gorcCfgPath, "gorc_bootstrap.sh"),
		)
		s.Require().NoError(err)

		resource, err := s.dkrPool.RunWithOptions(
			&dockertest.RunOptions{
				Name:      orch.instanceName(),
				NetworkID: s.dkrNet.Network.ID,
				Mounts: []string{
					fmt.Sprintf("%s/:/root/gorc", gorcCfgPath),
				},
				Repository: "umeenet/umeed",
				Env: []string{
					fmt.Sprintf("UMEE_E2E_ORCH_MNEMONIC=%s", orch.mnemonic),
					fmt.Sprintf("UMEE_E2E_ETH_PRIV_KEY=%s", val.ethereumKey.privateKey),
				},
				Entrypoint: []string{
					"sh",
					"-c",
					"chmod +x /root/gorc/gorc_bootstrap.sh && /root/gorc/gorc_bootstrap.sh",
				},
			},
			noRestart,
		)
		s.Require().NoError(err)

		s.orchResources[i] = resource
		s.T().Logf("started orchestrator container: %s", resource.Container.ID)
	}

	// TODO: [bez] Determine if there is a way to check the health or status of
	// the gorc orchestrator processes. For now, we search the logs to determine
	// when each orchestrator resource has seen a validator set update.
	match := "relayer::valset_relaying: Consideration: looks good"
	for _, resource := range s.orchResources {
		s.T().Logf("waiting for orchestrator to be healthy: %s", resource.Container.ID)

		s.Require().Eventuallyf(
			func() bool {
				var containerLogsBuf bytes.Buffer
				s.Require().NoError(s.dkrPool.Client.Logs(
					docker.LogsOptions{
						Container:    resource.Container.ID,
						OutputStream: &containerLogsBuf,
						Stdout:       true,
						Stderr:       true,
					},
				))

				return strings.Contains(containerLogsBuf.String(), match)
			},
			30*time.Second,
			2*time.Second,
			"orchestrator %s not healthy",
			resource.Container.ID,
		)
	}
}

func (s *IntegrationTestSuite) runGaiaNetwork() {
	s.T().Log("starting Gaia network container...")

	gaiaVal := s.chain.gaiaValidators[0]
	gaiaCfgPath := path.Join(gaiaVal.configDir(), "cfg")

	s.Require().NoError(os.MkdirAll(gaiaCfgPath, 0755))
	_, err := copyFile(
		filepath.Join("./scripts/", "gaia_bootstrap.sh"),
		filepath.Join(gaiaCfgPath, "gaia_bootstrap.sh"),
	)
	s.Require().NoError(err)

	_, err = copyFile(
		filepath.Join("./docker/", "gaia.Dockerfile"),
		filepath.Join(gaiaVal.configDir(), "gaia.Dockerfile"),
	)
	s.Require().NoError(err)

	s.gaiaResource, err = s.dkrPool.BuildAndRunWithBuildOptions(
		&dockertest.BuildOptions{
			Dockerfile: "gaia.Dockerfile",
			ContextDir: gaiaVal.configDir(),
		},
		&dockertest.RunOptions{
			Name:      gaiaVal.instanceName(),
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/root/.gaia", gaiaVal.configDir()),
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
			status, err := s.gaiaRPC.Status(context.Background())
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

	gaiaVal := s.chain.gaiaValidators[0]
	umeeVal := s.chain.validators[0]
	hermesCfgPath := path.Join(s.chain.configDir(), "hermes")

	s.Require().NoError(os.MkdirAll(hermesCfgPath, 0755))
	_, err := copyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.hermesResource, err = s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:      "umee-gaia-relayer",
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/home/hermes", hermesCfgPath),
			},
			ExposedPorts: []string{
				"3031",
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"3031/tcp": {{HostIP: "", HostPort: "3031"}},
			},
			Repository: "informalsystems/hermes",
			Tag:        "0.7.3",
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

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}
