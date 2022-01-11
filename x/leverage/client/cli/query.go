package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/x/leverage/types"
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
		GetCmdQueryReserveAmount(),
		GetCmdQueryCollateral(),
		GetCmdQueryCollateralSetting(),
		GetCmdQueryExchangeRate(),
		GetCmdQueryLendAPY(),
		GetCmdQueryBorrowAPY(),
		GetCmdQueryMarketSize(),
		GetCmdQueryBorrowLimit(),
		GetCmdQueryLiquidationTargets(),
	)

	return cmd
}

// GetCmdQueryAllRegisteredTokens returns a CLI command handler to query for all
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
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryParams returns a CLI command handler to query for the x/leverage
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

			resp, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryBorrowed returns a CLI command handler to query for the amount of
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

			req := &types.QueryBorrowedRequest{
				Address: args[0],
			}
			if d, err := cmd.Flags().GetString(FlagDenom); len(d) > 0 && err == nil {
				req.Denom = d
			}

			resp, err := queryClient.Borrowed(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	cmd.Flags().String(FlagDenom, "", "Query for a specific denomination")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryReserveAmount returns a CLI command handler to query for the
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

			req := &types.QueryReserveAmountRequest{
				Denom: args[0],
			}

			resp, err := queryClient.ReserveAmount(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryCollateralSetting returns a CLI command handler to query for the collateral
// setting of a specific token denom for an address.
func GetCmdQueryCollateralSetting() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collateral-setting [addr] [denom]",
		Args:  cobra.ExactArgs(2),
		Short: "Query for the collateral setting of a specific denom for an address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryCollateralSettingRequest{
				Address: args[0],
				Denom:   args[1],
			}

			resp, err := queryClient.CollateralSetting(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryCollateral returns a CLI command handler to query for the amount of
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

			req := &types.QueryCollateralRequest{
				Address: args[0],
			}
			if d, err := cmd.Flags().GetString(FlagDenom); len(d) > 0 && err == nil {
				req.Denom = d
			}

			resp, err := queryClient.Collateral(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	cmd.Flags().String(FlagDenom, "", "Query for a specific denomination")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryExchangeRate returns a CLI command handler to query for the
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

			req := &types.QueryExchangeRateRequest{
				Denom: args[0],
			}

			resp, err := queryClient.ExchangeRate(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryAvailableBorrow returns a CLI command handler to query for the
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

			req := &types.QueryAvailableBorrowRequest{
				Denom: args[0],
			}

			resp, err := queryClient.AvailableBorrow(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryLendAPY returns a CLI command handler to query for the
// lend APY of a specific uToken.
func GetCmdQueryLendAPY() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lend-apy [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the lend APY of a specified denomination",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryLendAPYRequest{
				Denom: args[0],
			}

			resp, err := queryClient.LendAPY(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryBorrowAPY returns a CLI command handler to query for the
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

			req := &types.QueryBorrowAPYRequest{
				Denom: args[0],
			}

			resp, err := queryClient.BorrowAPY(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryMarketSize returns a CLI command handler to query for the
// Market Size of a specific token.
func GetCmdQueryMarketSize() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market-size [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the USD market size of a specified denomination",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryMarketSizeRequest{
				Denom: args[0],
			}

			resp, err := queryClient.MarketSize(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryBorrowLimit returns a CLI command handler to query for the
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

			req := &types.QueryBorrowLimitRequest{
				Address: args[0],
			}

			resp, err := queryClient.BorrowLimit(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryLiquidationTargets returns a CLI command handler to query for
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

			req := &types.QueryLiquidationTargetsRequest{}

			resp, err := queryClient.LiquidationTargets(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
