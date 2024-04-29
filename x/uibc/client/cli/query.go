package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v6/util/cli"
	"github.com/umee-network/umee/v6/x/uibc"
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
		QueryParams(),
		GetOutflows(),
		GetInflows(),
		GetQuotaExpireTime(),
		QueryDenomOwners(),
	)

	return cmd
}

// GetQuotaExpireTime returns end time for the current quota period.
func GetQuotaExpireTime() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quota-expire-time",
		Args:  cobra.NoArgs,
		Short: "Get the current ibc quota expire time.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := uibc.NewQueryClient(clientCtx)

			resp, err := queryClient.QuotaExpires(cmd.Context(), &uibc.QueryQuotaExpires{})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetInflows returns total inflow sum, if denom specified it will return quota inflow of the denom.
func GetInflows() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inflows [denom]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Get the total ibc inflow sum of registered tokens.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := uibc.NewQueryClient(clientCtx)

			req := &uibc.QueryInflows{}
			if len(args) > 0 && len(args[0]) != 0 {
				req.Denom = args[0]
			}

			resp, err := queryClient.Inflows(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryParams creates a Cobra command to query for the x/uibc
// module parameters.
func QueryParams() *cobra.Command {
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
			if len(args) > 0 && len(args[0]) != 0 {
				queryReq.Denom = args[0]
			}
			resp, err := queryClient.Outflows(cmd.Context(), &queryReq)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// QueryDenomOwners creates the Query/DenomOwners CLI.
func QueryDenomOwners() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denom-owners [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Queries for all account addresses that own a particular token denomination.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := uibc.NewQueryClient(clientCtx)
			resp, err := queryClient.DenomOwners(cmd.Context(), &banktypes.QueryDenomOwnersRequest{
				Denom:      args[0],
				Pagination: pageReq,
			})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "denom-owners")

	return cmd
}
