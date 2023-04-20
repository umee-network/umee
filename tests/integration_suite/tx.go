package integrationsuite_test

import (
	"fmt"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

type TestTransaction struct {
	Msg         string
	Command     *cobra.Command
	Args        []string
	ExpectedErr *errors.Error
}

func (t TestTransaction) Run(s *IntegrationTestSuite) {
	require := s.Require()
	clientCtx := s.Network.Validators[0].ClientCtx
	txFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.Network.Validators[0].Address),
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "10000000"),
		fmt.Sprintf("--%s=%s", flags.FlagFees, "1000000uumee"),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, t.Command, append(t.Args, txFlags...))
	require.NoError(err, t.Msg)

	resp := &sdk.TxResponse{}
	err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp)
	require.NoError(err, t.Msg)

	if t.ExpectedErr == nil {
		require.Equal(0, int(resp.Code), "msg: %s\nresp: %s", t.Msg, resp)
	} else {
		require.Equal(int(t.ExpectedErr.ABCICode()), int(resp.Code), t.Msg)
	}
}
