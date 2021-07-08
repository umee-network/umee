package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmconfig "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/umee-network/umee/app"
)

const (
	flagOutputDir         = "output-dir"
	flagNodeDirPrefix     = "node-dir-prefix"
	flagNodeDaemonHome    = "node-daemon-home"
	flagStartingIPAddress = "starting-ip-address"
	flagNumValidators     = "num-validators"
)

func localnetCmd(mbm module.BasicManager, genBalIterator banktypes.GenesisBalancesIterator) *cobra.Command {
	cmdCfg := viper.New()

	cmd := &cobra.Command{
		Use:   "localnet",
		Short: "Initialize files for an Umee local network",
		Long: `Create the necessary initialization files and directories needed to
bootstrap a configurable local Umee testing network. A certain number of
validators will be created, where for each validator we populate the necessary
files (e.g. priv_validator, genesis.json, config.toml, app.toml, etc...).
	
Note, strict routability for addresses is turned off in the config file.
	`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return cmdCfg.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			tmCfg := server.GetServerContextFromCmd(cmd).Config

			if err := initLocalnet(clientCtx, tmCfg, cmdCfg, mbm, genBalIterator, cmd.InOrStdin()); err != nil {
				return err
			}

			cmd.PrintErrf("Successfully bootstrapped %d validators!\n", cmdCfg.GetInt(flagNumValidators))
			return nil
		},
	}

	cmd.Flags().StringP(flagOutputDir, "o", "", "Directory to store initialization data")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Keyring backend to use (os|file|test)")
	cmd.Flags().String(flags.FlagChainID, "", "Chain ID to use for the local network (if empty, a random one will be generated)")
	cmd.Flags().String(server.FlagMinGasPrices, fmt.Sprintf("0.000006%s", app.BondDenom), "Minimum gas prices for validators to accept for transactions")
	cmd.Flags().String(flagNodeDirPrefix, "node", "Directory prefix for each validator (e.g. results in node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, "umeed", "Home directory of the node's configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1", "Starting IP address (192.168.0.1 results in persistent peers list <ID>@192.168.0.1:46656, <ID>@192.168.0.2:46656, ...)")
	cmd.Flags().Int(flagNumValidators, 4, "Number of validators to bootstrap")
	cmd.Flags().String(flags.FlagKeyAlgorithm, string(hd.Secp256k1Type), "Key signing algorithm for generated keys")

	return cmd
}

func initLocalnet(
	clientCtx client.Context,
	tmCfg *tmconfig.Config,
	cmdCfg *viper.Viper,
	mbm module.BasicManager,
	genBalIterator banktypes.GenesisBalancesIterator,
	reader io.Reader,
) error {

	chainID := cmdCfg.GetString(flags.FlagChainID)
	if chainID == "" {
		chainID = "chain-" + tmrand.NewRand().Str(6)
	}

	numValidators := cmdCfg.GetInt(flagNumValidators)
	nodeIDs := make([]string, numValidators)
	valPubKeys := make([]cryptotypes.PubKey, numValidators)

	umeeCfg := srvconfig.DefaultConfig()
	umeeCfg.MinGasPrices = cmdCfg.GetString(server.FlagMinGasPrices)
	umeeCfg.API.Enable = true
	umeeCfg.Telemetry.Enabled = true
	umeeCfg.Telemetry.PrometheusRetentionTime = 60
	umeeCfg.Telemetry.EnableHostnameLabel = false
	umeeCfg.Telemetry.GlobalLabels = [][]string{{"chain_id", chainID}}

	var (
		genAccounts []authtypes.GenesisAccount
		genBalances []banktypes.Balance
		genFiles    []string
	)

	inBuf := bufio.NewReader(reader)
	nodeDirPrefix := cmdCfg.GetString(flagNodeDirPrefix)
	outputDir := cmdCfg.GetString(flagOutputDir)
	nodeDaemonHome := cmdCfg.GetString(flagNodeDaemonHome)
	startingIPAddress := cmdCfg.GetString(flagStartingIPAddress)
	keyringBackend := cmdCfg.GetString(flags.FlagKeyringBackend)
	keyAlgo := cmdCfg.GetString(flags.FlagKeyAlgorithm)

	// generate private keys, node IDs, and initial transactions
	for i := 0; i < numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)

		tmCfg.SetRoot(nodeDir)
		tmCfg.RPC.ListenAddress = "tcp://0.0.0.0:26657"

		if err := os.MkdirAll(filepath.Join(nodeDir, "config"), 0755); err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		tmCfg.Moniker = nodeDirName

		ip, err := getIP(i, startingIPAddress)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		nodeIDs[i], valPubKeys[i], err = genutil.InitializeNodeValidatorFiles(tmCfg)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		memo := fmt.Sprintf("%s@%s:26656", nodeIDs[i], ip)
		genFiles = append(genFiles, tmCfg.GenesisFile())

		kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, nodeDir, inBuf)
		if err != nil {
			return err
		}

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(keyAlgo, keyringAlgos)
		if err != nil {
			return err
		}

		addr, secret, err := server.GenerateSaveCoinKey(kb, nodeDirName, true, algo)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		info := map[string]string{"secret": secret}

		keySeedBz, err := json.Marshal(info)
		if err != nil {
			return err
		}

		if err := writeFile("key_seed.json", nodeDir, keySeedBz); err != nil {
			return err
		}

		accTokens := sdk.TokensFromConsensusPower(1000)
		accStakingTokens := sdk.TokensFromConsensusPower(500)
		coins := sdk.Coins{
			sdk.NewCoin(fmt.Sprintf("%stoken", nodeDirName), accTokens),
			sdk.NewCoin(app.BondDenom, accStakingTokens),
		}

		genBalances = append(genBalances, banktypes.Balance{Address: addr.String(), Coins: coins.Sort()})
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(addr, nil, 0, 0))

		valTokens := sdk.TokensFromConsensusPower(100)
		createValMsg, err := stakingtypes.NewMsgCreateValidator(
			sdk.ValAddress(addr),
			valPubKeys[i],
			sdk.NewCoin(app.BondDenom, valTokens),
			stakingtypes.NewDescription(nodeDirName, "", "", "", ""),
			stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
			sdk.OneInt(),
		)
		if err != nil {
			return err
		}

		txBuilder := clientCtx.TxConfig.NewTxBuilder()
		if err := txBuilder.SetMsgs(createValMsg); err != nil {
			return err
		}

		txBuilder.SetMemo(memo)

		txFactory := tx.Factory{}
		txFactory = txFactory.
			WithChainID(chainID).
			WithMemo(memo).
			WithKeybase(kb).
			WithTxConfig(clientCtx.TxConfig)

		if err := tx.Sign(txFactory, nodeDirName, txBuilder, true); err != nil {
			return err
		}

		txBz, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			return err
		}

		genTxsDir := filepath.Join(outputDir, "gentxs")
		if err := writeFile(fmt.Sprintf("%v.json", nodeDirName), genTxsDir, txBz); err != nil {
			return err
		}

		srvconfig.WriteConfigFile(filepath.Join(nodeDir, "config/app.toml"), umeeCfg)
	}

	if err := initGenFiles(
		clientCtx,
		mbm,
		chainID,
		numValidators,
		genAccounts,
		genBalances,
		genFiles,
	); err != nil {
		return err
	}

	return collectGenFiles(
		clientCtx,
		tmCfg,
		nodeIDs,
		valPubKeys,
		numValidators,
		chainID,
		outputDir,
		nodeDirPrefix,
		nodeDaemonHome,
		genBalIterator,
	)
}

func initGenFiles(
	clientCtx client.Context,
	mbm module.BasicManager,
	chainID string,
	numValidators int,
	genAccounts []authtypes.GenesisAccount,
	genBalances []banktypes.Balance,
	genFiles []string,
) error {

	appGenState := mbm.DefaultGenesis(clientCtx.JSONMarshaler)

	// set the accounts in the genesis state
	var authGenState authtypes.GenesisState
	if err := clientCtx.JSONMarshaler.UnmarshalJSON(appGenState[authtypes.ModuleName], &authGenState); err != nil {
		return err
	}

	accounts, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		return err
	}

	authGenState.Accounts = accounts
	bz, err := clientCtx.JSONMarshaler.MarshalJSON(&authGenState)
	if err != nil {
		return err
	}

	appGenState[authtypes.ModuleName] = bz

	// set the balances in the genesis state
	var bankGenState banktypes.GenesisState
	if err := clientCtx.JSONMarshaler.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState); err != nil {
		return err
	}

	bankGenState.Balances = banktypes.SanitizeGenesisBalances(genBalances)
	for _, bal := range bankGenState.Balances {
		bankGenState.Supply = bankGenState.Supply.Add(bal.Coins...)
	}

	bz, err = clientCtx.JSONMarshaler.MarshalJSON(&bankGenState)
	if err != nil {
		return err
	}

	appGenState[banktypes.ModuleName] = bz

	appGenStateJSON, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc := types.GenesisDoc{
		ChainID:    chainID,
		AppState:   appGenStateJSON,
		Validators: nil,
	}

	// generate empty genesis files for each validator and save
	for i := 0; i < numValidators; i++ {
		if err := genDoc.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}

	return nil
}

func collectGenFiles(
	clientCtx client.Context,
	tmCfg *tmconfig.Config,
	nodeIDs []string,
	valPubKeys []cryptotypes.PubKey,
	numValidators int,
	chainID, outputDir, nodeDirPrefix, nodeDaemonHome string,
	genBalIterator banktypes.GenesisBalancesIterator,
) error {

	var appState json.RawMessage
	genTime := tmtime.Now()

	for i := 0; i < numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		gentxsDir := filepath.Join(outputDir, "gentxs")
		tmCfg.Moniker = nodeDirName

		tmCfg.SetRoot(nodeDir)

		nodeID, valPubKey := nodeIDs[i], valPubKeys[i]
		initCfg := genutiltypes.NewInitConfig(chainID, gentxsDir, nodeID, valPubKey)

		genDoc, err := types.GenesisDocFromFile(tmCfg.GenesisFile())
		if err != nil {
			return err
		}

		nodeAppState, err := genutil.GenAppStateFromConfig(
			clientCtx.JSONMarshaler,
			clientCtx.TxConfig,
			tmCfg,
			initCfg,
			*genDoc,
			genBalIterator,
		)
		if err != nil {
			return err
		}

		if appState == nil {
			// set the canonical application state (they should not differ)
			appState = nodeAppState
		}

		genFile := tmCfg.GenesisFile()

		// overwrite each validator's genesis file to have a canonical genesis time
		if err := genutil.ExportGenesisFileWithTime(genFile, chainID, nil, appState, genTime); err != nil {
			return err
		}
	}

	return nil
}

func getIP(i int, startingIPAddr string) (ip string, err error) {
	if len(startingIPAddr) == 0 {
		ip, err = server.ExternalIP()
		if err != nil {
			return "", err
		}

		return ip, nil
	}

	ipv4 := net.ParseIP(startingIPAddr).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("%v: non ipv4 address", ip)
	}

	for j := 0; j < i; j++ {
		ipv4[3]++
	}

	return ipv4.String(), nil
}

func writeFile(name string, dir string, contents []byte) error {
	writePath := filepath.Join(dir)
	file := filepath.Join(writePath, name)

	err := tmos.EnsureDir(writePath, 0755)
	if err != nil {
		return err
	}

	err = tmos.WriteFile(file, contents, 0644)
	if err != nil {
		return err
	}

	return nil
}
