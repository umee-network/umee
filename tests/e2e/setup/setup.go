package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
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

	"github.com/umee-network/umee/v6/app"
	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/client"
	"github.com/umee-network/umee/v6/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v6/x/oracle/types"
	"github.com/umee-network/umee/v6/x/uibc"
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

	// codec
	s.cdc = encodingConfig.Codec

	s.Chain, err = newChain()
	s.Require().NoError(err)

	s.T().Logf("starting e2e infrastructure; chain-id: %s; datadir: %s", s.Chain.ID, s.Chain.dataDir)

	s.DkrPool, err = dockertest.NewPool("")
	s.Require().NoError(err)

	s.DkrNet, err = s.DkrPool.CreateNetwork(fmt.Sprintf("%s-testnet", s.Chain.ID))
	s.Require().NoError(err)

	s.initNodes()            // create umee validator nodes and their config.toml, genesis.json
	s.initGenesis()          // modify genesis file, add gentxs, and save to each validator
	s.initValidatorConfigs() // modify config.toml and app.toml for each validator
	s.runValidators()
	if !s.MinNetwork {
		s.runPriceFeeder(2) // index of the validator voting on prices
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

// initGenesis modifies the genesis.json in s.Chain.Validators[0]'s config directory
// to include new initial parameters and state for multiple modules, then adds the
// gentxs from all validators, and overwrites genesis.json on each validator
// with the new file.
func (s *E2ETestSuite) initGenesis() {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(s.Chain.Validators[0].configDir())
	config.Moniker = s.Chain.Validators[0].moniker

	genFilePath := config.GenesisFile()
	s.T().Log("modifying genesis file; validator_0 config:", genFilePath)
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	s.Require().NoError(err)

	var bech32GenState bech32ibctypes.GenesisState
	s.Require().NoError(s.cdc.UnmarshalJSON(appGenState[bech32ibctypes.ModuleName], &bech32GenState))

	// bech32
	bech32GenState.NativeHRP = sdk.GetConfig().GetBech32AccountAddrPrefix()

	bz, err := s.cdc.MarshalJSON(&bech32GenState)
	s.Require().NoError(err)
	appGenState[bech32ibctypes.ModuleName] = bz

	// Leverage
	var leverageGenState leveragetypes.GenesisState
	s.Require().NoError(s.cdc.UnmarshalJSON(appGenState[leveragetypes.ModuleName], &leverageGenState))

	leverageGenState.Registry = append(leverageGenState.Registry,
		fixtures.Token(appparams.BondDenom, appparams.DisplayDenom, 6),
		fixtures.Token(ATOMBaseDenom, ATOM, uint32(ATOMExponent)),
	)

	bz, err = s.cdc.MarshalJSON(&leverageGenState)
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
	govGenState.VotingParams.VotingPeriod = &votingPeriod
	govGenState.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, sdk.NewInt(100)))

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
	s.T().Log("creating genesis txs")
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
	s.T().Log("writing updated genesis file to each validator")
	for _, val := range s.Chain.Validators {
		writeFile(filepath.Join(val.configDir(), "config", "genesis.json"), bz)
	}
}

// initValidatorConfigs modifies config.toml and app.toml for all s.Chain.Validators
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
		appConfig.MinGasPrices = minGasPrice
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

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}
