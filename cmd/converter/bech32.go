package main

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/spf13/cobra"
)

func bech32Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bech32 bech32_source_address1 bech32_source_address2...",
		Short: "Convert a list of bech32 source addresses",
		Example: `converter bech32 --dst-prefix umee juno1ey69r37gfxvxg62sh4r0ktpuc46pzjrm5cxnjg
converter bech32 cosmos130w5tzz97ks2cyhxrtlp7kpwnf97r742c89y9v cosmos1g3r2txznhs3l2vwe5sajaccgnmd3lcuhntnyug`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dstPrefix := cmd.Flag("dst-prefix").Value.String()
			for _, a := range args {
				_, bz, err := bech32.DecodeAndConvert(a)
				if err != nil {
					return fmt.Errorf("can't decode arg %q, %w", a, err)
				}
				bech32Addr, err := bech32.ConvertAndEncode(dstPrefix, bz)
				if err != nil {
					return fmt.Errorf("wrong dest prefix %q, %w", dstPrefix, err)
				}
				cmd.Println(bech32Addr)
			}
			return nil
		},
	}
	cmd.Flags().String("dst-prefix", "umee", "destination bech32 prefix")

	return cmd
}
