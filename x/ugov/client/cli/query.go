package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/v4/util/cli"
	"github.com/umee-network/umee/v4/x/ugov"
)

// GetQueryCmd returns the CLI query commands for the x/ugov module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        ugov.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", ugov.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetMinGasPrice(),
	)

	return cmd
}

// GetMinGasPrice creates a Cobra command to query for the x/ugov min gas price
func GetMinGasPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "min-gas-price",
		Args:  cobra.NoArgs,
		Short: "Query the minimum gas price",
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
