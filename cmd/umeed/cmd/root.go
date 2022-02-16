package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	bridgecmd "github.com/Gravity-Bridge/Gravity-Bridge/module/cmd/gravity/cmd"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingcli "github.com/cosmos/cosmos-sdk/x/auth/vesting/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/umee-network/umee/app"
	umeeappbeta "github.com/umee-network/umee/app/beta"
	"github.com/umee-network/umee/app/params"
)

// EnableBeta defines an ldflag that enables the beta version of the application
// to be built.
var EnableBeta string

// NewRootCmd returns the root command handler for the Umee daemon.
func NewRootCmd() (*cobra.Command, params.EncodingConfig) {
	var beta bool

	switch {
	case len(EnableBeta) > 0:
		// Handle the case where a build flag is provided, which used when building
		// the binary.
		v, err := strconv.ParseBool(EnableBeta)
		if err != nil {
			panic(fmt.Sprintf("failed to parse EnableBeta build flag: %s", err))
		}

		beta = v

	case len(os.Getenv("UMEE_ENABLE_BETA")) > 0:
		// Handle the case where an env var is provided, which is used when running
		// with Starport where we cannot control build flags/inputs.
		v, err := strconv.ParseBool(os.Getenv("UMEE_ENABLE_BETA"))
		if err != nil {
			panic(fmt.Sprintf("failed to parse env var 'UMEE_ENABLE_BETA': %s", err))
		}

		beta = v
	}

	var (
		encodingConfig params.EncodingConfig
		moduleManager  module.BasicManager
	)
	if beta {
		encodingConfig = umeeappbeta.MakeEncodingConfig()
		moduleManager = umeeappbeta.ModuleBasics
	} else {
		encodingConfig = app.MakeEncodingConfig()
		moduleManager = app.ModuleBasics
	}

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(app.DefaultNodeHome)

	rootCmd := &cobra.Command{
		Use:   app.Name + "d",
		Short: "Umee application network daemon and client",
		Long: `A daemon and client for interacting with the Umee network. Umee is a
Universal Capital Facility that can collateralize assets on one blockchain
towards borrowing assets on another blockchain.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			return server.InterceptConfigsPreRunHandler(cmd, "", nil)
		},
	}

	ac := appCreator{
		encCfg:        encodingConfig,
		moduleManager: moduleManager,
		beta:          beta,
	}

	initRootCmd(rootCmd, ac)

	return rootCmd, encodingConfig
}

func initRootCmd(rootCmd *cobra.Command, ac appCreator) {
	// We allow two variants of the gentx command:
	//
	// 1. The standard one provided by the SDK, mainly motivated for testing
	// and local network purposes.
	// 2. The Gravity Bridge variant which allows validators to provide key
	// delegation material.
	bridgeGenTxCmd := bridgecmd.GenTxCmd(
		ac.moduleManager,
		ac.encCfg.TxConfig,
		banktypes.GenesisBalancesIterator{},
		app.DefaultNodeHome,
	)
	bridgeGenTxCmd.Use = strings.Replace(bridgeGenTxCmd.Use, "gentx", "gentx-gravity", 1)

	rootCmd.AddCommand(
		addGenesisAccountCmd(app.DefaultNodeHome),
		genutilcli.InitCmd(ac.moduleManager, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		genutilcli.MigrateGenesisCmd(),
		genutilcli.ValidateGenesisCmd(ac.moduleManager),
		genutilcli.GenTxCmd(
			ac.moduleManager,
			ac.encCfg.TxConfig,
			banktypes.GenesisBalancesIterator{},
			app.DefaultNodeHome,
		),
		bridgeGenTxCmd,
		tmcli.NewCompletionCmd(rootCmd, true),
		debugCmd(),
	)

	server.AddCommands(rootCmd, app.DefaultNodeHome, ac.newApp, ac.appExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(ac),
		txCommand(ac),
		keys.Commands(app.DefaultNodeHome),
	)
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

func queryCommand(ac appCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying sub-commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	ac.moduleManager.AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func txCommand(ac appCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions sub-commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		flags.LineBreak,
		vestingcli.GetTxCmd(),
	)

	ac.moduleManager.AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

// nolint: unused
func overwriteFlagDefaults(c *cobra.Command, defaults map[string]string) {
	set := func(s *pflag.FlagSet, key, val string) {
		if f := s.Lookup(key); f != nil {
			f.DefValue = val
			_ = f.Value.Set(val)
		}
	}

	for key, val := range defaults {
		set(c.Flags(), key, val)
		set(c.PersistentFlags(), key, val)
	}

	for _, c := range c.Commands() {
		overwriteFlagDefaults(c, defaults)
	}
}
