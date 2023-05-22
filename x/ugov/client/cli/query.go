package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/v4/util/cli"
	"github.com/umee-network/umee/v4/x/ugov"
)

// GetQueryCmd returns the CLI query commands for the x/uibc module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        ugov.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", ugov.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		QueryMinGasPrice(),
	)

	return cmd
}

// QueryMinGasPrice creates the Msg/QueryMinGasPrice CLI.
func QueryMinGasPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "min-gas-price",
		Args:  cobra.NoArgs,
		Short: "Query the tx minimum gas price",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := ugov.NewQueryClient(clientCtx)
			resp, err := queryClient.MinGasPrice(cmd.Context(), &ugov.QueryMinGasPrice{})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
