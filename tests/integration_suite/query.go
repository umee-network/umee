package integrationsuite_test

import (
	"fmt"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

type TestQuery struct {
	Msg              string
	Command          *cobra.Command
	Args             []string
	ExpectErr        bool
	ResponseType     proto.Message
	ExpectedResponse proto.Message
}

func (t TestQuery) Run(s *IntegrationTestSuite) {
	require := s.Require()
	clientCtx := s.Network.Validators[0].ClientCtx
	queryFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	out, err := clitestutil.ExecTestCLICmd(clientCtx, t.Command, append(t.Args, queryFlags...))

	if t.ExpectErr {
		require.Error(err, t.Msg)
	} else {
		require.NoError(err, t.Msg)

		err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), t.ResponseType)
		require.NoError(err, t.Msg)
		require.Equal(t.ExpectedResponse.String(), t.ResponseType.String(), t.Msg)
	}
}
