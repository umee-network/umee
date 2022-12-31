package gRPC

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type UmeeClient struct {
	clientContext  *client.Context
	keyringKeyring keyring.Keyring
	keyringRecord  *keyring.Record
}

func NewUmeeClient(
	chainID string,
	tmrpcEndpoint string,
	grpcEndpoint string,
	accountMnemonic string,
) (*UmeeClient, error) {
	var err error

	uc := UmeeClient{}

	uc.keyringRecord, uc.keyringKeyring, err = CreateAccountFromMnemonic("val1", accountMnemonic)
	if err != nil {
		return nil, err
	}

	return &uc, nil
}

func (uc UmeeClient) createClientContext() {
	clientCtx := client.Context{
		ChainID:           oc.ChainID,
		InterfaceRegistry: oc.Encoding.InterfaceRegistry,
		Output:            os.Stderr,
		BroadcastMode:     flags.BroadcastSync,
		TxConfig:          oc.Encoding.TxConfig,
		AccountRetriever:  authtypes.AccountRetriever{},
		Codec:             oc.Encoding.Codec,
		LegacyAmino:       oc.Encoding.Amino,
		Input:             os.Stdin,
		NodeURI:           oc.TMRPC,
		Client:            tmRPC,
		Keyring:           kr,
		FromAddress:       oc.OracleAddr,
		FromName:          keyInfo.Name,
		From:              keyInfo.Name,
		OutputFormat:      "json",
		UseLedger:         false,
		Simulate:          false,
		GenerateOnly:      false,
		Offline:           false,
		SkipConfirm:       true,
	}

	return clientCtx, nil
}
