package tx

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdkparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	tmjsonclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

type Client struct {
	ChainID       string
	TMRPCEndpoint string

	ClientContext *client.Context
	gasAdjustment float64

	keyringKeyring keyring.Keyring
	keyringRecord  *keyring.Record
	txFactory      *tx.Factory
	encCfg         sdkparams.EncodingConfig
}

// Initializes a cosmos sdk client context and transaction factory for
// signing and broadcasting transactions
func NewClient(
	chainID string,
	tmrpcEndpoint string,
	accountName string,
	accountMnemonic string,
	gasAdjustment float64,
	encCfg sdkparams.EncodingConfig,
) (c *Client, err error) {
	c = &Client{
		ChainID:       chainID,
		TMRPCEndpoint: tmrpcEndpoint,
		gasAdjustment: gasAdjustment,
		encCfg:        encCfg,
	}

	c.keyringRecord, c.keyringKeyring, err = CreateAccountFromMnemonic(accountName, accountMnemonic, encCfg.Codec)
	if err != nil {
		return nil, err
	}

	err = c.initClientCtx()
	if err != nil {
		return nil, err
	}
	c.initTxFactory()

	return c, err
}

func (c *Client) initClientCtx() error {
	fromAddress, _ := c.keyringRecord.GetAddress()

	tmHTTPClient, err := tmjsonclient.DefaultHTTPClient(c.TMRPCEndpoint)
	if err != nil {
		return err
	}
	tmRPCClient, err := rpchttp.NewWithClient(c.TMRPCEndpoint, "/websocket", tmHTTPClient)
	if err != nil {
		return err
	}

	c.ClientContext = &client.Context{
		ChainID:           c.ChainID,
		InterfaceRegistry: c.encCfg.InterfaceRegistry,
		Output:            os.Stderr,
		BroadcastMode:     flags.BroadcastBlock,
		TxConfig:          c.encCfg.TxConfig,
		AccountRetriever:  authtypes.AccountRetriever{},
		Codec:             c.encCfg.Codec,
		LegacyAmino:       c.encCfg.Amino,
		Input:             os.Stdin,
		NodeURI:           c.TMRPCEndpoint,
		Client:            tmRPCClient,
		Keyring:           c.keyringKeyring,
		FromAddress:       fromAddress,
		FromName:          c.keyringRecord.Name,
		From:              c.keyringRecord.Name,
		OutputFormat:      "json",
		UseLedger:         false,
		Simulate:          false,
		GenerateOnly:      false,
		Offline:           false,
		SkipConfirm:       true,
	}
	return nil
}

func (c *Client) initTxFactory() {
	f := tx.Factory{}.
		WithAccountRetriever(c.ClientContext.AccountRetriever).
		WithChainID(c.ChainID).
		WithTxConfig(c.ClientContext.TxConfig).
		WithGasAdjustment(c.gasAdjustment).
		WithKeybase(c.ClientContext.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithSimulateAndExecute(true)
	c.txFactory = &f
}

func (c *Client) BroadcastTx(msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	return BroadcastTx(*c.ClientContext, *c.txFactory, msgs...)
}
