package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v2/x/leverage/types"
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
		GetCmdSupply(),
		GetCmdWithdraw(),
		GetCmdCollateralize(),
		GetCmdDecollateralize(),
		GetCmdBorrow(),
		GetCmdRepay(),
		GetCmdLiquidate(),
	)

	return cmd
}

// GetCmdSupply creates a Cobra command to generate or broadcast a
// transaction with a MsgSupply message.
func GetCmdSupply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supply [supplier] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Supply a specified amount of a supported asset",
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

			msg := types.NewMsgSupply(clientCtx.GetFromAddress(), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdWithdraw creates a Cobra command to generate or broadcast a
// transaction with a MsgWithdraw message.
func GetCmdWithdraw() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [supplier] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Withdraw a specified amount of a supplied asset",
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

			msg := types.NewMsgWithdraw(clientCtx.GetFromAddress(), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdCollateralize creates a Cobra command to generate or broadcast a
// transaction with a MsgCollateralize message.
func GetCmdCollateralize() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collateralize [borrower] [coin]",
		Args:  cobra.ExactArgs(2),
		Short: "Add uTokens as collateral",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgCollateralize(clientCtx.GetFromAddress(), coin)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdDecollateralize returns a CLI command handler to generate or broadcast a
// transaction with a MsgDecollateralize message.
func GetCmdDecollateralize() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decollateralize [borrower] [coin]",
		Args:  cobra.ExactArgs(2),
		Short: "Remove uTokens from collateral",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[1])
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
		Use:   "borrow [borrower] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Borrow a specified amount of a supported asset",
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

			msg := types.NewMsgBorrow(clientCtx.GetFromAddress(), asset)

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
		Use:   "repay [borrower] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Repay a specified amount of a borrowed asset",
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

			msg := types.NewMsgRepay(clientCtx.GetFromAddress(), asset)

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
		Use:   "liquidate [liquidator] [borrower] [amount] [reward-denom]",
		Args:  cobra.ExactArgs(4),
		Short: "Liquidate a specified amount of a borrower's debt for a chosen reward denomination",
		Long: strings.TrimSpace(
			fmt.Sprintf(`
Liquidate up to a specified amount of a borrower's debt for a chosen reward denomination.

Example:
$ umeed tx leverage liquidate %s %s  50000000uumee u/uumee --from mykey`,
				"umee16jgsjqp7h0mpahlkw3p6vp90vd3jhn5tz6lcex",
				"umee1qqy7cst5qm83ldupph2dcq0wypprkfpc9l3jg2",
			),
		),

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			borrowerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			asset, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			rewardDenom := args[3]

			msg := types.NewMsgLiquidate(clientCtx.GetFromAddress(), borrowerAddr, asset, rewardDenom)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCmdSubmitUpdateRegistryProposal creates a Cobra command to generate
// or broadcast a transaction with a governance proposal message containing a
// UpdateRegistryProposal.
func NewCmdSubmitUpdateRegistryProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-registry [proposal-file] [deposit]",
		Args:  cobra.ExactArgs(2),
		Short: "Submit a update leverage registry governance proposal",
		Long: strings.TrimSpace(
			`Submit a leverage registry governance proposal along with an initial deposit.
The proposal details must be supplied via a JSON file. Please see the UpdateRegistryProposal
type for a complete description of the expected input.

Example:
$ umeed tx gov submit-proposal update-registry </path/to/proposal.json> <deposit> [flags...]

Where proposal.json contains:

{
  "title": "Update the Leverage Token Registry",
  "description": "Update the uumee token in the leverage registry.",
  "registry": [
    {
      "base_denom": "uumee",
      "reserve_factor": "0.1",
      "collateral_weight": "0.05",
      "liquidation_threshold": "0.05",
      "base_borrow_rate": "0.02",
      "kink_borrow_rate": "0.2",
      "max_borrow_rate": "1.5",
      "kink_utilization": "0.2",
      "liquidation_incentive": "0.1",
      "symbol_denom": "UMEE",
      "exponent": 6,
      "enable_msg_supply": true,
      "enable_msg_borrow": true,
      "blacklist": false
    },
    // ...
  ]
}
`,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := ParseUpdateRegistryProposal(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			content := types.NewUpdateRegistryProposal(proposal.Title, proposal.Description, proposal.Registry)

			msg, err := gov1b1.NewMsgSubmitProposal(content, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}
