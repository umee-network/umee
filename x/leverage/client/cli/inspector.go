package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v4/util/cli"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

// GetCmdQueryInspect creates a Cobra command to query for the inspector command.
func GetCmdQueryInspect() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect [symbol] [mode] [mode-min] [sort] [sort-min]",
		Args:    cobra.MinimumNArgs(3),
		Short:   "Inspect accounts with the leverage module.",
		Example: "umeed q leverage inspect OSMO LTV 0.75 collateral 100",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			modemin, err := strconv.ParseFloat(args[2], 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryInspect{
				Symbol:  args[0],
				Mode:    args[1],
				ModeMin: modemin,
			}
			if len(args) >= 4 {
				req.Sort = args[3]
			}
			if len(args) >= 5 {
				sortmin, err := strconv.ParseFloat(args[4], 64)
				if err != nil {
					return err
				}

				req.SortMin = sortmin
			}
			resp, err := queryClient.Inspect(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryInspectNeat creates a Cobra command to query for the inspector command.
func GetCmdQueryInspectNeat() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect-neat [symbol] [mode] [mode-min] [sort] [sort-min]",
		Args:    cobra.MinimumNArgs(3),
		Short:   "Inspect accounts with the leverage module.",
		Example: "umeed q leverage inspect-neat ALL borrowed 100 danger 0.5",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			modemin, err := strconv.ParseFloat(args[2], 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryInspectNeat{
				Symbol:  args[0],
				Mode:    args[1],
				ModeMin: modemin,
			}
			if len(args) >= 4 {
				req.Sort = args[3]
			}
			if len(args) >= 5 {
				sortmin, err := strconv.ParseFloat(args[4], 64)
				if err != nil {
					return err
				}

				req.SortMin = sortmin
			}
			resp, err := queryClient.InspectNeat(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryRiskData creates a Cobra command to query for the risk data command.
func GetCmdQueryRiskData() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "risk-data",
		Args:    cobra.ExactArgs(0),
		Short:   "Inspect accounts with the leverage module.",
		Example: "umeed q leverage risk-data",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryRiskData{}
			resp, err := queryClient.RiskData(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
