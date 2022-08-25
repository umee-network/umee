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

	abort bool // stop interdependent tests on the first error for clarity

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

// TestCases are queries and transactions that can be run, and return a boolean
// which indicates to abort the test suite if true
type TestCase interface {
	Run(s *IntegrationTestSuite) bool
}

// runTestCases runs test transactions or queries, stopping early if an error occurs
func (s *IntegrationTestSuite) runTestCases(tcs ...TestCase) {
	for _, t := range tcs {
		if !s.abort {
			s.abort = t.Run(s)
		}
	}
}

type testTransaction struct {
	name        string
	command     *cobra.Command
	args        []string
	expectedErr *errors.Error
}

type testQuery struct {
	name             string
	command          *cobra.Command
	args             []string
	expectErr        bool
	responseType     proto.Message
	expectedResponse proto.Message
}

func (t testTransaction) Run(s *IntegrationTestSuite) (abort bool) {
	clientCtx := s.network.Validators[0].ClientCtx

	txFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	t.args = append(t.args, txFlags...)

	s.Run(t.name, func() {
		out, err := clitestutil.ExecTestCLICmd(clientCtx, t.command, t.args)
		s.Require().NoError(err)
		if err != nil {
			abort = true
		}

		resp := &sdk.TxResponse{}
		err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp)
		s.Require().NoError(err, out.String())
		if err != nil {
			abort = true
		}

		if t.expectedErr == nil {
			s.Require().Equal(0, int(resp.Code), "events %v", resp.Events)
			if int(resp.Code) != 0 {
				abort = true
			}
		} else {
			s.Require().Equal(int(t.expectedErr.ABCICode()), int(resp.Code))
			if int(resp.Code) != int(t.expectedErr.ABCICode()) {
				abort = true
			}
		}
	})
	return abort
}

func (t testQuery) Run(s *IntegrationTestSuite) (abort bool) {
	clientCtx := s.network.Validators[0].ClientCtx

	queryFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	t.args = append(t.args, queryFlags...)

	s.Run(t.name, func() {
		out, err := clitestutil.ExecTestCLICmd(clientCtx, t.command, t.args)

		if t.expectErr {
			s.Require().Error(err)
			if err == nil {
				abort = true
			}
		} else {
			s.Require().NoError(err)
			if err != nil {
				abort = true
			}

			err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), t.responseType)
			s.Require().NoError(err, out.String())
			if err != nil {
				abort = true
			}

			s.Require().Equal(t.expectedResponse, t.responseType)
			if !s.Assert().Equal(t.expectedResponse, t.responseType) {
				abort = true
			}
		}
	})
	return abort
}
