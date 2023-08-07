package cmd

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/version"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"github.com/spf13/cobra"
)

// ibcDenom create an ibc denom by base denom and channel-id
func ibcDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibc-denom [base-denom] [channel-id]",
		Short: "Create an ibc denom by base denom and channel id",
		Args:  cobra.ExactArgs(2),
		Long: fmt.Sprintf(`Create an ibc denom by base denom and channel id.

Example:
$ %s ibc-denom uumee channel-22`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			baseDenom := args[0]
			channelID := args[1]

			denomTrace := ibctransfertypes.DenomTrace{
				Path:      fmt.Sprintf("transfer/%s", channelID),
				BaseDenom: baseDenom,
			}

			if err := denomTrace.Validate(); err != nil {
				return err
			}

			cmd.Println(denomTrace.IBCDenom())
			return nil
		},
	}

	return cmd
}
