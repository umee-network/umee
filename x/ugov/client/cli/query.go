package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/v6/util/cli"
	"github.com/umee-network/umee/v6/x/ugov"
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
		QueryInflationParams(),
		QueryInflationCyleEnd(),
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

// QueryInflationParams create the Msg/QueryInflationParams CLI.
func QueryInflationParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inflation-params",
		Args:  cobra.NoArgs,
		Short: "Query the inflation params",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := ugov.NewQueryClient(clientCtx)
			resp, err := queryClient.InflationParams(cmd.Context(), &ugov.QueryInflationParams{})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryInflationCyleEnd create the Msg/QueryInflationCyleEnd CLI.
func QueryInflationCyleEnd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inflation-cycle-end",
		Args:  cobra.NoArgs,
		Short: "Query the When the Inflation Cycle is Started",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := ugov.NewQueryClient(clientCtx)
			resp, err := queryClient.InflationCycleEnd(cmd.Context(), &ugov.QueryInflationCycleEnd{})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
