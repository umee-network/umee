package tests

import (
	"fmt"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v4/x/uibc"
	"github.com/umee-network/umee/v4/x/uibc/client/cli"
)

func (s *IntegrationTestSuite) TestQueryParams(t *testing.T) {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	args := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryParams(), args)
	assert.NilError(t, err)

	var res uibc.QueryParamsResponse
	assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
	assert.DeepEqual(t, res.Params, uibc.DefaultParams())
}

func (s *IntegrationTestSuite) TestGetQuota(t *testing.T) {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	tests := []struct {
		name        string
		args        []string
		errMsg      string
		noOfRecords int
	}{
		{
			name: "Get ibc-transfer quota of all denoms",
			args: []string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			errMsg:      "",
			noOfRecords: 1,
		},
		{
			name: "Get ibc-transfer quota of denom umee",
			args: []string{
				"uumee",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			errMsg:      "",
			noOfRecords: 1,
		},
		/* {
			name: "Get ibc-transfer quota of dummy denom ",
			args: []string{
				"dummy",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			errMsg:      "no quota for ibc denom",
			noOfRecords: 0,
		}, */
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetOutflows(), tc.args)
			if tc.errMsg == "" {
				var res uibc.QueryOutflowsResponse
				assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				assert.Equal(t, len(res.Outflows), tc.noOfRecords)
			} else {
				assert.ErrorContains(t, err, tc.errMsg)
			}
		})
	}
}
