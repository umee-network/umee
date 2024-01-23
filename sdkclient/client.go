package sdkclient

import (
	"context"
	"log"
	"net"
	"os"
	"strings"
	"time"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
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
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client provides basic capabilities to connect to a Cosmos SDK based chain and execute
// transactions and queries. The object should be extended by another struct to provide
// chain specific transactions and queries. Example:
// https://github.com/umee-network/umee/blob/main/client
// Accounts are generated using the list of mnemonics. Each string must be a sequence of words,
// eg: `["w11 w12 w13", "w21 w22 w23"]`. Keyring names for created accounts will be: val1, val2....
type Client struct {
	ChainID       string
	TMRPCEndpoint string

	GrpcConn     *grpc.ClientConn
	grpcEndpoint string
	QueryTimeout time.Duration

	ClientContext *client.Context
	gasAdjustment float64

	keyringKeyring keyring.Keyring
	keyringRecord  []*keyring.Record
	txFactory      *tx.Factory
	encCfg         testutil.TestEncodingConfig

	logger *log.Logger
}

func NewClient(
	chainDataDir,
	chainID,
	tmrpcEndpoint,
	grpcEndpoint string,
	mnemonics map[string]string,
	gasAdjustment float64,
	encCfg testutil.TestEncodingConfig,
) (uc Client, err error) {
	c := Client{
		grpcEndpoint:  grpcEndpoint,
		QueryTimeout:  10 * time.Second, // TODO: customize
		ChainID:       chainID,
		TMRPCEndpoint: tmrpcEndpoint,
		gasAdjustment: gasAdjustment,
		encCfg:        encCfg,
		logger:        log.New(os.Stderr, "chain-client", log.LstdFlags),
	}
	c.keyringKeyring, err = keyring.New(keyringAppName, keyring.BackendTest, chainDataDir, nil, encCfg.Codec)
	if err != nil {
		return c, err
	}

	for accKey, menomic := range mnemonics {
		kr, err := CreateAccountFromMnemonic(c.keyringKeyring, accKey, menomic)
		c.keyringRecord = append(c.keyringRecord, kr)
		if err != nil {
			return c, err
		}
	}

	c.GrpcConn, err = grpc.Dial(
		c.grpcEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerFunc),
	)
	if err != nil {
		return c, err
	}

	err = c.initClientCtx()
	if err != nil {
		return c, err
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

// Broadcasts transaction. On success, increments the client sequence number.
func (c *Client) BroadcastTx(idx int, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	var err error
	r := c.keyringRecord[idx]
	cctx := *c.ClientContext
	cctx.FromName = r.Name
	cctx.FromAddress, err = r.GetAddress()
	if err != nil {
		c.logger.Fatalln("can't get keyring record, idx=", idx, err)
	}
	f := c.txFactory.WithFromName(r.Name)
	resp, err := BroadcastTx(cctx, f, msgs...)
	if err == nil {
		c.SetAccSeq(0)
		// c.IncAccSeq()
	}
	return resp, err
}

func (c *Client) GetAccSeq() uint64 {
	return c.txFactory.Sequence()
}

func (c *Client) IncAccSeq() {
	c.SetAccSeq(c.txFactory.Sequence() + 1)
}

func (c *Client) SetAccSeq(seq uint64) {
	*c.txFactory = c.txFactory.WithSequence(seq)
}

func (c *Client) WithAsyncBlock() *Client {
	c.ClientContext.BroadcastMode = flags.BroadcastAsync
	return c
}

func (c *Client) SenderAddr() sdk.AccAddress {
	addr, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		log.Fatal("can't get first keyring record", err)
	}
	return addr
}

func (c Client) NewChainHeightListener(ctx context.Context, logger zerolog.Logger) (*ChainHeightListener, error) {
	return NewChainHeightListener(ctx, c.TmClient(), logger)
}

func (c Client) TmClient() rpcclient.Client {
	return c.ClientContext.Client.(*rpchttp.HTTP)
}

// NewCtx creates a noop context
func (c Client) NewCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.QueryTimeout)
}

func dialerFunc(_ context.Context, addr string) (net.Conn, error) {
	return Connect(addr)
}

// Connect dials the given address and returns a net.Conn.
// The protoAddr argument should be prefixed with the protocol,
// eg. "tcp://127.0.0.1:8080" or "unix:///tmp/test.sock".
func Connect(protoAddr string) (net.Conn, error) {
	proto, address := protocolAndAddress(protoAddr)
	conn, err := net.Dial(proto, address)
	return conn, err
}

// protocolAndAddress splits an address into the protocol and address components.
// For instance, "tcp://127.0.0.1:8080" will be split into "tcp" and "127.0.0.1:8080".
// If the address has no protocol prefix, the default is "tcp".
func protocolAndAddress(listenAddr string) (string, string) {
	parts := strings.SplitN(listenAddr, "://", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return "tcp", listenAddr
}
