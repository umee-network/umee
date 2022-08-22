package cmd

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	umeeapp "github.com/umee-network/umee/v2/app"
)

const (
	flagBech32HRP = "bech32-hrp"
)

// debugCmd returns a command handler for debugging addresses and public keys.
// It is based off of the SDK's debug command root handler with modified
// sub-commands.
func debugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Commands to aid in debugging addresses and public keys",
		RunE:  client.ValidateCmd,
	}

	cmd.AddCommand(debug.PubkeyCmd())
	cmd.AddCommand(debug.PubkeyRawCmd())
	cmd.AddCommand(debugAddrCmd())
	cmd.AddCommand(debug.RawBytesCmd())

	return cmd
}

func debugAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addr [address]",
		Short: "Convert an address between hex and bech32",
		Long: fmt.Sprintf(`Convert an address between hex encoding and bech32.

Example:
$ %s debug addr cosmos1e0jnq2sun3dzjh8p2xq95kk0expwmd7shwjpfg
			`, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addrStr := args[0]

			var (
				bz  []byte
				err error
			)

			// try HEX then Bech32
			bz, err = hex.DecodeString(addrStr)
			if err != nil {
				bech32HRP, err := cmd.Flags().GetString(flagBech32HRP)
				if err != nil {
					return err
				}

				bz, err = sdk.GetFromBech32(addrStr, bech32HRP)
				if err != nil {
					return errors.New("failed to decode address as HEX and Bech32")
				}
			}

			if err := sdk.VerifyAddressFormat(bz); err != nil {
				return fmt.Errorf("failed to verify converted address: %w", err)
			}

			cmd.Printf("Address (HEX): %X\n", bz)
			cmd.Printf("Address Bech32 Account: %s\n", sdk.AccAddress(bz))
			cmd.Printf("Address Bech32 Validator Operator: %s\n", sdk.ValAddress(bz))

			return nil
		},
	}

	cmd.Flags().String(flagBech32HRP, umeeapp.AccountAddressPrefix,
		"Input Bech32 HRP (use only when address input is a Bech32 address")
	return cmd
}
