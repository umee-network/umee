//go:build experimental
// +build experimental

package tests

import (
	"fmt"
	"testing"

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

	tests := []struct {
		name        string
		args        []string
		errExpected bool
		noOfRecords int
	}{
		{
			name: "Get ibc-transfer quota of all denoms",
			args: []string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			errExpected: false,
			noOfRecords: 0,
		},
		{
			name: "Get ibc-transfer quota of denom umee",
			args: []string{
				"uumee",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			errExpected: true,
			noOfRecords: 0,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetQuota(), tc.args)
			if tc.errExpected {
				s.Require().Error(err)
			} else {
				var res uibc.QueryQuotaResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(res.Quota), tc.noOfRecords)
			}
		})
	}
}
