package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/v3/util/cli"
	"github.com/umee-network/umee/v3/x/ibctransfer/types"
)

// GetQueryCmd returns the CLI query commands for the x/ibc-rate-limit module.
func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryParams(),
		GetRateLimitsForIBCDenoms(),
	)

	return cmd
}

// GetCmdQueryParams creates a Cobra command to query for the x/ibc-rate-limit
// module parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the x/ibc-rate-limit module parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			resp, err := queryClient.Params(cmd.Context(), &types.QueryParams{})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetRateLimitsForIBCDenoms
func GetRateLimitsForIBCDenoms() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-limits [ibc-denom]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Get the rate limits for ibc denoms",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			if len(args) > 0 {
				resp, err := queryClient.RateLimitsOfIBCDenom(cmd.Context(), &types.QueryRateLimitsOfIBCDenom{IbcDenom: args[0]})
				return cli.PrintOrErr(resp, err, clientCtx)
			}

			resp, err := queryClient.RateLimitsOfIBCDenoms(cmd.Context(), &types.QueryRateLimitsOfIBCDenoms{})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
