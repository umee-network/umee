package setup

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
	"time"

	tmconfig "github.com/cometbft/cometbft/config"
	tmjson "github.com/cometbft/cometbft/libs/json"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/umee-network/umee/v5/app"
	appparams "github.com/umee-network/umee/v5/app/params"
	"github.com/umee-network/umee/v5/client"
	"github.com/umee-network/umee/v5/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v5/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v5/x/oracle/types"
	"github.com/umee-network/umee/v5/x/uibc"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type E2ETestSuite struct {
	suite.Suite

	tmpDirs             []string
	Chain               *chain
	gaiaRPC             *rpchttp.HTTP
	DkrPool             *dockertest.Pool
	DkrNet              *dockertest.Network
	GaiaResource        *dockertest.Resource
	HermesResource      *dockertest.Resource
	priceFeederResource *dockertest.Resource
	ValResources        []*dockertest.Resource
	Umee                client.Client
	cdc                 codec.Codec
	MinNetwork          bool // MinNetwork defines which runs only validator wihtout price-feeder, gaia and ibc-relayer
}

func (s *E2ETestSuite) SetupSuite() {
	var err error
	s.T().Log("setting up e2e integration test suite...")

	db := dbm.NewMemDB()
	app := app.New(
		nil,
		db,
		nil,
		true,
		map[int64]bool{},
		"",
		0,
		app.EmptyAppOptions{},
		nil,
		nil,
	)
	encodingConfig = testutil.TestEncodingConfig{
		InterfaceRegistry: app.InterfaceRegistry(),
		Codec:             app.AppCodec(),
		TxConfig:          app.GetTxConfig(),
		Amino:             app.LegacyAmino(),
	}

	// codec
	s.cdc = encodingConfig.Codec

	s.Chain, err = newChain()
	s.Require().NoError(err)

	s.T().Logf("starting e2e infrastructure; chain-id: %s; datadir: %s", s.Chain.ID, s.Chain.dataDir)

	s.DkrPool, err = dockertest.NewPool("")
	s.Require().NoError(err)

	s.DkrNet, err = s.DkrPool.CreateNetwork(fmt.Sprintf("%s-testnet", s.Chain.ID))
	s.Require().NoError(err)

	s.initNodes() // init validator nodes
	s.initGenesis()
	s.initValidatorConfigs()
	s.runValidators()
	if !s.MinNetwork {
		s.runPriceFeeder()
		s.runGaiaNetwork()
		time.Sleep(3 * time.Second) // wait for gaia to start
		s.runIBCRelayer()
	} else {
		s.T().Log("running minimum network withut gaia,price-feeder and ibc-relayer")
	}
	s.initUmeeClient()
}

func (s *E2ETestSuite) TearDownSuite() {
	if str := os.Getenv("UMEE_E2E_SKIP_CLEANUP"); len(str) > 0 {
		skipCleanup, err := strconv.ParseBool(str)
		s.Require().NoError(err)

		if skipCleanup {
			return
		}
	}

	s.T().Log("tearing down e2e integration test suite...")

	if !s.MinNetwork {
		s.Require().NoError(s.DkrPool.Purge(s.GaiaResource))
		s.Require().NoError(s.DkrPool.Purge(s.HermesResource))
		s.Require().NoError(s.DkrPool.Purge(s.priceFeederResource))
	}

	for _, vc := range s.ValResources {
		s.Require().NoError(s.DkrPool.Purge(vc))
	}

	s.Require().NoError(s.DkrPool.RemoveNetwork(s.DkrNet))

	os.RemoveAll(s.Chain.dataDir)
	for _, td := range s.tmpDirs {
		os.RemoveAll(td)
	}
}

func (s *E2ETestSuite) initNodes() {
	s.Require().NoError(s.Chain.createAndInitValidators(s.cdc, 3))
	s.Require().NoError(s.Chain.createAndInitGaiaValidator(s.cdc))

	// initialize a genesis file for the first validator
	val0ConfigDir := s.Chain.Validators[0].configDir()
	for _, val := range s.Chain.Validators {
		valAddr, err := val.KeyInfo.GetAddress()
		s.Require().NoError(err)
		s.Require().NoError(
			addGenesisAccount(s.cdc, val0ConfigDir, "", InitBalanceStr, valAddr),
		)
	}

	// copy the genesis file to the remaining validators
	for _, val := range s.Chain.Validators[1:] {
		_, err := copyFile(
			filepath.Join(val0ConfigDir, "config", "genesis.json"),
			filepath.Join(val.configDir(), "config", "genesis.json"),
		)
		s.Require().NoError(err)
	}
}

func (s *E2ETestSuite) initGenesis() {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(s.Chain.Validators[0].configDir())
	config.Moniker = s.Chain.Validators[0].moniker

	genFilePath := config.GenesisFile()
	s.T().Log("starting e2e infrastructure; validator_0 config:", genFilePath)
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	s.Require().NoError(err)

	// Leverage
	var leverageGenState leveragetypes.GenesisState
	s.Require().NoError(s.cdc.UnmarshalJSON(appGenState[leveragetypes.ModuleName], &leverageGenState))

	leverageGenState.Registry = append(leverageGenState.Registry,
		fixtures.Token(appparams.BondDenom, appparams.DisplayDenom, 6),
		fixtures.Token(ATOMBaseDenom, ATOM, uint32(ATOMExponent)),
	)

	bz, err := s.cdc.MarshalJSON(&leverageGenState)
	s.Require().NoError(err)
	appGenState[leveragetypes.ModuleName] = bz

	// Oracle
	var oracleGenState oracletypes.GenesisState
	s.Require().NoError(s.cdc.UnmarshalJSON(appGenState[oracletypes.ModuleName], &oracleGenState))

	oracleGenState.Params.HistoricStampPeriod = 5
	oracleGenState.Params.MaximumPriceStamps = 4
	oracleGenState.Params.MedianStampPeriod = 20
	oracleGenState.Params.MaximumMedianStamps = 2

	oracleGenState.Params.AcceptList = append(oracleGenState.Params.AcceptList, oracletypes.Denom{
		BaseDenom:   ATOMBaseDenom,
		SymbolDenom: ATOM,
		Exponent:    uint32(ATOMExponent),
	})

	bz, err = s.cdc.MarshalJSON(&oracleGenState)
	s.Require().NoError(err)
	appGenState[oracletypes.ModuleName] = bz

	// Gov
	var govGenState govtypesv1.GenesisState
	s.Require().NoError(s.cdc.UnmarshalJSON(appGenState[govtypes.ModuleName], &govGenState))

	votingPeriod := 5 * time.Second
	govGenState.Params.VotingPeriod = &votingPeriod
	govGenState.Params.MinDeposit = sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, sdk.NewInt(100)))

	bz, err = s.cdc.MarshalJSON(&govGenState)
	s.Require().NoError(err)
	appGenState[govtypes.ModuleName] = bz

	// Bank
	var bankGenState banktypes.GenesisState
	s.Require().NoError(s.cdc.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState))

	bankGenState.DenomMetadata = append(bankGenState.DenomMetadata, banktypes.Metadata{
		Description: "An example stable token",
		Display:     PhotonDenom,
		Base:        PhotonDenom,
		Symbol:      PhotonDenom,
		Name:        PhotonDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    PhotonDenom,
				Exponent: 0,
			},
		},
	})

	bz, err = s.cdc.MarshalJSON(&bankGenState)
	s.Require().NoError(err)
	appGenState[banktypes.ModuleName] = bz

	// uibc (ibc quota)
	var uibcGenState uibc.GenesisState
	s.Require().NoError(s.cdc.UnmarshalJSON(appGenState[uibc.ModuleName], &uibcGenState))

	// 100$ for each token
	uibcGenState.Params.TokenQuota = sdk.NewDec(100)
	// 120$ for all tokens on quota duration
	uibcGenState.Params.TotalQuota = sdk.NewDec(120)
	// quotas will reset every 300 seconds
	uibcGenState.Params.QuotaDuration = time.Second * 300

	bz, err = s.cdc.MarshalJSON(&uibcGenState)
	s.Require().NoError(err)
	appGenState[uibc.ModuleName] = bz

	var genUtilGenState genutiltypes.GenesisState
	s.Require().NoError(s.cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState))

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(s.Chain.Validators))
	for i, val := range s.Chain.Validators {
		var createValmsg sdk.Msg
		if i == 2 {
			createValmsg, err = val.buildCreateValidatorMsg(stakeAmountCoin2)
		} else {
			createValmsg, err = val.buildCreateValidatorMsg(stakeAmountCoin)
		}
		s.Require().NoError(err)

		signedTx, err := val.signMsg(s.cdc, createValmsg)
		s.Require().NoError(err)

		txRaw, err := s.cdc.MarshalJSON(signedTx)
		s.Require().NoError(err)

		genTxs[i] = txRaw
	}

	genUtilGenState.GenTxs = genTxs

	bz, err = s.cdc.MarshalJSON(&genUtilGenState)
	s.Require().NoError(err)
	appGenState[genutiltypes.ModuleName] = bz

	bz, err = json.MarshalIndent(appGenState, "", "  ")
	s.Require().NoError(err)

	genDoc.AppState = bz

	bz, err = tmjson.MarshalIndent(genDoc, "", "  ")
	s.Require().NoError(err)

	// write the updated genesis file to each validator
	for _, val := range s.Chain.Validators {
		writeFile(filepath.Join(val.configDir(), "config", "genesis.json"), bz)
	}
}

func (s *E2ETestSuite) initValidatorConfigs() {
	for i, val := range s.Chain.Validators {
		tmCfgPath := filepath.Join(val.configDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(tmCfgPath)
		s.Require().NoError(vpr.ReadInConfig())

		valConfig := tmconfig.DefaultConfig()
		valConfig.Consensus.SkipTimeoutCommit = true
		s.Require().NoError(vpr.Unmarshal(valConfig))

		valConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
		valConfig.P2P.AddrBookStrict = false
		valConfig.P2P.ExternalAddress = fmt.Sprintf("%s:%d", val.instanceName(), 26656)
		valConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		valConfig.StateSync.Enable = false
		valConfig.LogLevel = "info"

		var peers []string

		for j := 0; j < len(s.Chain.Validators); j++ {
			if i == j {
				continue
			}

			peer := s.Chain.Validators[j]
			peerID := fmt.Sprintf("%s@%s%d:26656", peer.nodeKey.ID(), peer.moniker, j)
			peers = append(peers, peerID)
		}

		valConfig.P2P.PersistentPeers = strings.Join(peers, ",")

		tmconfig.WriteConfigFile(tmCfgPath, valConfig)

		// set application configuration
		appCfgPath := filepath.Join(val.configDir(), "config", "app.toml")

		appConfig := srvconfig.DefaultConfig()
		appConfig.API.Enable = true
		appConfig.API.Address = "tcp://0.0.0.0:1317"
		appConfig.MinGasPrices = minGasPrice
		appConfig.GRPC.Address = "0.0.0.0:9090"
		appConfig.Pruning = "nothing"

		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}
}

func (s *E2ETestSuite) runValidators() {
	s.T().Log("starting Umee validator containers...")

	s.ValResources = make([]*dockertest.Resource, len(s.Chain.Validators))
	for i, val := range s.Chain.Validators {
		runOpts := &dockertest.RunOptions{
			Name:      val.instanceName(),
			NetworkID: s.DkrNet.Network.ID,
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

		resource, err := s.DkrPool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		s.ValResources[i] = resource
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

func (s *E2ETestSuite) runGaiaNetwork() {
	s.T().Log("starting Gaia network container...")

	tmpDir, err := os.MkdirTemp("", "umee-e2e-testnet-gaia-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	gaiaVal := s.Chain.GaiaValidators[0]

	gaiaCfgPath := path.Join(tmpDir, "cfg")
	s.Require().NoError(os.MkdirAll(gaiaCfgPath, 0o755))

	_, err = copyFile(
		filepath.Join("./scripts/", "gaia_bootstrap.sh"),
		filepath.Join(gaiaCfgPath, "gaia_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.GaiaResource, err = s.DkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       gaiaVal.instanceName(),
			Repository: "ghcr.io/umee-network/gaia-e2e",
			Tag:        "latest",
			NetworkID:  s.DkrNet.Network.ID,
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
				fmt.Sprintf("UMEE_E2E_GAIA_CHAIN_ID=%s", GaiaChainID),
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

	endpoint := fmt.Sprintf("tcp://%s", s.GaiaResource.GetHostPort("26657/tcp"))
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

	s.T().Logf("started Gaia network container: %s", s.GaiaResource.Container.ID)
}

func (s *E2ETestSuite) runIBCRelayer() {
	s.T().Log("starting Hermes relayer container...")

	tmpDir, err := os.MkdirTemp("", "umee-e2e-testnet-hermes-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	gaiaVal := s.Chain.GaiaValidators[0]
	// umeeVal for the relayer needs to be a different account
	// than what we use for runPriceFeeder.
	umeeVal := s.Chain.Validators[1]
	hermesCfgPath := path.Join(tmpDir, "hermes")

	s.Require().NoError(os.MkdirAll(hermesCfgPath, 0o755))
	_, err = copyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.HermesResource, err = s.DkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       "umee-gaia-relayer",
			Repository: "ghcr.io/umee-network/hermes-e2e", //"informalsystems/hermes", //
			Tag:        "latest",
			NetworkID:  s.DkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/home/hermes", hermesCfgPath),
			},
			ExposedPorts: []string{"3031"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"3031/tcp": {{HostIP: "", HostPort: "3031"}},
			},
			Env: []string{
				fmt.Sprintf("UMEE_E2E_GAIA_CHAIN_ID=%s", GaiaChainID),
				fmt.Sprintf("UMEE_E2E_UMEE_CHAIN_ID=%s", s.Chain.ID),
				fmt.Sprintf("UMEE_E2E_GAIA_VAL_MNEMONIC=%s", gaiaVal.mnemonic),
				fmt.Sprintf("UMEE_E2E_UMEE_VAL_MNEMONIC=%s", umeeVal.mnemonic),
				fmt.Sprintf("UMEE_E2E_GAIA_VAL_HOST=%s", s.GaiaResource.Container.Name[1:]),
				fmt.Sprintf("UMEE_E2E_UMEE_VAL_HOST=%s", s.ValResources[0].Container.Name[1:]),
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

	endpoint := fmt.Sprintf("http://%s/state", s.HermesResource.GetHostPort("3031/tcp"))
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

	s.T().Logf("started Hermes relayer container: %s", s.HermesResource.Container.ID)

	// create the client, connection and channel between the Umee and Gaia chains
	s.connectIBCChains()
}

func (s *E2ETestSuite) runPriceFeeder() {
	s.T().Log("starting price-feeder container...")

	umeeVal := s.Chain.Validators[2]
	umeeValAddr, err := umeeVal.KeyInfo.GetAddress()
	s.Require().NoError(err)

	grpcEndpoint := fmt.Sprintf("tcp://%s:%s", umeeVal.instanceName(), "9090")
	tmrpcEndpoint := fmt.Sprintf("http://%s:%s", umeeVal.instanceName(), "26657")

	s.priceFeederResource, err = s.DkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       "umee-price-feeder",
			NetworkID:  s.DkrNet.Network.ID,
			Repository: PriceFeederContainerRepo,
			Mounts: []string{
				fmt.Sprintf("%s/:/root/.umee", umeeVal.configDir()),
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				PriceFeederServerPort: {{HostIP: "", HostPort: "7171"}},
			},
			Env: []string{
				fmt.Sprintf("PRICE_FEEDER_PASS=%s", keyringPassphrase),
				fmt.Sprintf("ACCOUNT_ADDRESS=%s", umeeValAddr),
				fmt.Sprintf("ACCOUNT_VALIDATOR=%s", sdk.ValAddress(umeeValAddr)),
				fmt.Sprintf("KEYRING_DIR=%s", "/root/.umee"),
				fmt.Sprintf("ACCOUNT_CHAIN_ID=%s", s.Chain.ID),
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

	endpoint := fmt.Sprintf("http://%s/api/v1/prices", s.priceFeederResource.GetHostPort(PriceFeederServerPort))

	checkHealth := func() bool {
		resp, err := http.Get(endpoint)
		if err != nil {
			s.T().Log("Price feeder endpoint not available", err, endpoint)
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
	for i := 0; i < PriceFeederMaxStartupTime; i++ {
		isHealthy = checkHealth()
		if isHealthy {
			break
		}
		time.Sleep(time.Second)
	}

	if !isHealthy {
		err := s.DkrPool.Client.Logs(docker.LogsOptions{
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

func (s *E2ETestSuite) initUmeeClient() {
	var err error
	mnemonics := make(map[string]string)
	for index, v := range s.Chain.Validators {
		mnemonics[fmt.Sprintf("val%d", index)] = v.mnemonic
	}
	ecfg := app.MakeEncodingConfig()
	s.Umee, err = client.NewClient(
		s.Chain.dataDir,
		s.Chain.ID,
		"tcp://localhost:26657",
		"tcp://localhost:9090",
		mnemonics,
		1,
		ecfg,
	)
	s.Require().NoError(err)
}

func (s *E2ETestSuite) connectIBCChains() {
	s.T().Logf("connecting %s and %s chains via IBC", s.Chain.ID, GaiaChainID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.DkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.HermesResource.Container.ID,
		User:         "root",
		Cmd: []string{
			"hermes",
			"create",
			"channel",
			"--a-chain=" + s.Chain.ID,  // chain-PXbLU5
			"--b-chain=" + GaiaChainID, // test-gaia-chain
			"--a-port=transfer",
			"--b-port=transfer",
			"--new-client-connection",
			"--yes",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.DkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(
		err,
		"failed connect chains; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Containsf(
		errBuf.String(),
		"connection handshake already finished for Connection",
		"failed to connect chains via IBC: %s", errBuf.String(),
	)

	s.T().Logf("connected %s and %s chains via IBC", s.Chain.ID, GaiaChainID)
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}
