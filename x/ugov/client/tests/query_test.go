package tests

import (
	"fmt"
	"testing"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	itestsuite "github.com/umee-network/umee/v5/tests/cli"
	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/ugov"
	"github.com/umee-network/umee/v5/x/ugov/client/cli"
)

func (s *IntegrationTests) TestMinGasPrice(_ *testing.T) {
	queries := []itestsuite.TestQuery{
		{
			Name:    "Query min gas price for tx",
			Command: cli.QueryMinGasPrice(),
			Args: []string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			Response: &ugov.QueryMinGasPriceResponse{},
			ExpectedResponse: &ugov.QueryMinGasPriceResponse{
				MinGasPrice: coin.UmeeDec("0.1"),
			},
			ErrMsg: "",
		},
	}

	s.RunTestQueries(queries...)
}

func (s *IntegrationTests) TestInflationParams(_ *testing.T) {
	queries := []itestsuite.TestQuery{
		{
			Name:    "Query inflation params",
			Command: cli.QueryInflationParams(),
			Args: []string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			Response: &ugov.QueryInflationParamsResponse{},
			ExpectedResponse: &ugov.QueryInflationParamsResponse{
				Params: ugov.DefaultInflationParams(),
			},
			ErrMsg: "",
		},
	}

	s.RunTestQueries(queries...)
}
