package tests

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	itestsuite "github.com/umee-network/umee/v4/tests/integration_suite"
	"github.com/umee-network/umee/v4/x/uibc"
	"github.com/umee-network/umee/v4/x/uibc/client/cli"
)

func (s *IntegrationTests) TestQueryParams(t *testing.T) {
	queries := []itestsuite.TestQuery{
		{
			Msg:     "Query params",
			Command: cli.GetCmdQueryParams(),
			Args: []string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			ExpectErr:    false,
			ResponseType: &uibc.QueryParamsResponse{},
			ExpectedResponse: &uibc.QueryParamsResponse{
				Params: uibc.DefaultParams(),
			},
		},
	}

	s.RunTestQueries(queries...)
}

func (s *IntegrationTests) TestGetQuota(t *testing.T) {
	queries := []itestsuite.TestQuery{
		{
			Msg:     "Get ibc-transfer quota of all denoms",
			Command: cli.GetOutflows(),
			Args: []string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			ExpectErr:    false,
			ResponseType: &uibc.QueryOutflowsResponse{},
			ExpectedResponse: &uibc.QueryOutflowsResponse{
				Amount: sdk.NewDec(0),
			},
		},
		{
			Msg:     "Get ibc-transfer quota of denom umee",
			Command: cli.GetOutflows(),
			Args: []string{
				"uumee",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			ExpectErr:    false,
			ResponseType: &uibc.QueryOutflowsResponse{},
			ExpectedResponse: &uibc.QueryOutflowsResponse{
				Amount: sdk.NewDec(0),
			},
		},
	}

	s.RunTestQueries(queries...)
}
