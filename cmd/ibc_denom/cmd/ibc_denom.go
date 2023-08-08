package cmd

import (
	"fmt"

	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"github.com/spf13/cobra"
)

// ibcDenomCmd create the cli cmd for making ibc denom by base denom and channel-id
func ibcDenomCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibc-denom [base-denom] [channel-id]",
		Short: "Create an ibc denom by base denom and channel id",
		Args:  cobra.ExactArgs(2),
		Long: `Create an ibc denom by base denom and channel id.

Example:
$ ibc-denom uumee channel-22`,
		RunE: func(cmd *cobra.Command, args []string) error {

			ibcDenom, err := ibcDenom(args[0], args[1])
			if err != nil {
				return err
			}
			cmd.Println(ibcDenom)
			return nil
		},
	}

	return cmd
}

func ibcDenom(baseDenom, channelID string) (string, error) {
	denomTrace := ibctransfertypes.DenomTrace{
		Path:      fmt.Sprintf("transfer/%s", channelID),
		BaseDenom: baseDenom,
	}

	if err := denomTrace.Validate(); err != nil {
		return "", err
	}

	return denomTrace.IBCDenom(), nil
}
