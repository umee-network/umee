package tests

import (
	"fmt"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
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
