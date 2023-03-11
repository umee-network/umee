package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v4/util/cli"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

// GetCmdQueryBorrowers creates a Cobra command to query for
// all borrowers sorted by borrowed value.
func GetCmdQueryBorrowers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "borrowers",
		Args:  cobra.ExactArgs(0),
		Short: "Query for all borrower addresses sorted by borrowed value",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryBorrowers{MinimumValue: sdk.MustNewDecFromStr("0")}
			resp, err := queryClient.Borrowers(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
