package tests

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/umee-network/umee/x/leverage/client/cli"
	"github.com/umee-network/umee/x/leverage/types"
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

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")

	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestQueryAllRegisteredTokens() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	flags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryAllRegisteredTokens(), flags)
	s.Require().NoError(err)

	var resp types.QueryRegisteredTokensResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
}

func (s *IntegrationTestSuite) TestQueryParams() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	flags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryParams(), flags)
	s.Require().NoError(err)

	var resp types.QueryParamsResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
}

func (s *IntegrationTestSuite) TestQueryBorrowed() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	s.Run("get_all_borrowed", func() {
		// TODO: We need to setup borrowing first prior to testing this out.
		//
		// Ref: https://github.com/umee-network/umee/issues/94
		flags := []string{
			val.Address.String(),
			fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		}

		out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryBorrowed(), flags)
		s.Require().NoError(err)

		var resp types.QueryBorrowedResponse
		s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
	})

	s.Run("get_denom_borrowed", func() {
		// TODO: We need to setup borrowing first prior to testing this out.
		//
		// Ref: https://github.com/umee-network/umee/issues/94
		flags := []string{
			val.Address.String(),
			fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			fmt.Sprintf("--%s=uumee", cli.FlagDenom),
		}

		out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryBorrowed(), flags)
		s.Require().NoError(err)

		var resp types.QueryBorrowedResponse
		s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
	})
}

func (s *IntegrationTestSuite) TestQueryReserveAmount() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	// TODO: We need to setup borrowing first prior to testing this out.
	//
	// Ref: https://github.com/umee-network/umee/issues/94
	flags := []string{
		"uumee",
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryReserveAmount(), flags)
	s.Require().NoError(err)

	var resp types.QueryReserveAmountResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
}

func (s *IntegrationTestSuite) TestQueryCollateral() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	s.Run("get_all_borrowed", func() {
		// TODO: We need to setup borrowing first prior to testing this out.
		//
		// Ref: https://github.com/umee-network/umee/issues/94
		flags := []string{
			val.Address.String(),
			fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		}

		out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryCollateral(), flags)
		s.Require().NoError(err)

		var resp types.QueryCollateralResponse
		s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
	})

	s.Run("get_denom_borrowed", func() {
		// TODO: We need to setup borrowing first prior to testing this out.
		//
		// Ref: https://github.com/umee-network/umee/issues/94
		flags := []string{
			val.Address.String(),
			fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			fmt.Sprintf("--%s=uumee", cli.FlagDenom),
		}

		out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryCollateral(), flags)
		s.Require().NoError(err)

		var resp types.QueryCollateralResponse
		s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
	})
}

func (s *IntegrationTestSuite) TestQueryCollateralSetting() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	// TODO: We need to setup borrowing first prior to testing this out.
	//
	// Ref: https://github.com/umee-network/umee/issues/94
	flags := []string{
		val.Address.String(),
		"u/uumee",
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryCollateralSetting(), flags)
	s.Require().NoError(err)

	var resp types.QueryCollateralSettingResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
}

func (s *IntegrationTestSuite) TestQueryExchangeRate() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	flags := []string{
		"uumee",
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryExchangeRate(), flags)
	s.Require().NoError(err)

	var resp types.QueryExchangeRateResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
}

func (s *IntegrationTestSuite) TestQueryBorrowLimit() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	flags := []string{
		val.Address.String(),
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryBorrowLimit(), flags)
	s.Require().NoError(err)

	var resp types.QueryBorrowLimitResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
}

func (s *IntegrationTestSuite) TestQueryLiquidationTargets() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	flags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryLiquidationTargets(), flags)
	s.Require().NoError(err)

	var resp types.QueryLiquidationTargetsResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
}

func (s *IntegrationTestSuite) TestCmdLend() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name        string
		command     *cobra.Command
		args        []string
		expectedErr *sdkerrors.Error
	}{
		{
			"invalid asset",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"1000uabcd",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			types.ErrInvalidAsset,
		},
		{
			"valid lend",
			cli.GetCmdLendAsset(),
			[]string{
				val.Address.String(),
				"1000uumee",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.command, tc.args)
			s.Require().NoError(err)

			resp := &sdk.TxResponse{}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp), out.String())

			if tc.expectedErr == nil {
				s.Require().Equal(uint32(0), resp.Code)
			} else {
				s.Require().Equal(tc.expectedErr.ABCICode(), resp.Code)
			}
		})
	}
}
