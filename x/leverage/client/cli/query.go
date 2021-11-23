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
