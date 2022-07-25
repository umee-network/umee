package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v2/util/cli"
	"github.com/umee-network/umee/v2/x/leverage/types"
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
		GetCmdQueryAllRegisteredTokens(),
		GetCmdQueryParams(),
		GetCmdQueryBorrowed(),
		GetCmdQueryBorrowedValue(),
		GetCmdQuerySupplied(),
		GetCmdQuerySuppliedValue(),
		GetCmdQueryReserveAmount(),
		GetCmdQueryCollateral(),
		GetCmdQueryCollateralValue(),
		GetCmdQueryExchangeRate(),
		GetCmdQuerySupplyAPY(),
		GetCmdQueryBorrowAPY(),
		GetCmdQueryTotalSuppliedValue(),
		GetCmdQueryTotalSupplied(),
		GetCmdQueryBorrowLimit(),
		GetCmdQueryLiquidationThreshold(),
		GetCmdQueryLiquidationTargets(),
		GetCmdQueryMarketSummary(),
		GetCmdQueryTotalCollateral(),
		GetCmdQueryTotalBorrowed(),
	)

	return cmd
}

// GetCmdQueryAllRegisteredTokens creates a Cobra command to query for all
// the registered tokens in the x/leverage module.
func GetCmdQueryAllRegisteredTokens() *cobra.Command {
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

// GetCmdQueryBorrowed creates a Cobra command to query for the amount of
// total borrowed tokens for a given address.
func GetCmdQueryBorrowed() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "borrowed [addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the total amount of borrowed tokens for an address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryBorrowed{
				Address: args[0],
			}
			if d, err := cmd.Flags().GetString(FlagDenom); len(d) > 0 && err == nil {
				req.Denom = d
			}
			resp, err := queryClient.Borrowed(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	cmd.Flags().String(FlagDenom, "", "Query for a specific denomination")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryBorrowedValue creates a Cobra command to query for the USD
// value of total borrowed tokens for a given address.
func GetCmdQueryBorrowedValue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "borrowed-value [addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the total USD value of borrowed tokens for an address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryBorrowedValue{
				Address: args[0],
			}
			if d, err := cmd.Flags().GetString(FlagDenom); len(d) > 0 && err == nil {
				req.Denom = d
			}
			resp, err := queryClient.BorrowedValue(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	cmd.Flags().String(FlagDenom, "", "Query for value of only a specific denomination")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQuerySupplied creates a Cobra command to query for the amount of
// tokens supplied by a given address.
func GetCmdQuerySupplied() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supplied [addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the total amount of tokens supplied by an address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QuerySupplied{
				Address: args[0],
			}
			if d, err := cmd.Flags().GetString(FlagDenom); len(d) > 0 && err == nil {
				req.Denom = d
			}
			resp, err := queryClient.Supplied(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	cmd.Flags().String(FlagDenom, "", "Query for a specific denomination")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQuerySuppliedValue creates a Cobra command to query for the USD value of
// total tokens supplied by a given address.
func GetCmdQuerySuppliedValue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supplied-value [addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the USD value of tokens supplied by an address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QuerySuppliedValue{
				Address: args[0],
			}
			if d, err := cmd.Flags().GetString(FlagDenom); len(d) > 0 && err == nil {
				req.Denom = d
			}
			resp, err := queryClient.SuppliedValue(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	cmd.Flags().String(FlagDenom, "", "Query for value of only a specific denomination")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryReserveAmount creates a Cobra command to query for the
// reserved amount of a specific token.
func GetCmdQueryReserveAmount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reserved [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the amount reserved of a specified denomination",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryReserveAmount{
				Denom: args[0],
			}
			resp, err := queryClient.ReserveAmount(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryCollateral creates a Cobra command to query for the amount of
// total collateral tokens for a given address.
func GetCmdQueryCollateral() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collateral [addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the total amount of collateral tokens for an address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryCollateral{
				Address: args[0],
			}
			if d, err := cmd.Flags().GetString(FlagDenom); len(d) > 0 && err == nil {
				req.Denom = d
			}
			resp, err := queryClient.Collateral(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	cmd.Flags().String(FlagDenom, "", "Query for a specific denomination")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryCollateralValue creates a Cobra command to query for the USD
// value of total collateral tokens for a given address.
func GetCmdQueryCollateralValue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collateral-value [addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the total USD value of collateral tokens for an address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryCollateralValue{
				Address: args[0],
			}
			if d, err := cmd.Flags().GetString(FlagDenom); len(d) > 0 && err == nil {
				req.Denom = d
			}
			resp, err := queryClient.CollateralValue(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	cmd.Flags().String(FlagDenom, "", "Query for value of only a specific denomination")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryExchangeRate creates a Cobra command to query for the
// exchange rate of a specific uToken.
func GetCmdQueryExchangeRate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-rate [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the exchange rate of a specified denomination",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryExchangeRate{
				Denom: args[0],
			}
			resp, err := queryClient.ExchangeRate(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryAvailableBorrow creates a Cobra command to query for the
// available amount to borrow of a specific denom.
func GetCmdQueryAvailableBorrow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "available-borrow [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the available amount to borrow of a specified denomination",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryAvailableBorrow{
				Denom: args[0],
			}
			resp, err := queryClient.AvailableBorrow(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQuerySupplyAPY creates a Cobra command to query for the
// supply APY of a specific uToken.
func GetCmdQuerySupplyAPY() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supply-apy [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the supply APY of a specified denomination",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QuerySupplyAPY{
				Denom: args[0],
			}
			resp, err := queryClient.SupplyAPY(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryBorrowAPY creates a Cobra command to query for the
// borrow APY of a specific token.
func GetCmdQueryBorrowAPY() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "borrow-apy [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the borrow APY of a specified denomination",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryBorrowAPY{
				Denom: args[0],
			}
			resp, err := queryClient.BorrowAPY(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryTotalSuppliedValue creates a Cobra command to query for the
// total supply of a specific token.
func GetCmdQueryTotalSuppliedValue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-supplied-value [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the USD value of the total supply of a specified denomination",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryTotalSuppliedValue{
				Denom: args[0],
			}
			resp, err := queryClient.TotalSuppliedValue(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryTotalSupplied creates a Cobra command to query for the
// Total Supplied of a specific token, measured in base tokens.
func GetCmdQueryTotalSupplied() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-supplied [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the total supply of a specified denomination measured in base tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryTotalSupplied{
				Denom: args[0],
			}
			resp, err := queryClient.TotalSupplied(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryBorrowLimit creates a Cobra command to query for the
// borrow limit of a specific borrower.
func GetCmdQueryBorrowLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "borrow-limit [addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the borrow limit of a specified borrower",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryBorrowLimit{
				Address: args[0],
			}
			resp, err := queryClient.BorrowLimit(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryLiquidationThreshold creates a Cobra command to query a
// liquidation threshold of a specific borrower.
func GetCmdQueryLiquidationThreshold() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidation-threshold [addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query a liquidation threshold of a specified borrower",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryLiquidationThreshold{
				Address: args[0],
			}
			resp, err := queryClient.LiquidationThreshold(cmd.Context(), req)
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

// GetCmdQueryLiquidationTargets creates a Cobra command to query for
// all eligible liquidation targets
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

// GetCmdQueryTotalCollateral creates a Cobra command to query for the
// total collateral amount of a specific token.
func GetCmdQueryTotalCollateral() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-collateral [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the total amount of collateral of a uToken denomination",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryTotalCollateral{
				Denom: args[0],
			}
			resp, err := queryClient.TotalCollateral(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryTotalBorrowed creates a Cobra command to query for the
// total borrowed amount of a specific token.
func GetCmdQueryTotalBorrowed() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-borrowed [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the total amount borrowed of a token denomination",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryTotalBorrowed{
				Denom: args[0],
			}
			resp, err := queryClient.TotalBorrowed(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
