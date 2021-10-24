package cmd

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/ethbinding/peggy"
)

const (
	flagEthNode     = "eth-node"
	flagEthPrivKey  = "eth-priv-key"
	flagEthGasPrice = "eth-gas-price"
	flagEthGasLimit = "eth-gas-limit"
)

func deployPeggyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-peggy",
		Short: "Deploy and initialize the Peggy (Gravity Bridge) smart contract on Ethereum",
		Long: `Deploy and initialize the Peggy (Gravity Bridge) smart contract on Ethereum.
The contract will be initialized with the current active set of validators and their
respective powers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// clientCtx := client.GetClientContextFromCmd(cmd)

			ethNode, err := cmd.Flags().GetString(flagEthNode)
			if err != nil {
				return err
			}

			ethClient, err := ethclient.Dial(ethNode)
			if err != nil {
				return fmt.Errorf("failed to dial Ethereum node: %w", err)
			}

			ethPrivKeyHexStr, err := cmd.Flags().GetString(flagEthPrivKey)
			if err != nil {
				return err
			}

			privateKey, err := ethcrypto.HexToECDSA(ethPrivKeyHexStr)
			if err != nil {
				return fmt.Errorf("failed to decode private key: %w", err)
			}

			publicKey := privateKey.Public()
			publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			if !ok {
				return fmt.Errorf("invalid public key; expected: %T, got: %T", &ecdsa.PublicKey{}, publicKey)
			}

			goCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			fromAddress := ethcrypto.PubkeyToAddress(*publicKeyECDSA)

			nonce, err := ethClient.PendingNonceAt(goCtx, fromAddress)
			if err != nil {
				return err
			}

			goCtx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			ethChainID, err := ethClient.ChainID(goCtx)
			if err != nil {
				return fmt.Errorf("failed to get Ethereum chain ID: %w", err)
			}

			auth, err := bind.NewKeyedTransactorWithChainID(privateKey, ethChainID)
			if err != nil {
				return fmt.Errorf("failed to create Ethereum transactor: %w", err)
			}

			var gasPrice *big.Int

			gasPriceInt, err := cmd.Flags().GetUint64(flagEthGasPrice)
			switch {
			case err != nil:
				return err

			case gasPriceInt > 0:
				gasPrice = new(big.Int).SetUint64(gasPriceInt)

			default:
				gasPrice, err = ethClient.SuggestGasPrice(context.Background())
				if err != nil {
					return fmt.Errorf("failed to get Ethereum gas estimate: %w", err)
				}
			}

			gasLimit, err := cmd.Flags().GetUint64(flagEthGasLimit)
			if err != nil {
				return err
			}

			auth.Nonce = new(big.Int).SetUint64(nonce)
			auth.Value = big.NewInt(0) // in wei
			auth.GasLimit = gasLimit   // in units
			auth.GasPrice = gasPrice

			address, tx, _, err := peggy.DeploySolidity(auth, ethClient)
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
	cmd.Flags().Uint64(flagEthGasPrice, 0, "The Ethereum gas price to include in the transactions")
	cmd.Flags().Uint64(flagEthGasLimit, 6000000, "The Ethereum gas limit to include in the transactions")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func deployPeggyContract() {

}

func initPeggyContract() {

}
