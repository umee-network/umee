package main

import (
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/spf13/cobra"
)

func bech32Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bech32 bech32_source_address [bech32_dest_prefix]",
		Short: "Convert from bech32_source_address to a bech32 address with a given prefix (umee default)",
		Example: `converter bech32 juno1ey69r37gfxvxg62sh4r0ktpuc46pzjrm5cxnjg umee
converter bech32 cosmos130w5tzz97ks2cyhxrtlp7kpwnf97r742c89y9v`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dstPrefix := "umee"
			if len(args) == 2 {
				dstPrefix = args[1]
			}
			_, bz, err := bech32.DecodeAndConvert(args[0])
			if err != nil {
				return err
			}

			bech32Addr, err := bech32.ConvertAndEncode(dstPrefix, bz)
			if err != nil {
				return err
			}

			cmd.Println(bech32Addr)
			return nil
		},
	}

	return cmd
}
