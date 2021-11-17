package cli

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"

	"github.com/umee-network/umee/x/peggy/types"
)

const (
	flagOrchEthSig = "orch-eth-sig"
	flagEthPrivKey = "eth-priv-key"
)

func GetTxCmd(storeKey string) *cobra.Command {
	peggyTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Peggy transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	peggyTxCmd.AddCommand([]*cobra.Command{
		CmdSendToEth(),
		CmdRequestBatch(),
		CmdSetOrchestratorAddress(),
		GetUnsafeTestingCmd(),
	}...)

	return peggyTxCmd
}

func GetUnsafeTestingCmd() *cobra.Command {
	testingTxCmd := &cobra.Command{
		Use:                        "unsafe_testing",
		Short:                      "helpers for testing. not going into production",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	testingTxCmd.AddCommand([]*cobra.Command{
		CmdUnsafeETHPrivKey(),
		CmdUnsafeETHAddr(),
	}...)

	return testingTxCmd
}

func CmdUnsafeETHPrivKey() *cobra.Command {
	return &cobra.Command{
		Use:   "gen-eth-key",
		Short: "Generate and print a new ecdsa key",
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := ethcrypto.GenerateKey()
			if err != nil {
				return sdkerrors.Wrap(err, "can not generate key")
			}
			k := "0x" + hex.EncodeToString(ethcrypto.FromECDSA(key))
			println(k)
			return nil
		},
	}
}

func CmdUnsafeETHAddr() *cobra.Command {
	return &cobra.Command{
		Use:   "eth-address",
		Short: "Print address for an ECDSA eth key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			privKeyString := args[0][2:]
			privateKey, err := ethcrypto.HexToECDSA(privKeyString)
			if err != nil {
				log.Fatal(err)
			}
			// You've got to do all this to get an Eth address from the private key
			publicKey := privateKey.Public()
			publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			if !ok {
				log.Fatal("error casting public key to ECDSA")
			}
			ethAddress := ethcrypto.PubkeyToAddress(*publicKeyECDSA).Hex()
			println(ethAddress)
			return nil
		},
	}
}

func CmdSendToEth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-to-eth [eth-dest] [amount] [bridge-fee]",
		Short: "Adds a new entry to the transaction pool to withdraw an amount from the Ethereum bridge contract",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cosmosAddr := cliCtx.GetFromAddress()

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "amount")
			}
			bridgeFee, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return sdkerrors.Wrap(err, "bridge fee")
			}

			if len(amount) > 1 || len(bridgeFee) > 1 {
				return fmt.Errorf("coin amounts too long, expecting just 1 coin amount for both amount and bridgeFee")
			}

			// Make the message
			msg := types.MsgSendToEth{
				Sender:    cosmosAddr.String(),
				EthDest:   args[0],
				Amount:    amount[0],
				BridgeFee: bridgeFee[0],
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			// Send it
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdRequestBatch() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build-batch [denom]",
		Short: "Build a new batch on the cosmos side for pooled withdrawal transactions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cosmosAddr := cliCtx.GetFromAddress()

			denom := args[0]

			msg := types.MsgRequestBatch{
				Orchestrator: cosmosAddr.String(),
				Denom:        denom,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			// Send it
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSetOrchestratorAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-orchestrator-address [validator-acc-address] [orchestrator-acc-address] [ethereum-address]",
		Short: "Allows validators to delegate their voting responsibilities to a given key.",
		Long: `Set a validator's Ethereum and orchestrator addresses. The delegate
key owner must sign over a binary Proto-encoded SetOrchestratorAddressesSignMsg
message. The message contains the delegated key owner's address and current
account nonce.

An operator may provide an already generated signature via the --orch-eth-sig flag
or have the Ethereum signature automatically generated by providing the Ethereum
private key via the --eth-priv-key flag. If generating the Ethereum signature
manually, the operator must use the current account nonce.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			valAccAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			orcAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			var ethSig []byte
			if s, err := cmd.Flags().GetString(flagOrchEthSig); len(s) > 0 && err == nil {
				ethSig, err = hexutil.Decode(s)
				if err != nil {
					return err
				}
			} else {
				ethPrivKeyStr, err := cmd.Flags().GetString(flagEthPrivKey)
				if err != nil {
					return err
				}

				privKeyBz, err := hexutil.Decode(ethPrivKeyStr)
				if err != nil {
					return fmt.Errorf("failed to parse Ethereum private key: %w", err)
				}

				ethPrivKey, err := ethcrypto.ToECDSA(privKeyBz)
				if err != nil {
					return fmt.Errorf("failed to convert private key: %w", err)
				}

				queryClient := authtypes.NewQueryClient(clientCtx)
				res, err := queryClient.Account(cmd.Context(), &authtypes.QueryAccountRequest{Address: valAccAddr.String()})
				if err != nil {
					return err
				}

				var acc authtypes.AccountI
				if err := clientCtx.Codec.UnpackAny(res.Account, &acc); err != nil {
					return fmt.Errorf("failed to unmarshal account: %w", err)
				}

				signMsgBz := clientCtx.Codec.MustMarshal(&types.SetOrchestratorAddressesSignMsg{
					ValidatorAddress: sdk.ValAddress(valAccAddr).String(),
					Nonce:            acc.GetSequence(),
				})

				ethSig, err = types.NewEthereumSignature(ethcrypto.Keccak256Hash(signMsgBz), ethPrivKey)
				if err != nil {
					return fmt.Errorf("failed to create Ethereum signature: %w", err)
				}
			}

			msg := types.NewMsgSetOrchestratorAddress(valAccAddr, orcAddr, ethcmn.HexToAddress(args[2]), ethSig)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagOrchEthSig, "", "The Ethereum signature used to set the Orchestrator addresses")
	cmd.Flags().String(flagEthPrivKey, "", "The Ethereum private key used to set the Orchestrator addresses")

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
