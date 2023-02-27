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
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        incentive.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", incentive.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQueryUnbondings(),
		GetCmdQueryBonded(),
		GetCmdQueryTotalBonded(),
		GetCmdQueryPendingRewards(),
		GetCmdQueryUpcomingIncentivePrograms(),
		GetCmdQueryOngoingIncentivePrograms(),
		GetCmdQueryCompletedIncentivePrograms(),
		GetCmdQueryIncentiveProgram(),
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

// GetCmdQueryUnbondings creates a Cobra command to query all unbondings associated with a single account.
func GetCmdQueryUnbondings() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbondings [address]",
		Args:  cobra.ExactArgs(1),
		Short: fmt.Sprintf("Query all %s module unbondings associated with a single account", incentive.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := incentive.NewQueryClient(clientCtx)

			resp, err := queryClient.Unbondings(cmd.Context(), &incentive.QueryUnbondings{Address: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryPendingRewards creates a Cobra command to query the pending incentive rewards of a single account.
func GetCmdQueryPendingRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-rewards [address]",
		Args:  cobra.ExactArgs(1),
		Short: fmt.Sprintf("Query all pending %s module rewards associated with a single account", incentive.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := incentive.NewQueryClient(clientCtx)

			resp, err := queryClient.PendingRewards(cmd.Context(), &incentive.QueryPendingRewards{Address: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryBonded creates a Cobra command to query bonded tokens associated with a single account.
func GetCmdQueryBonded() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bonded [address] [denom]",
		Args:  cobra.RangeArgs(1, 2),
		Short: fmt.Sprintf("Query an address's uTokens bonded to the %s module", incentive.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := incentive.NewQueryClient(clientCtx)

			denom := ""
			if len(args) > 1 {
				denom = args[1]
			}

			resp, err := queryClient.Bonded(cmd.Context(), &incentive.QueryBonded{Address: args[0], Denom: denom})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryTotalBonded creates a Cobra command to query bonded tokens across all users.
func GetCmdQueryTotalBonded() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-bonded [denom]",
		Args:  cobra.RangeArgs(0, 1),
		Short: fmt.Sprintf("Query the total uTokens bonded to the %s module", incentive.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := incentive.NewQueryClient(clientCtx)

			denom := ""
			if len(args) > 0 {
				denom = args[0]
			}

			resp, err := queryClient.TotalBonded(cmd.Context(), &incentive.QueryTotalBonded{Denom: denom})
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
		Short: "Query all upcoming incentive programs",
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
		Short: "Query all ongoing incentive programs",
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
		Use:   "completed",
		Args:  cobra.NoArgs,
		Short: "Query all completed incentive programs",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := incentive.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			resp, err := queryClient.CompletedIncentivePrograms(cmd.Context(),
				&incentive.QueryCompletedIncentivePrograms{
					Pagination: pageReq,
				})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "completed incentive programs")

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
