package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/v4/util/cli"
	"github.com/umee-network/umee/v4/x/uibc"
)

// GetQueryCmd returns the CLI query commands for the x/uibc module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        uibc.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", uibc.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryParams(),
		GetOutflows(),
	)

	return cmd
}

// GetCmdQueryParams creates a Cobra command to query for the x/uibc
// module parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the uibc module parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := uibc.NewQueryClient(clientCtx)
			resp, err := queryClient.Params(cmd.Context(), &uibc.QueryParams{})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetOutflows returns cmd creator
func GetOutflows() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outflows [denom]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Get the outflows for ibc and native denoms",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := uibc.NewQueryClient(clientCtx)
			queryReq := uibc.QueryOutflows{}
			if len(args) > 0 {
				queryReq.Denom = args[0]
			}
			resp, err := queryClient.Outflows(cmd.Context(), &queryReq)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
