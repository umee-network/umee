package tests

import (
	"fmt"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/umee-network/umee/x/leverage/types"

	"github.com/umee-network/umee/x/leverage/client/cli"
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

	//
	//	Note: It appears that the list of registered tokens is
	//	empty as of the end of this function. Need to either
	// 	register `uumee` immediately or use custom genesis
	//	state.
	//
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

	// Note: Comment out once we're done figuring out how to register tokens in suite
	fmt.Println("\n\n" + string(out.Bytes()))
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
