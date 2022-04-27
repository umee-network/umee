package cmd

import (
	"os"
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm"
	bridgecmd "github.com/Gravity-Bridge/Gravity-Bridge/module/cmd/gravity/cmd"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingcli "github.com/cosmos/cosmos-sdk/x/auth/vesting/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/app/params"
)

// NewRootCmd returns the root command handler for the Umee daemon.
func NewRootCmd() (*cobra.Command, params.EncodingConfig) {
	encodingConfig := umeeapp.MakeEncodingConfig()
	moduleManager := umeeapp.ModuleBasics

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(umeeapp.DefaultNodeHome)

	rootCmd := &cobra.Command{
		Use:   umeeapp.Name + "d",
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
		umeeapp.DefaultNodeHome,
	)
	bridgeGenTxCmd.Use = strings.Replace(bridgeGenTxCmd.Use, "gentx", "gentx-gravity", 1)

	rootCmd.AddCommand(
		addGenesisAccountCmd(umeeapp.DefaultNodeHome),
		genutilcli.InitCmd(ac.moduleManager, umeeapp.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, umeeapp.DefaultNodeHome),
		genutilcli.MigrateGenesisCmd(),
		genutilcli.ValidateGenesisCmd(ac.moduleManager),
		genutilcli.GenTxCmd(
			ac.moduleManager,
			ac.encCfg.TxConfig,
			banktypes.GenesisBalancesIterator{},
			umeeapp.DefaultNodeHome,
		),
		bridgeGenTxCmd,
		tmcli.NewCompletionCmd(rootCmd, true),
		debugCmd(),
		GetWasmGenesisMsgCmd(umeeapp.DefaultNodeHome),
	)

	server.AddCommands(rootCmd, umeeapp.DefaultNodeHome, ac.newApp, ac.appExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(ac),
		txCommand(ac),
		keys.Commands(umeeapp.DefaultNodeHome),
	)
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
	wasm.AddModuleInitFlags(startCmd)
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
