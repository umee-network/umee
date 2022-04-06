package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/cmd/umeed/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, umeeapp.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
