package cmd

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"time"

	peggysol "github.com/InjectiveLabs/peggo/solidity/wrappers/Peggy.sol"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"

	peggytypes "github.com/umee-network/umee/x/peggy/types"
)

const (
	flagEthNode        = "eth-node"
	flagEthPrivKey     = "eth-priv-key"
	flagEthGasPrice    = "eth-gas-price"
	flagEthGasLimit    = "eth-gas-limit"
	flagPowerThreshold = "power-threshold"
)

func createBridgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridge",
		Short: "Commands to interface with the Ethereum Peggy (Gravity Bridge)",
	}

	cmd.AddCommand(
		deployPeggyCmd(),
		initPeggyCmd(),
		// TODO:
		// - Deploy ERC20 command
	)

	return cmd
}

// TODO: Support --admin capabilities
func deployPeggyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-peggy",
		Short: "Deploy the Peggy (Gravity Bridge) smart contract on Ethereum",
		RunE: func(cmd *cobra.Command, args []string) error {
			ethNode, err := cmd.Flags().GetString(flagEthNode)
			if err != nil {
				return err
			}

			ethClient, err := ethclient.Dial(ethNode)
			if err != nil {
				return fmt.Errorf("failed to dial Ethereum node: %w", err)
			}

			auth, err := buildTransactOpts(cmd, ethClient)
			if err != nil {
				return err
			}

			address, tx, _, err := peggysol.DeployPeggy(auth, ethClient)
			if err != nil {
				return fmt.Errorf("failed deploy Peggy (Gravity Bridge) contract: %w", err)
			}

			_, _ = fmt.Fprintf(os.Stderr,
				"Peggy (Gravity Bridge) contract successfully deployed!\nAddress: %s\nTransaction: %s\n",
				address.Hex(),
				tx.Hash().Hex(),
			)

			return nil
		},
	}

	cmd.Flags().String(flagEthNode, "http://localhost:8545", "The network address of an Ethereum node")
	cmd.Flags().String(flagEthPrivKey, "", "The hex-encoded private key of an Ethereum account")
	cmd.Flags().Uint64(flagEthGasPrice, 0, "The Ethereum gas price to include in the transaction")
	cmd.Flags().Uint64(flagEthGasLimit, 6000000, "The Ethereum gas limit to include in the transaction")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func initPeggyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-peggy [peggy-address]",
		Args:  cobra.ExactArgs(1),
		Short: "Initialize the Peggy (Gravity Bridge) smart contract on Ethereum",
		Long: `Initialize the Peggy (Gravity Bridge) smart contract on Ethereum using
the current validator set and their respective powers.

Note, each validator must have their Ethereum delegate keys registered on chain
prior to initializing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			peggyQueryClient := peggytypes.NewQueryClient(clientCtx)

			ethNode, err := cmd.Flags().GetString(flagEthNode)
			if err != nil {
				return err
			}

			ethClient, err := ethclient.Dial(ethNode)
			if err != nil {
				return fmt.Errorf("failed to dial Ethereum node: %w", err)
			}

			contract, err := peggysol.NewPeggy(ethcommon.HexToAddress(args[0]), ethClient)
			if err != nil {
				return fmt.Errorf("failed to create Peggy contract instance: %w", err)
			}

			auth, err := buildTransactOpts(cmd, ethClient)
			if err != nil {
				return err
			}

			powerThresholdInt, err := cmd.Flags().GetUint64(flagPowerThreshold)
			if err != nil {
				return err
			}

			powerThreshold := new(big.Int).SetUint64(powerThresholdInt)

			peggyParams, err := peggyQueryClient.Params(cmd.Context(), &peggytypes.QueryParamsRequest{})
			if err != nil {
				return fmt.Errorf("failed to query for Peggy params: %w", err)
			}

			var peggyID [32]byte
			copy(peggyID[:], peggyParams.Params.PeggyId)

			currValSet, err := peggyQueryClient.CurrentValset(cmd.Context(), &peggytypes.QueryCurrentValsetRequest{})
			if err != nil {
				return err
			}

			var (
				validators = make([]ethcommon.Address, len(currValSet.Valset.Members))
				powers     = make([]*big.Int, len(currValSet.Valset.Members))

				totalPower uint64
			)
			for i, member := range currValSet.Valset.Members {
				validators[i] = ethcommon.HexToAddress(member.EthereumAddress)
				powers[i] = new(big.Int).SetUint64(member.Power)
				totalPower += member.Power
			}

			if totalPower < powerThresholdInt {
				return fmt.Errorf(
					"refusing to deploy; total power (%d) < power threshold (%d)",
					totalPower, powerThresholdInt,
				)
			}

			tx, err := contract.Initialize(auth, peggyID, powerThreshold, validators, powers)
			if err != nil {
				return fmt.Errorf("failed to initialize Peggy (Gravity Bridge): %w", err)
			}

			_, _ = fmt.Fprintf(os.Stderr,
				"Peggy (Gravity Bridge) contract successfully initialized!\nTransaction: %s\n",
				tx.Hash().Hex(),
			)

			return nil
		},
	}

	cmd.Flags().String(flagEthNode, "http://localhost:8545", "The network address of an Ethereum node")
	cmd.Flags().String(flagEthPrivKey, "", "The hex-encoded private key of an Ethereum account")
	cmd.Flags().Uint64(flagEthGasPrice, 0, "The Ethereum gas price to include in the transaction")
	cmd.Flags().Uint64(flagEthGasLimit, 6000000, "The Ethereum gas limit to include in the transaction")
	cmd.Flags().Uint64(flagPowerThreshold, 2834678415, "The validator power threshold to initialize Peggy with")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func buildTransactOpts(cmd *cobra.Command, ethClient *ethclient.Client) (*bind.TransactOpts, error) {
	ethPrivKeyHexStr, err := cmd.Flags().GetString(flagEthPrivKey)
	if err != nil {
		return nil, err
	}

	privKey, err := ethcrypto.HexToECDSA(ethPrivKeyHexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid public key; expected: %T, got: %T", &ecdsa.PublicKey{}, publicKey)
	}

	goCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fromAddress := ethcrypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := ethClient.PendingNonceAt(goCtx, fromAddress)
	if err != nil {
		return nil, err
	}

	goCtx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ethChainID, err := ethClient.ChainID(goCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Ethereum chain ID: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, ethChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ethereum transactor: %w", err)
	}

	var gasPrice *big.Int

	gasPriceInt, err := cmd.Flags().GetUint64(flagEthGasPrice)
	switch {
	case err != nil:
		return nil, err

	case gasPriceInt > 0:
		gasPrice = new(big.Int).SetUint64(gasPriceInt)

	default:
		gasPrice, err = ethClient.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get Ethereum gas estimate: %w", err)
		}
	}

	gasLimit, err := cmd.Flags().GetUint64(flagEthGasLimit)
	if err != nil {
		return nil, err
	}

	auth.Nonce = new(big.Int).SetUint64(nonce)
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit   // in units
	auth.GasPrice = gasPrice

	return auth, nil
}
