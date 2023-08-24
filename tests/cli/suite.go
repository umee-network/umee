package itest

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/umee-network/umee/v6/x/uibc"

	"github.com/stretchr/testify/require"

	"github.com/ory/dockertest/v3/docker"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/cosmos/cosmos-sdk/server"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v6/x/oracle/types"

	"github.com/ory/dockertest/v3"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/client"
	"github.com/umee-network/umee/v6/tests/util"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

var (
	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics module.BasicManager
)

type CLISuite struct {
	T            *testing.T
	Require      *require.Assertions
	Chain        *util.Chain
	DkrPool      *dockertest.Pool
	DkrNet       *dockertest.Network
	ValResources []*dockertest.Resource
	Umee         client.Client
	cdc          codec.Codec
}

func (s *CLISuite) SetupSuite() {
	var err error
	s.T.Log("setting up cli test suite")
	s.Require = require.New(s.T)

	s.cdc = util.EncodingConfig.Codec

	s.Chain, err = util.NewChain()
	s.Require.NoError(err)

	s.DkrPool, err = dockertest.NewPool("")
	s.Require.NoError(err)

	s.DkrNet, err = s.DkrPool.CreateNetwork(fmt.Sprintf("%s-testnet", s.Chain.ID))
	s.Require.NoError(err)

	s.Require.NoError(s.Chain.InitNodes(s.cdc, 1))

	s.initGenesis()

	s.Require.NoError(s.Chain.InitValidatorConfigs())
	s.runValidators()
}

func (s *CLISuite) initGenesis() {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(s.Chain.Validators[0].ConfigDir())
	config.Moniker = s.Chain.Validators[0].Moniker

	genFilePath := config.GenesisFile()
	s.T.Log("starting e2e infrastructure; validator_0 config:", genFilePath)
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	s.Require.NoError(err)

	// Override default leverage registry with one more suitable for testing
	var leverageGenState leveragetypes.GenesisState
	if err := s.cdc.UnmarshalJSON(appGenState[leveragetypes.ModuleName], &leverageGenState); err != nil {
		panic(err)
	}
	leverageGenState.Registry = []leveragetypes.Token{
		fixtures.Token(params.BondDenom, params.DisplayDenom, 6),
	}

	bz, err := s.cdc.MarshalJSON(&leverageGenState)
	if err != nil {
		panic(err)
	}
	appGenState[leveragetypes.ModuleName] = bz

	var oracleGenState oracletypes.GenesisState
	if err := s.cdc.UnmarshalJSON(appGenState[oracletypes.ModuleName], &oracleGenState); err != nil {
		panic(err)
	}

	// Set mock exchange rates and a large enough vote period such that we won't
	// execute ballot voting and thus clear out previous exchange rates, since we
	// are not running a price-feeder.
	oracleGenState.Params.VotePeriod = 1000
	oracleGenState.ExchangeRates = append(
		oracleGenState.ExchangeRates, oracletypes.NewExchangeRateTuple(
			params.DisplayDenom, sdk.MustNewDecFromStr("34.21"),
		),
	)
	// Set mock historic medians to satisfy leverage module's 24 median requirement
	for i := 1; i <= 24; i++ {
		median := oracletypes.Price{
			ExchangeRateTuple: oracletypes.NewExchangeRateTuple(
				params.DisplayDenom,
				sdk.MustNewDecFromStr("34.21"),
			),
			BlockNum: uint64(i),
		}
		oracleGenState.Medians = append(oracleGenState.Medians, median)
	}

	bz, err = s.cdc.MarshalJSON(&oracleGenState)
	if err != nil {
		panic(err)
	}
	appGenState[oracletypes.ModuleName] = bz

	var govGenState govv1.GenesisState
	if err := s.cdc.UnmarshalJSON(appGenState[govtypes.ModuleName], &govGenState); err != nil {
		panic(err)
	}

	votingPeriod := time.Minute
	govGenState.VotingParams.VotingPeriod = &votingPeriod

	bz, err = s.cdc.MarshalJSON(&govGenState)
	if err != nil {
		panic(err)
	}
	appGenState[govtypes.ModuleName] = bz

	var uibcGenState uibc.GenesisState
	assert.NilError(s.T, s.cdc.UnmarshalJSON(appGenState[uibc.ModuleName], &uibcGenState))
	uibcGenState.Outflows = sdk.DecCoins{sdk.NewInt64DecCoin("uumee", 0)}
	uibcGenState.TotalOutflowSum = sdk.NewDec(10)

	bz, err = s.cdc.MarshalJSON(&uibcGenState)
	assert.NilError(s.T, err)
	appGenState[uibc.ModuleName] = bz

	bz, err = json.MarshalIndent(appGenState, "", "  ")
	s.Require.NoError(err)

	genDoc.AppState = bz

	bz, err = tmjson.MarshalIndent(genDoc, "", "  ")
	s.Require.NoError(err)

	// write the updated genesis file to each validator
	for _, val := range s.Chain.Validators {
		err = util.WriteFile(filepath.Join(val.ConfigDir(), "config", "genesis.json"), bz)
		if err != nil {
			panic(err)
		}
	}
}

func (s *CLISuite) runValidators() {
	s.T.Log("starting Umee validator containers...")

	s.ValResources = make([]*dockertest.Resource, len(s.Chain.Validators))
	for i, val := range s.Chain.Validators {
		runOpts := &dockertest.RunOptions{
			Name:      val.InstanceName(),
			NetworkID: s.DkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/root/.umee", val.ConfigDir()),
			},
			Repository: "umee-network/umeed-e2e",
		}

		// expose the first validator for debugging and communication
		if val.Index == 0 {
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

		resource, err := s.DkrPool.RunWithOptions(runOpts, util.NoRestart)
		s.Require.NoError(err)

		s.ValResources[i] = resource
		s.T.Logf("started Umee validator container: %s", resource.Container.ID)
	}

	rpcClient, err := rpchttp.New("tcp://localhost:26657", "/websocket")
	s.Require.NoError(err)

	s.Require.Eventually(
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

func (s *CLISuite) TearDownSuite() {
	s.T.Log("tearing down cli test suite")

	for _, vc := range s.ValResources {
		s.Require.NoError(s.DkrPool.Purge(vc))
	}

	s.Require.NoError(s.DkrPool.RemoveNetwork(s.DkrNet))
}

// runTestQuery
func (s *CLISuite) RunTestQueries(tqs ...TestQuery) {
	for _, tq := range tqs {
		// since steps of this test suite depend on previous transactions, we want to stop
		// on the first failure, rather than continue producing potentially inaccurate
		// errors as an effect of the first.
		// t.Run(s) stops properly, whereas s.Run would not
		s.T.Run(
			tq.Name, func(t *testing.T) {
				s.RunQuery(tq)
			},
		)
	}
}

func (s *CLISuite) RunQuery(tq TestQuery) {
	clientCtx := s.Chain.Validators[0].ClientCtx
	queryFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	out, err := clitestutil.ExecTestCLICmd(clientCtx, tq.Command, append(tq.Args, queryFlags...))

	if len(tq.ErrMsg) != 0 {
		assert.ErrorContains(s.T, err, tq.ErrMsg)
	} else {
		assert.NilError(s.T, err, tq.Name)

		err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), tq.Response)
		assert.NilError(s.T, err, tq.Name)
		assert.Equal(s.T, tq.ExpectedResponse.String(), tq.Response.String(), tq.Name)
	}
}

// runTestCases runs test transactions or queries, stopping early if an error occurs
func (s *CLISuite) RunTestTransactions(txs ...TestTransaction) {
	for _, tx := range txs {
		// since steps of this test suite depend on previous transactions, we want to stop
		// on the first failure, rather than continue producing potentially inaccurate
		// errors as an effect of the first.
		// t.Run(s) stops properly, whereas s.Run would not
		s.T.Run(
			tx.Name, func(t *testing.T) {
				s.RunTransaction(tx)
			},
		)
	}
}

func (s *CLISuite) RunTransaction(tx TestTransaction) {
	clientCtx := s.Chain.Validators[0].ClientCtx
	addr, _ := s.Chain.Validators[0].KeyInfo.GetAddress()
	txFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, addr),
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "10000000"),
		fmt.Sprintf("--%s=%s", flags.FlagFees, "1000000uumee"),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, tx.Command, append(tx.Args, txFlags...))
	assert.NilError(s.T, err, tx.Name)

	resp := &sdk.TxResponse{}
	err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp)
	assert.NilError(s.T, err, tx.Name)

	if tx.ExpectedErr == nil {
		assert.Equal(s.T, 0, int(resp.Code), "msg: %s\nresp: %s", tx.Name, resp)
	} else {
		assert.Equal(s.T, int(tx.ExpectedErr.ABCICode()), int(resp.Code), tx.Name)
	}
}
