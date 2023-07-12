package itest

import (
	"fmt"
	"testing"

	tmcli "github.com/cometbft/cometbft/libs/cli"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type E2ESuite struct {
	T       *testing.T
	Cfg     network.Config
	Network *network.Network
}

func (s *E2ESuite) SetupSuite() {
	s.T.Log("setting up integration test suite")

	network, err := network.New(s.T, s.T.TempDir(), s.Cfg)
	assert.NilError(s.T, err)
	s.Network = network

	_, err = s.Network.WaitForHeight(1)
	assert.NilError(s.T, err)
}

func (s *E2ESuite) TearDownSuite() {
	s.T.Log("tearing down integration test suite")

	s.Network.Cleanup()
}

// runTestQuery
func (s *E2ESuite) RunTestQueries(tqs ...TestQuery) {
	for _, tq := range tqs {
		// since steps of this test suite depend on previous transactions, we want to stop
		// on the first failure, rather than continue producing potentially inaccurate
		// errors as an effect of the first.
		// t.Run(s) stops properly, whereas s.Run would not
		s.T.Run(tq.Name, func(t *testing.T) {
			s.RunQuery(tq)
		})
	}
}

func (s *E2ESuite) RunQuery(tq TestQuery) {
	clientCtx := s.Network.Validators[0].ClientCtx
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
func (s *E2ESuite) RunTestTransactions(txs ...TestTransaction) {
	for _, tx := range txs {
		// since steps of this test suite depend on previous transactions, we want to stop
		// on the first failure, rather than continue producing potentially inaccurate
		// errors as an effect of the first.
		// t.Run(s) stops properly, whereas s.Run would not
		s.T.Run(tx.Name, func(t *testing.T) {
			s.RunTransaction(tx)
		})
	}
}

func (s *E2ESuite) RunTransaction(tx TestTransaction) {
	clientCtx := s.Network.Validators[0].ClientCtx
	txFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.Network.Validators[0].Address),
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
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
