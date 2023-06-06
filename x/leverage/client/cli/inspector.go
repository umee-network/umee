package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v5/util/cli"
	"github.com/umee-network/umee/v5/x/leverage/types"
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

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryInspect{
				Opts: &types.InspectOptions{
					Symbol:  args[0],
					Mode:    args[1],
					ModeMin: sdk.MustNewDecFromStr(args[2]),
				},
			}
			if len(args) >= 4 {
				req.Opts.Sort = args[3]
			}
			if len(args) >= 5 {
				req.Opts.SortMin = sdk.MustNewDecFromStr(args[4])
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

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryInspectNeat{
				Opts: &types.InspectOptions{
					Symbol:  args[0],
					Mode:    args[1],
					ModeMin: sdk.MustNewDecFromStr(args[2]),
				},
			}
			if len(args) >= 4 {
				req.Opts.Sort = args[3]
			}
			if len(args) >= 5 {
				req.Opts.SortMin = sdk.MustNewDecFromStr(args[4])
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
