package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v4/x/incentive"
)

// GetTxCmd returns the CLI transaction commands for the x/incentive module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        incentive.ModuleName,
		Short:                      fmt.Sprintf("Transaction commands for the %s module", incentive.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdClaim(),
		GetCmdBond(),
		GetCmdBeginUnbonding(),
		GetCmdSponsor(),
	)

	return cmd
}

// GetCmdClaim creates a Cobra command to generate or broadcast a
// transaction with a MsgClaim message.
func GetCmdClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim",
		Args:  cobra.ExactArgs(0),
		Short: "Claim any pending incentive rewards",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := incentive.NewMsgClaim(clientCtx.GetFromAddress())

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// transaction with a MsgBond message.
func GetCmdBond() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bond [tier] [utokens]",
		Args:  cobra.ExactArgs(2),
		Short: "Bond some uToken collateral to a specific unbonding tier",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			asset, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			tier, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return err
			}

			msg := incentive.NewMsgBond(clientCtx.GetFromAddress(), uint32(tier), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// transaction with a MsgBeginUnbonding message.
func GetCmdBeginUnbonding() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "begin-unbonding [tier] [utokens]",
		Args:  cobra.ExactArgs(2),
		Short: "Begin unbonding some currently bonded utokens from a specifc unbonding tier",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			asset, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			tier, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return err
			}

			msg := incentive.NewMsgBeginUnbonding(clientCtx.GetFromAddress(), uint32(tier), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// transaction with a MsgSponsor message.
func GetCmdSponsor() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sponsor [program-id] [tokens]",
		Args:  cobra.ExactArgs(2),
		Short: "Fund a governance-approved, not yet funded incentive program with its exact total reward tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			asset, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return err
			}

			msg := incentive.NewMsgSponsor(clientCtx.GetFromAddress(), uint32(id), asset)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
