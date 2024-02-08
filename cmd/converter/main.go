package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "converter",
		Short: "a tool to convert addresses and other utilities",
	}
	cmd.AddCommand(bech32Cmd())
	return cmd
}
