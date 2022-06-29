package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
		GetCmdLendAsset(),
		GetCmdWithdrawAsset(),
		GetCmdAddCollateral(),
		GetCmdRemoveCollateral(),
		GetCmdBorrowAsset(),
		GetCmdRepayAsset(),
		GetCmdLiquidate(),
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

			msg := types.NewMsgWithdrawAsset(clientCtx.GetFromAddress(), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdAddCollateral returns a CLI command handler to generate or broadcast a
// transaction with a MsgAddCollateral message.
func GetCmdAddCollateral() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-collateral [borrower] [coin]",
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

			asset, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgAddCollateral(clientCtx.GetFromAddress(), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdRemoveCollateral returns a CLI command handler to generate or broadcast a
// transaction with a MsgRemoveCollateral message.
func GetCmdRemoveCollateral() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-collateral [borrower] [coin]",
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

			asset, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgRemoveCollateral(clientCtx.GetFromAddress(), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdBorrowAsset returns a CLI command handler to generate or broadcast a
// transaction with a MsgBorrowAsset message.
func GetCmdBorrowAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "borrow-asset [borrower] [amount]",
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

			msg := types.NewMsgBorrowAsset(clientCtx.GetFromAddress(), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdRepayAsset returns a CLI command handler to generate or broadcast a
// transaction with a MsgRepayAsset message.
func GetCmdRepayAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repay-asset [borrower] [amount]",
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

			msg := types.NewMsgRepayAsset(clientCtx.GetFromAddress(), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdLiquidate returns a CLI command handler to generate or broadcast a
// transaction with a MsgLiquidate message.
func GetCmdLiquidate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidate [liquidator] [borrower] [amount] [reward]",
		Args:  cobra.ExactArgs(4),
		Short: "Liquidate a specified amount of a borrower's debt for a chosen reward denomination",
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

			reward, err := sdk.ParseCoinNormalized(args[3])
			if err != nil {
				return err
			}

			msg := types.NewMsgLiquidate(clientCtx.GetFromAddress(), borrowerAddr, asset, reward)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCmdSubmitUpdateRegistryProposal returns a CLI command handler to generate
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
      "enable_msg_lend": true,
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

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}
