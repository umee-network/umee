package tests

import (
	"fmt"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	network, err := network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)
	s.network = network

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")

	s.network.Cleanup()
}

// runTestQuery
func (s *IntegrationTestSuite) runTestQueries(tqs ...testQuery) {
	for _, t := range tqs {
		t.Run(s)
	}
}

// runTestCases runs test transactions or queries, stopping early if an error occurs
func (s *IntegrationTestSuite) runTestTransactions(txs ...testTransaction) {
	for _, t := range txs {
		t.Run(s)
	}
}

type testTransaction struct {
	msg         string
	command     *cobra.Command
	args        []string
	expectedErr *errors.Error
}

type testQuery struct {
	msg              string
	command          *cobra.Command
	args             []string
	expectErr        bool
	responseType     proto.Message
	expectedResponse proto.Message
}

func (t testTransaction) Run(s *IntegrationTestSuite) {
	clientCtx := s.network.Validators[0].ClientCtx

	txFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.network.Validators[0].Address),
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	t.args = append(t.args, txFlags...)

	out, err := clitestutil.ExecTestCLICmd(clientCtx, t.command, t.args)
	s.Require().NoError(err, t.msg)

	resp := &sdk.TxResponse{}
	err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp)
	s.Require().NoError(err, t.msg)

	if t.expectedErr == nil {
		s.Require().Equal(0, int(resp.Code), t.msg)
	} else {
		s.Require().Equal(int(t.expectedErr.ABCICode()), int(resp.Code), t.msg)
	}
}

func (t testQuery) Run(s *IntegrationTestSuite) {
	clientCtx := s.network.Validators[0].ClientCtx

	queryFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	t.args = append(t.args, queryFlags...)

	out, err := clitestutil.ExecTestCLICmd(clientCtx, t.command, t.args)

	if t.expectErr {
		s.Require().Error(err, t.msg)
	} else {
		s.Require().NoError(err, t.msg)

		err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), t.responseType)
		s.Require().NoError(err, t.msg)

		s.Require().Equal(t.expectedResponse.String(), t.responseType.String())
	}
}
