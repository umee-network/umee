package tests

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	"github.com/umee-network/umee/x/oracle/client/cli"
	"github.com/umee-network/umee/x/oracle/types"
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

	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestDelegateFeedConsent() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			name: "invalid operator address",
			args: []string{
				"foo",
				s.network.Validators[1].Address.String(),
			},
			expectErr: true,
			respType:  &sdk.TxResponse{},
		},
		{
			name: "invalid feeder address",
			args: []string{
				val.Address.String(),
				"foo",
			},
			expectErr: true,
			respType:  &sdk.TxResponse{},
		},
		{
			name: "valid transaction",
			args: []string{
				val.Address.String(),
				s.network.Validators[1].Address.String(),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdDelegateFeedConsent(), tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryFeedDelegate() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	args := []string{
		val.ValAddress.String(),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryFeederDelegation(), args)
	s.Require().NoError(err)

	var resp types.QueryFeederDelegationResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
}
