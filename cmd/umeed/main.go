package main

import (
	"os"
	"strings"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/cmd/umeed/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, strings.ToUpper(umeeapp.Name), umeeapp.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
