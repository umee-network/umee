package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v6/util/cli"
	"github.com/umee-network/umee/v6/x/auction"
)

// GetQueryCmd returns the CLI query commands for the x/auction module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        auction.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", auction.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		RewardsParams(),
		RewardsAuction(),
	)

	return cmd
}

func RewardsParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards-params",
		Args:  cobra.NoArgs,
		Short: "Query x/auction rewards params",
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx, q, err := prepareQueryCtx(cmd)
			if err != nil {
				return err
			}

			req := &auction.QueryRewardsParams{}
			resp, err := q.RewardsParams(cmd.Context(), req)
			return cli.PrintOrErr(resp, err, cctx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func RewardsAuction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards-auction [id]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Query rewards auction state",
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx, q, err := prepareQueryCtx(cmd)
			if err != nil {
				return err
			}

			req := auction.QueryRewardsAuction{}
			if len(args) > 0 {
				id, err := strconv.ParseInt(args[0], 10, 32)
				if err != nil || id < 0 {
					return errors.New("id argument must be a positive integer")
				}
				req.Id = uint32(id)
			}
			resp, err := q.RewardsAuction(cmd.Context(), &req)
			return cli.PrintOrErr(resp, err, cctx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func prepareQueryCtx(cmd *cobra.Command) (client.Context, auction.QueryClient, error) {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return clientCtx, nil, err
	}
	return clientCtx, auction.NewQueryClient(clientCtx), err
}
