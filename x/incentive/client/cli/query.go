package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v6/util/cli"
	"github.com/umee-network/umee/v6/x/incentive"
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
		QueryParams(),
		QueryAccountBonds(),
		QueryCurrentRates(),
		QueryActualRates(),
		QueryTotalBonded(),
		QueryTotalUnbonding(),
		QueryPendingRewards(),
		QueryUpcomingIncentivePrograms(),
		QueryOngoingIncentivePrograms(),
		QueryCompletedIncentivePrograms(),
		QueryIncentiveProgram(),
	)

	return cmd
}

// QueryParams creates a Cobra command to query for the x/incentive
// module parameters.
func QueryParams() *cobra.Command {
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
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// QueryAccountBonds creates a Cobra command to query all bonds and unbondings associated with a single account.
func QueryAccountBonds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-bonds [address]",
		Args:  cobra.ExactArgs(1),
		Short: fmt.Sprintf("Query all %s module bonds and unbondings associated with an account", incentive.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := incentive.NewQueryClient(clientCtx)

			resp, err := queryClient.AccountBonds(cmd.Context(), &incentive.QueryAccountBonds{Address: args[0]})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// QueryPendingRewards creates a Cobra command to query the pending incentive rewards of a single account.
func QueryPendingRewards() *cobra.Command {
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
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// QueryCurrentRates creates a Cobra command to query current annual rewards for a reference amount
// of a given bonded uToken.
func QueryCurrentRates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current-rates [denom]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "Query the current annual rewards for a reference amount of a given bonded uToken.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			denom := ""
			if len(args) > 0 {
				denom = args[0]
			}

			queryClient := incentive.NewQueryClient(clientCtx)
			resp, err := queryClient.CurrentRates(cmd.Context(), &incentive.QueryCurrentRates{UToken: denom})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// QueryActualRates creates a Cobra command to query current annual rewards for a reference amount
// of a given bonded uToken.
func QueryActualRates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "actual-rates [denom]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "Query the current annual rewards (as an APY) given bonded uToken.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			denom := ""
			if len(args) > 0 {
				denom = args[0]
			}

			queryClient := incentive.NewQueryClient(clientCtx)
			resp, err := queryClient.ActualRates(cmd.Context(), &incentive.QueryActualRates{UToken: denom})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// QueryTotalBonded creates a Cobra command to query bonded tokens across all users.
func QueryTotalBonded() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-bonded [denom]",
		Args:  cobra.RangeArgs(0, 1),
		Short: fmt.Sprintf("Query the total uTokens bonded to the %s module", incentive.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			denom := ""
			if len(args) > 0 {
				denom = args[0]
			}

			queryClient := incentive.NewQueryClient(clientCtx)
			resp, err := queryClient.TotalBonded(cmd.Context(), &incentive.QueryTotalBonded{Denom: denom})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// QueryTotalUnbonding creates a Cobra command to query unbonding tokens across all users.
func QueryTotalUnbonding() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-unbonding [denom]",
		Args:  cobra.RangeArgs(0, 1),
		Short: fmt.Sprintf("Query the total uTokens unbonding from the %s module", incentive.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			denom := ""
			if len(args) > 0 {
				denom = args[0]
			}

			queryClient := incentive.NewQueryClient(clientCtx)
			resp, err := queryClient.TotalUnbonding(cmd.Context(), &incentive.QueryTotalUnbonding{Denom: denom})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// QueryUpcomingIncentivePrograms creates a Cobra command to query for all upcoming
// incentive programs.
func QueryUpcomingIncentivePrograms() *cobra.Command {
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
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// QueryOngoingIncentivePrograms creates a Cobra command to query for all ongoing
// incentive programs.
func QueryOngoingIncentivePrograms() *cobra.Command {
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
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// QueryCompletedIncentivePrograms creates a Cobra command to query for all completed
// incentive programs.
func QueryCompletedIncentivePrograms() *cobra.Command {
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
			resp, err := queryClient.CompletedIncentivePrograms(cmd.Context(),
				&incentive.QueryCompletedIncentivePrograms{})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "completed incentive programs")
	return cmd
}

// QueryIncentiveProgram creates a Cobra command to query a single incentive program
// by its ID.
func QueryIncentiveProgram() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "program [id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query a single incentive program by its ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			id, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return err
			}

			queryClient := incentive.NewQueryClient(clientCtx)
			resp, err := queryClient.IncentiveProgram(cmd.Context(), &incentive.QueryIncentiveProgram{Id: uint32(id)})
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
