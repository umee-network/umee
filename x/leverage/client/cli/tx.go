package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/x/leverage/types"
)

// GetTxCmd returns the CLI transaction commands for the x/leverage module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Transaction commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdLendAsset(),
		GetCmdWithdrawAsset(),
		GetCmdSetCollateral(),
	)

	return cmd
}

// GetCmdLendAsset returns a CLI command handler to generate or broadcast a
// transaction with a MsgLendAsset message.
func GetCmdLendAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lend-asset [lender] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Lend a specified amount of a supported asset",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			asset, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgLendAsset(clientCtx.GetFromAddress(), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdWithdrawAsset returns a CLI command handler to generate or broadcast a
// transaction with a MsgWithdrawAsset message.
func GetCmdWithdrawAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-asset [lender] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Withdraw a specified amount of a loaned supported asset",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			asset, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawAsset(clientCtx.GetFromAddress(), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdSetCollateral returns a CLI command handler to generate or broadcast a
// transaction with a MsgSetCollateral message.
func GetCmdSetCollateral() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-collateral [borrower] [denom] [toggle]",
		Args:  cobra.ExactArgs(3),
		Short: "Enable or disable an asset type to be used as collateral",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			toggle, err := strconv.ParseBool(args[2])
			if err != nil {
				return fmt.Errorf("failed to parse toggle: %w", err)
			}

			msg := types.NewMsgSetCollateral(clientCtx.GetFromAddress(), args[1], toggle)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
