package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v4/x/incentive"
)

// Flag constants
const (
	FlagDenom = "denom"
)

// GetQueryCmd returns the CLI query commands for the x/incentive module.
func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        incentive.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", incentive.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQueryIncentiveProgram(),
		GetCmdQueryUpcomingIncentivePrograms(),
		GetCmdQueryOngoingIncentivePrograms(),
		GetCmdQueryCompletedIncentivePrograms(),
		GetCmdQueryUnbondings(),
		GetCmdQueryBonded(),
		GetCmdQueryTotalBonded(),
		GetCmdQueryPendingRewards(),
	)

	return cmd
}

// GetCmdQueryParams creates a Cobra command to query for the x/incentive
// module parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: fmt.Sprintf("Query the %s module parameters", incentive.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := incentive.NewQueryClient(clientCtx)

			resp, err := queryClient.Params(cmd.Context(), &incentive.QueryParams{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryIncentiveProgram creates a Cobra command to query a single incentive program
// by its ID.
func GetCmdQueryIncentiveProgram() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incentive-program [id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query a single incentive program by its ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := incentive.NewQueryClient(clientCtx)

			id, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return err
			}

			resp, err := queryClient.IncentiveProgram(cmd.Context(), &incentive.QueryIncentiveProgram{Id: uint32(id)})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// TODO: Add all queries
