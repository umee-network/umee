package cmd

import (
	"os"
	"strings"

	bridgecmd "github.com/Gravity-Bridge/Gravity-Bridge/module/cmd/gravity/cmd"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	tmcfg "github.com/tendermint/tendermint/config"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	umeeapp "github.com/umee-network/umee/v3/app"
	"github.com/umee-network/umee/v3/app/params"
)

// NewRootCmd returns the root command handler for the Umee daemon.
func NewRootCmd() (*cobra.Command, params.EncodingConfig) {
	encodingConfig := umeeapp.MakeEncodingConfig()
	moduleManager := umeeapp.ModuleBasics

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(umeeapp.DefaultNodeHome).
		WithViper(umeeapp.Name)

	rootCmd := &cobra.Command{
		Use:   umeeapp.Name + "d",
		Short: "Umee application network daemon and client",
		Long: `A daemon and client for interacting with the Umee network. Umee is a
Universal Capital Facility that can collateralize assets on one blockchain
towards borrowing assets on another blockchain.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			tmConfig := initTendermintConfig()
			return server.InterceptConfigsPreRunHandler(cmd, "", nil, tmConfig)
		},
	}

	ac := appCreator{
		encCfg:        encodingConfig,
		moduleManager: moduleManager,
	}

	initRootCmd(rootCmd, ac)

	return rootCmd, encodingConfig
}

// initTendermintConfig helps to override default Tendermint Config values.
// return tmcfg.DefaultConfig if no custom configuration is required for the application.
func initTendermintConfig() *tmcfg.Config {
	cfg := tmcfg.DefaultConfig()

	// these values put a higher strain on node memory
	// cfg.P2P.MaxNumInboundPeers = 100
	// cfg.P2P.MaxNumOutboundPeers = 40

	return cfg
}

func initRootCmd(rootCmd *cobra.Command, a appCreator) {
	// We allow two variants of the gentx command:
	//
	// 1. The standard one provided by the SDK, mainly motivated for testing
	// and local network purposes.
	// 2. The Gravity Bridge variant which allows validators to provide key
	// delegation material.
	bridgeGenTxCmd := bridgecmd.GenTxCmd(
		a.moduleManager,
		a.encCfg.TxConfig,
		banktypes.GenesisBalancesIterator{},
		umeeapp.DefaultNodeHome,
	)
	bridgeGenTxCmd.Use = strings.Replace(bridgeGenTxCmd.Use, "gentx", "gentx-gravity", 1)
	gentxModule := umeeapp.ModuleBasics[genutiltypes.ModuleName].(genutil.AppModuleBasic)

	rootCmd.AddCommand(
		genutilcli.InitCmd(a.moduleManager, umeeapp.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, umeeapp.DefaultNodeHome, gentxModule.GenTxValidator),
		genutilcli.MigrateGenesisCmd(),
		genutilcli.GenTxCmd(
			a.moduleManager,
			a.encCfg.TxConfig,
			banktypes.GenesisBalancesIterator{},
			umeeapp.DefaultNodeHome,
		),
		bridgeGenTxCmd,
		genutilcli.ValidateGenesisCmd(a.moduleManager),
		addGenesisAccountCmd(umeeapp.DefaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		debugCmd(),
		config.Cmd(),
	)

	server.AddCommands(rootCmd, umeeapp.DefaultNodeHome, a.newApp, a.appExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(a),
		txCommand(a),
		keys.Commands(umeeapp.DefaultNodeHome),
	)

	// add rosetta
	rootCmd.AddCommand(server.RosettaCommand(a.encCfg.InterfaceRegistry, a.encCfg.Codec))
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

func queryCommand(ac appCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
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
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetAuxToFeeCommand(),
	)

	ac.moduleManager.AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}
