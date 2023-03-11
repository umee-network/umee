package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v4/util/cli"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

// GetCmdQueryInspect creates a Cobra command to query for the inspector command.
func GetCmdQueryInspect() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect [flavor] [denom] [value]",
		Args:    cobra.ExactArgs(3),
		Short:   "Inspect accounts with the leverage module.",
		Example: "umeed q leverage inspect danger-by-borrowed all 0.95",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryInspect{
				Flavor: args[0],
				Symbol: args[1],
				Value:  sdk.MustNewDecFromStr(args[2]),
			}
			resp, err := queryClient.Inspect(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
