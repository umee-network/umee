package tests

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/x/oracle/client/cli"
	"github.com/umee-network/umee/v2/x/oracle/types"
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
	t := s.T()
	t.Log("setting up integration test suite")

	var err error
	s.network, err = network.New(t, t.TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")

	s.network.Cleanup()
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
	inactiveVal := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			name: "valid request",
			args: []string{
				val.ValAddress.String(),
			},
			expectErr: false,
			respType:  &types.QueryFeederDelegationResponse{},
		},
		{
			name: "invalid address",
			args: []string{
				"invalid_address",
			},
			expectErr: true,
			respType:  &types.QueryFeederDelegationResponse{},
		},
		{
			name: "non-existent validator",
			args: []string{
				inactiveVal.String(),
			},
			expectErr: true,
			respType:  &types.QueryFeederDelegationResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			tc.args = append(tc.args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryFeederDelegation(), tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryExchangeRates() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	args := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryExchangeRates(), args)
	s.Require().NoError(err)

	var res types.QueryExchangeRatesResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))

	s.Require().Len(res.ExchangeRates, 1)
	s.Require().Equal(res.ExchangeRates[0].Denom, umeeapp.DisplayDenom)
	s.Require().False(res.ExchangeRates[0].Amount.IsZero())
}

func (s *IntegrationTestSuite) TestQueryParams() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	args := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryParams(), args)
	s.Require().NoError(err)

	var res types.QueryParamsResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))

	s.Require().NotEmpty(res.Params.AcceptList)
}

func (s *IntegrationTestSuite) TestQueryExchangeRate() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			name: "valid denom",
			args: []string{
				"UMEE",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			expectErr: false,
			respType:  &types.QueryExchangeRatesResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryExchangeRate(), tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}
