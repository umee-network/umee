package main

import (
	"os"

	"github.com/umee-network/umee/v5/cmd/ibc_denom/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
