package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/v6/x/uibc"
)

// GetTxCmd returns the CLI transaction commands for the x/uibc module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        uibc.ModuleName,
		Short:                      fmt.Sprintf("Transaction commands for the %s module", uibc.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	return cmd
}
