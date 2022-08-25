package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v3/util/cli"
	"github.com/umee-network/umee/v3/x/leverage/types"
)

// Flag constants
const (
	FlagDenom = "denom"
)

// GetQueryCmd returns the CLI query commands for the x/leverage module.
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
		GetCmdQueryRegisteredTokens(),
		GetCmdQueryMarketSummary(),
		GetCmdQueryAccountBalances(),
		GetCmdQueryAccountSummary(),
		GetCmdQueryLiquidationTargets(),
	)

	return cmd
}

// GetCmdQueryParams creates a Cobra command to query for the x/leverage
// module parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the x/leverage module parameters",
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

// GetCmdQueryRegisteredTokens creates a Cobra command to query for all
// the registered tokens in the x/leverage module.
func GetCmdQueryRegisteredTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registered-tokens",
		Args:  cobra.NoArgs,
		Short: "Query for all the current registered tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			resp, err := queryClient.RegisteredTokens(cmd.Context(), &types.QueryRegisteredTokens{})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryMarketSummary creates a Cobra command to query for the
// Market Summary of a specific token.
func GetCmdQueryMarketSummary() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market-summary [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the market summary of a specified denomination",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryMarketSummary{
				Denom: args[0],
			}
			resp, err := queryClient.MarketSummary(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryAccountBalances creates a Cobra command to query for the
// supply, collateral, and borrow positions of an account.
func GetCmdQueryAccountBalances() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-balances [addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the total supplied, collateral, and borrowed tokens for an address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryAccountBalances{
				Address: args[0],
			}
			resp, err := queryClient.AccountBalances(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryAccountSummary creates a Cobra command to query for USD
// values representing an account's positions and borrowing limits.
func GetCmdQueryAccountSummary() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-summary [addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for position USD values and borrowing limits for an address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryAccountSummary{
				Address: args[0],
			}
			resp, err := queryClient.AccountSummary(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryLiquidationTargets creates a Cobra command to query for
// all eligible liquidation targets.
func GetCmdQueryLiquidationTargets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidation-targets",
		Args:  cobra.ExactArgs(0),
		Short: "Query for all borrower addresses eligible for liquidation",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryLiquidationTargets{}
			resp, err := queryClient.LiquidationTargets(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
