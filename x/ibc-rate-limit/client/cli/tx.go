package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/types"
)

// GetTxCmd returns the CLI transaction commands for the x/ibc-rate-limit module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Transaction commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		SampleTOkenRegistration(),
	)

	return cmd
}

// TODO: remove after test
func SampleTOkenRegistration() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Args:  cobra.NoArgs,
		Short: "Supply a specified amount of a supported asset",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			addRateLimits := []types.MsgRateLimit{
				{
					IbcDenom:     "sai",
					OutflowLimit: 1234,
					TimeWindow:   1000,
				},
				{
					IbcDenom:     "raj",
					OutflowLimit: 1234,
					TimeWindow:   1000,
				},
			}

			msg := types.NewIbcDenomsRateLimits(clientCtx.FromAddress.String(), "title", "desc", addRateLimits, nil)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
