package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v5/x/refileverage/types"
)

// GetTxCmd returns the CLI transaction commands for the x/refileverage module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Transaction commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdSupplyCollateral(),
		GetCmdDecollateralize(),
		GetCmdBorrow(),
		GetCmdRepay(),
		GetCmdLiquidate(),
	)

	return cmd
}

// GetCmdDecollateralize returns a CLI command handler to generate or broadcast a
// transaction with a MsgDecollateralize message.
func GetCmdDecollateralize() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decollateralize [coin]",
		Args:  cobra.ExactArgs(1),
		Short: "Remove uTokens from collateral",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgDecollateralize(clientCtx.GetFromAddress(), coin)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdBorrow creates a Cobra command to generate or broadcast a
// transaction with a MsgBorrow message.
func GetCmdBorrow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "borrow [amount] [eth recipient]",
		Args:  cobra.ExactArgs(2),
		Short: "Borrow a specified amount of a supported asset",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, ok := sdk.NewIntFromString(args[0])
			if !ok {
				return fmt.Errorf("amount must be a positive integer")
			}

			msg, err := types.NewMsgBorrow(clientCtx.GetFromAddress(), amount, args[1])
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdRepay creates a Cobra command to generate or broadcast a
// transaction with a MsgRepay message.
func GetCmdRepay() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repay [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "Repay a specified amount of a borrowed asset",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, ok := sdk.NewIntFromString(args[0])
			if !ok || !amount.IsPositive() {
				return fmt.Errorf("amount must be a integer number")
			}

			msg := types.NewMsgRepay(clientCtx.GetFromAddress(), amount)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdLiquidate creates a Cobra command to generate or broadcast a
// transaction with a MsgLiquidate message.
func GetCmdLiquidate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidate [borrower] [amount] [reward-denom]",
		Args:  cobra.ExactArgs(3),
		Short: "Liquidate a specified amount of a borrower's debt for a chosen reward denomination",
		Long: strings.TrimSpace(
			fmt.Sprintf(`
Liquidate up to a specified amount of a borrower's debt for a chosen reward denomination.

Example:
$ umeed tx refileverage liquidate %s  50000000uumee u/uumee --from mykey`,
				"umee1qqy7cst5qm83ldupph2dcq0wypprkfpc9l3jg2",
			),
		),

		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			borrowerAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			amount, ok := sdk.NewIntFromString(args[0])
			if !ok || !amount.IsPositive() {
				return fmt.Errorf("amount must be a integer number")
			}

			rewardDenom := args[2]

			msg := types.NewMsgLiquidate(clientCtx.GetFromAddress(), borrowerAddr, amount, rewardDenom)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdSupplyCollateral creates a Cobra command to generate or broadcast a
// transaction with a MsgSupply message.
func GetCmdSupplyCollateral() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supply-collateral [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "Supply and collateralize a specified amount of a supported asset",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			asset, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSupplyCollateral(clientCtx.GetFromAddress(), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
