package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/v6/x/auction"
)

// GetTxCmd returns the CLI transaction commands for the x/auction module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        auction.ModuleName,
		Short:                      fmt.Sprintf("Transaction commands for the %s module", auction.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		RewardsBid(),
	)

	return cmd
}

func RewardsBid() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rewards-bid [auction-id] [uumee-amount]",
		Args:    cobra.ExactArgs(2),
		Short:   "Places a bid for a rewards auction, auction-id must be an ID of the current auction",
		Example: "rewards-bid 12 10000uumee",
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			id, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || id < 0 {
				return errors.New("id argument must be a positive integer")
			}
			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			msg := auction.MsgRewardsBid{
				Sender: cctx.GetFromAddress().String(),
				Id:     uint32(id),
				Amount: coin}
			return tx.GenerateOrBroadcastTxCLI(cctx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
