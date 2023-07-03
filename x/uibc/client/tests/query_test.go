package tests

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	itestsuite "github.com/umee-network/umee/v5/tests/cli"
	"github.com/umee-network/umee/v5/x/uibc"
	"github.com/umee-network/umee/v5/x/uibc/client/cli"
)

func (s *IntegrationTests) TestQueryParams(_ *testing.T) {
	queries := []itestsuite.TestQuery{
		{
			Name:    "Query params",
			Command: cli.GetCmdQueryParams(),
			Args: []string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			Response: &uibc.QueryParamsResponse{},
			ExpectedResponse: &uibc.QueryParamsResponse{
				Params: uibc.DefaultParams(),
			},
			ErrMsg: "",
		},
	}

	s.RunTestQueries(queries...)
}

func (s *IntegrationTests) TestGetQuota(_ *testing.T) {
	queries := []itestsuite.TestQuery{
		{
			Name:    "Get ibc-transfer quota of all denoms",
			Command: cli.GetOutflows(),
			Args: []string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			Response: &uibc.QueryOutflowsResponse{},
			ExpectedResponse: &uibc.QueryOutflowsResponse{
				Amount: sdk.NewDec(0),
			},
			ErrMsg: "",
		},
		{
			Name:    "Get ibc-transfer quota of denom umee",
			Command: cli.GetOutflows(),
			Args: []string{
				"uumee",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			Response: &uibc.QueryOutflowsResponse{},
			ExpectedResponse: &uibc.QueryOutflowsResponse{
				Amount: sdk.NewDec(0),
			},
			ErrMsg: "",
		},
	}

	s.RunTestQueries(queries...)
}
