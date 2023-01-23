//go:build experimental
// +build experimental

package tests

import (
	"fmt"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/umee-network/umee/v4/x/uibc"
	"github.com/umee-network/umee/v4/x/uibc/client/cli"
)

func (s *IntegrationTestSuite) TestQueryParams() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	args := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryParams(), args)
	s.Require().NoError(err)

	var res uibc.QueryParamsResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))

	s.Require().NotEmpty(res.Params)
}

func (s *IntegrationTestSuite) TestGetQuota() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	args := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetQuota(), args)
	s.Require().NoError(err)

	var res uibc.QueryQuotaResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))

	s.Require().Equal(len(res.Quota), 0)
}
