package tx

import (
	"fmt"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdkparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	tmjsonclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

type Client struct {
	ChainID       string
	TMRPCEndpoint string

	ClientContext *client.Context
	gasAdjustment float64

	keyringKeyring keyring.Keyring
	keyringRecord  []*keyring.Record
	txFactory      *tx.Factory
	encCfg         sdkparams.EncodingConfig
}

// Initializes a cosmos sdk client context and transaction factory for
// signing and broadcasting transactions
func NewClient(
	chainID string,
	tmrpcEndpoint string,
	mnemonics []string,
	gasAdjustment float64,
	encCfg sdkparams.EncodingConfig,
) (c *Client, err error) {
	c = &Client{
		ChainID:       chainID,
		TMRPCEndpoint: tmrpcEndpoint,
		gasAdjustment: gasAdjustment,
		encCfg:        encCfg,
	}

	c.keyringKeyring, err = keyring.New(keyringAppName, keyring.BackendTest, "", nil, encCfg.Codec)
	if err != nil {
		return nil, err
	}

	for index, menomic := range mnemonics {
		kr, err := CreateAccountFromMnemonic(c.keyringKeyring, fmt.Sprintf("val%d", index), menomic, encCfg.Codec)
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
		WithSimulateAndExecute(true)
	c.txFactory = &f
}

func (c *Client) BroadcastTx(msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	c.ClientContext.From = c.keyringRecord[0].Name
	c.ClientContext.FromName = c.keyringRecord[0].Name
	c.ClientContext.FromAddress, _ = c.keyringRecord[0].GetAddress()
	return BroadcastTx(*c.ClientContext, *c.txFactory, msgs...)
}

func (c *Client) BroadcastTxVotes(proposalID uint64) error {
	for index := range c.keyringRecord {
		voter, err := c.keyringRecord[index].GetAddress()
		if err != nil {
			return err
		}

		voteType, err := govtypes.VoteOptionFromString("VOTE_OPTION_YES")
		if err != nil {
			return err
		}

		msg := govtypes.NewMsgVote(
			voter,
			proposalID,
			voteType,
		)

		c.ClientContext.From = c.keyringRecord[index].Name
		c.ClientContext.FromName = c.keyringRecord[index].Name
		c.ClientContext.FromAddress, _ = c.keyringRecord[index].GetAddress()

		for retry := 0; retry < 5; retry++ {
			// retry if txs fails, because sometimes account sequence mismatch occurs due to txs pending
			if _, err = BroadcastTx(*c.ClientContext, *c.txFactory, []sdk.Msg{msg}...); err == nil {
				break
			}
			time.Sleep(time.Second * 1)
		}

		if err != nil {
			return err
		}
	}

	return nil
}