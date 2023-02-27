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

// GetCmdQueryUpcomingIncentivePrograms creates a Cobra command to query for all upcoming
// incentive programs.
func GetCmdQueryUpcomingIncentivePrograms() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upcoming",
		Args:  cobra.NoArgs,
		Short: fmt.Sprintf("Query all upcoming incentive programs"),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := incentive.NewQueryClient(clientCtx)

			resp, err := queryClient.UpcomingIncentivePrograms(cmd.Context(),
				&incentive.QueryUpcomingIncentivePrograms{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryOngoingIncentivePrograms creates a Cobra command to query for all ongoing
// incentive programs.
func GetCmdQueryOngoingIncentivePrograms() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ongoing",
		Args:  cobra.NoArgs,
		Short: fmt.Sprintf("Query all ongoing incentive programs"),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := incentive.NewQueryClient(clientCtx)

			resp, err := queryClient.OngoingIncentivePrograms(cmd.Context(),
				&incentive.QueryOngoingIncentivePrograms{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryCompletedIncentivePrograms creates a Cobra command to query for all completed
// incentive programs.
func GetCmdQueryCompletedIncentivePrograms() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completed",  // TODO: pagination
		Args:  cobra.NoArgs, // TODO: pagination
		Short: fmt.Sprintf("Query all completed incentive programs"),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := incentive.NewQueryClient(clientCtx)

			resp, err := queryClient.CompletedIncentivePrograms(cmd.Context(),
				&incentive.QueryCompletedIncentivePrograms{
					// TODO: what pagination should we use for the CLI?
				})
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
