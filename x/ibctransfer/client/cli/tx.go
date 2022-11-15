package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/v3/x/ibctransfer"
)

// GetTxCmd returns the CLI transaction commands for the x/ibctransfer module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        ibctransfer.ModuleName,
		Short:                      fmt.Sprintf("Transaction commands for the %s module", ibctransfer.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand()

	return cmd
}
