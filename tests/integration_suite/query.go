package integrationsuite_test

import (
	"fmt"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"gotest.tools/v3/assert"
)

type TestQuery struct {
	Msg              string
	Command          *cobra.Command
	Args             []string
	ResponseType     proto.Message
	ExpectedResponse proto.Message
	ErrMsg           string
}

func (tq TestQuery) Run(s *IntegrationTestSuite) {
	clientCtx := s.Network.Validators[0].ClientCtx
	queryFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	out, err := clitestutil.ExecTestCLICmd(clientCtx, tq.Command, append(tq.Args, queryFlags...))

	if len(tq.ErrMsg) != 0 {
		assert.ErrorContains(s.T, err, tq.ErrMsg)
	} else {
		assert.NilError(s.T, err, tq.Msg)

		err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), tq.ResponseType)
		assert.NilError(s.T, err, tq.Msg)
		assert.Equal(s.T, tq.ExpectedResponse.String(), tq.ResponseType.String(), tq.Msg)
	}
}
