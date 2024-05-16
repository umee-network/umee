package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v6/util/cli"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

// Flag constants
const (
	FlagDenom = "denom"
)

// GetQueryCmd returns the CLI query commands for the x/leverage module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		QueryParams(),
		QueryRegisteredTokens(),
		QuerySpecialAssets(),
		QueryMarketSummary(),
		QueryAccountBalances(),
		QueryAccountSummary(),
		QueryAccountsSummary(),
		QueryLiquidationTargets(),
		QueryBadDebts(),
		QueryMaxWithdraw(),
		QueryMaxBorrow(),
		QueryInspect(),
		QueryInspectAccount(),
	)

	return cmd
}

// QueryParams creates a Cobra command to query for the x/leverage
// module parameters.
func QueryParams() *cobra.Command {
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

// QueryRegisteredTokens creates a Cobra command to query for all
// the registered tokens in the x/leverage module.
func QueryRegisteredTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registered-tokens [base_denom]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Query for all the current registered tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryRegisteredTokens{}
			if len(args) == 1 {
				req.BaseDenom = args[0]
			}
			resp, err := queryClient.RegisteredTokens(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QuerySpecialAssets creates a Cobra command to query for all
// the special asset pairs in the x/leverage module.
func QuerySpecialAssets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "special-assets",
		Args:  cobra.RangeArgs(0, 1),
		Short: "Query for all special asset pairs, or only those affecting a single collateral token.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QuerySpecialAssets{}
			if len(args) > 0 {
				req.Denom = args[0]
			}
			resp, err := queryClient.SpecialAssets(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryMarketSummary creates a Cobra command to query for the
// Market Summary of a specific token.
func QueryMarketSummary() *cobra.Command {
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

// QueryAccountBalances creates a Cobra command to query for the
// supply, collateral, and borrow positions of an account.
func QueryAccountBalances() *cobra.Command {
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

// QueryAccountSummary creates a Cobra command to query for USD
// values representing an account's positions and borrowing limits.
func QueryAccountSummary() *cobra.Command {
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

// QueryAccountsSummary creates a Cobra command to query the USD
// values representing an all account's positions and borrowing limits.
func QueryAccountsSummary() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts-summary",
		Args:  cobra.NoArgs,
		Short: "Query the position USD values and borrowing limits for an all accounts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryAccountsSummary{
				Pagination: pageReq,
			}
			resp, err := queryClient.AccountsSummary(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "accounts-summary")

	return cmd
}

// QueryLiquidationTargets creates a Cobra command to query for
// all eligible liquidation targets.
func QueryLiquidationTargets() *cobra.Command {
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

// QueryBadDebts creates a Cobra command to query for
// all bad debts.
func QueryBadDebts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bad-debts",
		Args:  cobra.ExactArgs(0),
		Short: "Query for all bad debts",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryBadDebts{}
			resp, err := queryClient.BadDebts(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryMaxWithdraw creates a Cobra command to query for
// the maximum amount of a given token an address can withdraw.
func QueryMaxWithdraw() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "max-withdraw [addr] [denom]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Query for the maximum amount of a given base token an address can withdraw",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryMaxWithdraw{
				Address: args[0],
			}
			if len(args) > 1 {
				req.Denom = args[1]
			}
			if err := req.ValidateBasic(); err != nil {
				return err
			}
			resp, err := queryClient.MaxWithdraw(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryMaxBorrow creates a Cobra command to query for
// the maximum amount of a given token an address can borrow.
func QueryMaxBorrow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "max-borrow [addr] [denom]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Query for the maximum amount of a given base token an address can borrow",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryMaxBorrow{
				Address: args[0],
			}
			if len(args) > 1 {
				req.Denom = args[1]
			}
			resp, err := queryClient.MaxBorrow(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryInspect creates a Cobra command to query for the inspector command.
func QueryInspect() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect [symbol] [borrowed] [collateral [danger] [ltv]",
		Args:    cobra.MinimumNArgs(2),
		Short:   "Inspect accounts with the leverage module, filtered with various minimum values.",
		Example: "umeed q leverage inspect OSMO 100 0 0.9 0",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryInspect{
				Symbol: args[0],
			}
			req.Borrowed, err = strconv.ParseFloat(args[1], 64)
			if err != nil {
				return err
			}
			if len(args) >= 3 {
				req.Collateral, err = strconv.ParseFloat(args[2], 64)
				if err != nil {
					return err
				}
			}
			if len(args) >= 4 {
				req.Danger, err = strconv.ParseFloat(args[3], 64)
				if err != nil {
					return err
				}
			}
			if len(args) >= 5 {
				req.Ltv, err = strconv.ParseFloat(args[4], 64)
				if err != nil {
					return err
				}
			}
			resp, err := queryClient.Inspect(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryInspectAccount creates a Cobra command to inspect a single account.
func QueryInspectAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect-account [addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Inspect a single account",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryInspectAccount{
				Address: args[0],
			}
			resp, err := queryClient.InspectAccount(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
