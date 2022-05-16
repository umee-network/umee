package tests

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	stop bool // allows skipping of remaining tests once a required transaction has failed

	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")

	s.network.Cleanup()
}

type testTransaction struct {
	name        string
	command     *cobra.Command
	args        []string
	expectedErr *sdkerrors.Error
}

// runTestTransactions implements repetitive test logic where transaction commands must be run
// with flags to bypass gas and signature checks, then checked for success or an expected error
func runTestTransactions(s *IntegrationTestSuite, tcs []testTransaction) {
	if s.stop {
		return
	}

	clientCtx := s.network.Validators[0].ClientCtx

	txFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	for _, tc := range tcs {
		tc := tc // for the linter
		tc.args = append(tc.args, txFlags...)

		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.command, tc.args)
			s.Require().NoError(err)
			if err != nil {
				s.stop = true
			}

			resp := &sdk.TxResponse{}
			err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp)
			s.Require().NoError(err, out.String())
			if err != nil {
				s.stop = true
			}

			if tc.expectedErr == nil {
				s.Require().Equal(0, int(resp.Code))
				if int(resp.Code) != 0 {
					s.stop = true
				}
			} else {
				s.Require().Equal(int(tc.expectedErr.ABCICode()), int(resp.Code))
				if int(resp.Code) != int(tc.expectedErr.ABCICode()) {
					s.stop = true
				}
			}
		})
	}
}

type testQuery struct {
	name             string
	command          *cobra.Command
	args             []string
	expectErr        bool
	responseType     proto.Message
	expectedResponse proto.Message
}

// runTestQueries implements repetitive test logic where query commands must be run
// then checked for an exact response (including matching fields)
func runTestQueries(s *IntegrationTestSuite, tcs []testQuery) {
	if s.stop {
		return
	}

	clientCtx := s.network.Validators[0].ClientCtx

	queryFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	for _, tc := range tcs {
		tc := tc // for the linter
		tc.args = append(tc.args, queryFlags...)

		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.command, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.responseType), out.String())
				s.Require().Equal(tc.expectedResponse, tc.responseType)
			}
		})
	}
}
