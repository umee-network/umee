package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/v5/x/metoken"
)

// GetTxCmd returns the CLI transaction commands for the x/metoken module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        metoken.ModuleName,
		Short:                      fmt.Sprintf("Transaction commands for the %s module", metoken.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdSwap(),
		GetCmdRedeem(),
	)

	return cmd
}

// GetCmdSwap creates a Cobra command to generate or broadcast a transaction with a MsgSwap message.
// Both arguments are required.
func GetCmdSwap() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap [coin] [metoken_denom]",
		Args:  cobra.ExactArgs(2),
		Short: "swap a specified amount of an accepted asset for the selected meToken",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			asset, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := metoken.NewMsgSwap(clientCtx.GetFromAddress(), asset, args[1])
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRedeem creates a Cobra command to generate or broadcast a transaction with a MsgRedeem message.
// Both arguments are required.
func GetCmdRedeem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeem [metoken] [redeem_denom]",
		Args:  cobra.ExactArgs(2),
		Short: "redeem a specified amount of meToken for the selected asset accepted by the index",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			meToken, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := metoken.NewMsgRedeem(clientCtx.GetFromAddress(), meToken, args[1])
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
