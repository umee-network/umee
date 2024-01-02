package tx

import (
	"log"
	"os"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	tmjsonclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type Client struct {
	ChainID       string
	TMRPCEndpoint string

	ClientContext *client.Context
	gasAdjustment float64

	keyringKeyring keyring.Keyring
	keyringRecord  []*keyring.Record
	txFactory      *tx.Factory
	encCfg         testutil.TestEncodingConfig

	logger *log.Logger
}

// Initializes a cosmos sdk client context and transaction factory for
// signing and broadcasting transactions by passing chainDataDir and remaining func arguments
// Note: For signing the transactions accounts are created by names like this val0, val1....
func NewClient(
	logger *log.Logger,
	chainDataDir,
	chainID,
	tmrpcEndpoint string,
	mnemonics map[string]string,
	gasAdjustment float64,
	encCfg testutil.TestEncodingConfig,
) (c *Client, err error) {
	c = &Client{
		ChainID:       chainID,
		TMRPCEndpoint: tmrpcEndpoint,
		gasAdjustment: gasAdjustment,
		encCfg:        encCfg,
		logger:        logger,
	}

	c.keyringKeyring, err = keyring.New(keyringAppName, keyring.BackendTest, chainDataDir, nil, encCfg.Codec)
	if err != nil {
		return nil, err
	}

	for accKey, menomic := range mnemonics {
		kr, err := CreateAccountFromMnemonic(c.keyringKeyring, accKey, menomic)
		c.keyringRecord = append(c.keyringRecord, kr)
		if err != nil {
			return nil, err
		}
	}

	err = c.initClientCtx()
	if err != nil {
		return nil, err
	}
	c.initTxFactory()

	return c, err
}

func (c *Client) initClientCtx() error {
	fromAddress, _ := c.keyringRecord[0].GetAddress()

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
		BroadcastMode:     flags.BroadcastSync,
		TxConfig:          c.encCfg.TxConfig,
		AccountRetriever:  authtypes.AccountRetriever{},
		Codec:             c.encCfg.Codec,
		LegacyAmino:       c.encCfg.Amino,
		Input:             os.Stdin,
		NodeURI:           c.TMRPCEndpoint,
		Client:            tmRPCClient,
		Keyring:           c.keyringKeyring,
		FromAddress:       fromAddress,
		FromName:          c.keyringRecord[0].Name,
		From:              c.keyringRecord[0].Name,
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
		WithSimulateAndExecute(true).
		WithFees("20000000uumee").
		WithGas(0)
	c.txFactory = &f
}

func (c *Client) BroadcastTx(idx int, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	var err error
	r := c.keyringRecord[idx]
	c.ClientContext.From = r.Name
	c.ClientContext.FromName = r.Name
	c.ClientContext.FromAddress, err = r.GetAddress()
	f := c.txFactory.WithFromName(r.Name)
	if err != nil {
		c.logger.Fatalln("can't get keyring record, idx=", idx, err)
	}
	return BroadcastTx(*c.ClientContext, f, msgs...)
}

func (c *Client) WithAccSeq(seq uint64) *Client {
	c.txFactory.WithSequence(seq)
	return c
}

func (c *Client) WithAsyncBlock() *Client {
	c.ClientContext.BroadcastMode = flags.BroadcastAsync
	return c
}

func (c *Client) SenderAddr() sdk.AccAddress {
	addr, _ := c.keyringRecord[0].GetAddress()
	return addr
}
