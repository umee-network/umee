package client

import (
	"context"
	"encoding/hex"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Package from https://github.com/umee-network/peggo/blob/main/cmd/peggo/client/cosmos.go

func init() {
	// set the address prefixes
	// config := sdk.GetConfig()

	// This is specific to Injective chain
	// ctypes.SetBech32Prefixes(config)
	// ctypes.SetBip44CoinType(config)
}

type CosmosClient interface {
	CanSignTransactions() bool
	FromAddress() sdk.AccAddress
	QueryClient() *grpc.ClientConn
	SyncBroadcastMsg(msgs ...sdk.Msg) (*sdk.TxResponse, error)
	AsyncBroadcastMsg(msgs ...sdk.Msg) (*sdk.TxResponse, error)
	QueueBroadcastMsg(msgs ...sdk.Msg) error
	ClientContext() client.Context
	Close()
}

// ProtocolAndAddress splits an address into the protocol and address components.
// For instance, "tcp://127.0.0.1:8080" will be split into "tcp" and "127.0.0.1:8080".
// If the address has no protocol prefix, the default is "tcp".
func ProtocolAndAddress(listenAddr string) (string, string) {
	protocol, address := "tcp", listenAddr
	parts := strings.SplitN(address, "://", 2)
	if len(parts) == 2 {
		protocol, address = parts[0], parts[1]
	}
	return protocol, address
}

// Connect dials the given address and returns a net.Conn. The protoAddr argument should be prefixed with the protocol,
// eg. "tcp://127.0.0.1:8080" or "unix:///tmp/test.sock"
func Connect(protoAddr string) (net.Conn, error) {
	proto, address := ProtocolAndAddress(protoAddr)
	conn, err := net.Dial(proto, address)
	return conn, err
}

func dialerFunc(ctx context.Context, addr string) (net.Conn, error) {
	return Connect(addr)
}

// NewCosmosClient creates a new gRPC client that communicates with gRPC server at protoAddr.
// protoAddr must be in form "tcp://127.0.0.1:8080" or "unix:///tmp/test.sock", protocol is required.
func NewCosmosClient(
	ctx client.Context,
	protoAddr string,
	factory tx.Factory,
	options ...CosmosClientOption,
) (CosmosClient, error) {
	conn, err := grpc.Dial(protoAddr, grpc.WithInsecure(), grpc.WithContextDialer(dialerFunc))
	if err != nil {
		err := errors.Wrapf(err, "failed to connect to the gRPC: %s", protoAddr)
		return nil, err
	}

	opts := defaultCosmosClientOptions()
	for _, opt := range options {
		if err := opt(opts); err != nil {
			err = errors.Wrap(err, "error in a cosmos client option")
			return nil, err
		}
	}

	txFactory := factory
	if len(opts.GasPrices) > 0 {
		txFactory = txFactory.WithGasPrices(opts.GasPrices)
	}

	cc := &cosmosClient{
		ctx:  ctx,
		opts: opts,

		//logger: logger.With().Str("module", "cosmos_client").Logger(),

		conn:      conn,
		txFactory: txFactory,
		canSign:   ctx.Keyring != nil,
		syncMux:   new(sync.Mutex),
		msgC:      make(chan sdk.Msg, msgCommitBatchSizeLimit),
		doneC:     make(chan bool, 1),
	}

	if cc.canSign {
		var err error

		cc.accNum, cc.accSeq, err = cc.txFactory.AccountRetriever().GetAccountNumberSequence(ctx, ctx.GetFromAddress())
		if err != nil {
			err = errors.Wrap(err, "failed to get initial account num and seq")
			return nil, err
		}

		go cc.runBatchBroadcast()
	}

	return cc, nil
}

type cosmosClientOptions struct {
	GasPrices string
}

func defaultCosmosClientOptions() *cosmosClientOptions {
	return &cosmosClientOptions{}
}

type CosmosClientOption func(opts *cosmosClientOptions) error

func OptionGasPrices(gasPrices string) CosmosClientOption {
	return func(opts *cosmosClientOptions) error {
		_, err := sdk.ParseDecCoins(gasPrices)
		if err != nil {
			err = errors.Wrapf(err, "failed to ParseDecCoins %s", gasPrices)
			return err
		}

		opts.GasPrices = gasPrices
		return nil
	}
}

func (c *cosmosClient) syncNonce() {
	num, seq, err := c.txFactory.AccountRetriever().GetAccountNumberSequence(c.ctx, c.ctx.GetFromAddress())
	if err != nil {
		//c.logger.Err(err).Msg("failed to get account seq")
		return
	} else if num != c.accNum {
		/*c.logger.Panic().
		Uint64("account_num", num).
		Uint64("expected_account_num", c.accNum).
		Msg("account number changed during nonce sync")*/
	}

	c.accSeq = seq
}

type cosmosClient struct {
	ctx  client.Context
	opts *cosmosClientOptions
	//logger    zerolog.Logger
	conn      *grpc.ClientConn
	txFactory tx.Factory

	doneC   chan bool
	msgC    chan sdk.Msg
	syncMux *sync.Mutex

	accNum uint64
	accSeq uint64

	closed  int64
	canSign bool
}

func (c *cosmosClient) QueryClient() *grpc.ClientConn {
	return c.conn
}

func (c *cosmosClient) ClientContext() client.Context {
	return c.ctx
}

func (c *cosmosClient) CanSignTransactions() bool {
	return c.canSign
}

func (c *cosmosClient) FromAddress() sdk.AccAddress {
	if !c.canSign {
		return sdk.AccAddress{}
	}

	return c.ctx.FromAddress
}

var (
	ErrQueueClosed    = errors.New("queue is closed")
	ErrEnqueueTimeout = errors.New("enqueue timeout")
	ErrReadOnly       = errors.New("client is in read-only mode")
)

// SyncBroadcastMsg sends Tx to chain and waits until Tx is included in block.
func (c *cosmosClient) SyncBroadcastMsg(msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	c.syncMux.Lock()
	defer c.syncMux.Unlock()

	c.txFactory = c.txFactory.WithSequence(c.accSeq)
	c.txFactory = c.txFactory.WithAccountNumber(c.accNum)
	res, err := c.broadcastTx(c.ctx, c.txFactory, true, msgs...)
	if err != nil {
		if strings.Contains(err.Error(), "account sequence mismatch") {
			c.syncNonce()
			c.txFactory = c.txFactory.WithSequence(c.accSeq)
			c.txFactory = c.txFactory.WithAccountNumber(c.accNum)
			//c.logger.Debug().Uint64("nonce", c.accSeq).Msg("retrying broadcastTx with nonce")
			res, err = c.broadcastTx(c.ctx, c.txFactory, true, msgs...)
		}
		if err != nil {
			//resJSON, _ := json.MarshalIndent(res, "", "\t")
			//c.logger.Err(err).Int("size", len(msgs)).RawJSON("tx_response", resJSON).Msg("failed to (sync) broadcast tx")
			return nil, err
		}
	}

	c.accSeq++

	return res, nil
}

// AsyncBroadcastMsg sends Tx to chain and doesn't wait until Tx is included in block. This method
// cannot be used for rapid Tx sending, it is expected that you wait for transaction status with
// external tools. If you want sdk to wait for it, use SyncBroadcastMsg.
func (c *cosmosClient) AsyncBroadcastMsg(msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	c.syncMux.Lock()
	defer c.syncMux.Unlock()

	c.txFactory = c.txFactory.WithSequence(c.accSeq)
	c.txFactory = c.txFactory.WithAccountNumber(c.accNum)
	res, err := c.broadcastTx(c.ctx, c.txFactory, false, msgs...)
	if err != nil {
		if strings.Contains(err.Error(), "account sequence mismatch") {
			c.syncNonce()
			c.txFactory = c.txFactory.WithSequence(c.accSeq)
			c.txFactory = c.txFactory.WithAccountNumber(c.accNum)
			//c.logger.Debug().Uint64("nonce", c.accSeq).Msg("retrying broadcastTx with nonce")
			res, err = c.broadcastTx(c.ctx, c.txFactory, false, msgs...)
		}
		if err != nil {
			//resJSON, _ := json.MarshalIndent(res, "", "\t")
			//c.logger.Err(err).Int("size", len(msgs)).RawJSON("tx_response", resJSON).Msg("failed to (async) broadcast tx")
			return nil, err
		}
	}

	c.accSeq++

	return res, nil
}

const (
	defaultBroadcastStatusPoll = 100 * time.Millisecond
	defaultBroadcastTimeout    = 40 * time.Second
)

func (c *cosmosClient) broadcastTx(
	clientCtx client.Context,
	txf tx.Factory,
	await bool,
	msgs ...sdk.Msg,
) (*sdk.TxResponse, error) {

	txf, err := c.prepareFactory(clientCtx, txf)
	if err != nil {
		err = errors.Wrap(err, "failed to prepareFactory")
		return nil, err
	}

	if txf.SimulateAndExecute() || clientCtx.Simulate {
		_, adjusted, err := tx.CalculateGas(clientCtx, txf, msgs...)
		if err != nil {
			err = errors.Wrap(err, "failed to CalculateGas")
			return nil, err
		}

		txf = txf.WithGas(adjusted)
	}

	txn, err := tx.BuildUnsignedTx(txf, msgs...)
	if err != nil {
		err = errors.Wrap(err, "failed to BuildUnsignedTx")
		return nil, err
	}

	txn.SetFeeGranter(clientCtx.GetFeeGranterAddress())
	err = tx.Sign(txf, clientCtx.GetFromName(), txn, true)
	if err != nil {
		err = errors.Wrap(err, "failed to Sign Tx")
		return nil, err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txn.GetTx())
	if err != nil {
		err = errors.Wrap(err, "failed TxEncoder to encode Tx")
		return nil, err
	}

	res, err := clientCtx.BroadcastTxSync(txBytes)
	if !await || err != nil {
		return res, err
	}

	awaitCtx, cancelFn := context.WithTimeout(context.Background(), defaultBroadcastTimeout)
	defer cancelFn()

	txHash, _ := hex.DecodeString(res.TxHash)
	t := time.NewTimer(defaultBroadcastStatusPoll)

	for {
		select {
		case <-awaitCtx.Done():
			err := errors.Wrapf(ErrTimedOut, "%s", res.TxHash)
			t.Stop()
			return nil, err
		case <-t.C:
			resultTx, err := clientCtx.Client.Tx(awaitCtx, txHash, false)
			if err != nil {
				if errRes := client.CheckTendermintError(err, txBytes); errRes != nil {
					return errRes, err
				}

				// log.WithError(err).Warningln("Tx Error for Hash:", res.TxHash)

				t.Reset(defaultBroadcastStatusPoll)
				continue

			} else if resultTx.Height > 0 {
				res = sdk.NewResponseResultTx(resultTx, res.Tx, res.Timestamp)
				t.Stop()
				return res, err
			}

			t.Reset(defaultBroadcastStatusPoll)
		}
	}
}

var ErrTimedOut = errors.New("tx timed out")

// prepareFactory ensures the account defined by ctx.GetFromAddress() exists and
// if the account number and/or the account sequence number are zero (not set),
// they will be queried for and set on the provided Factory. A new Factory with
// the updated fields will be returned.
func (c *cosmosClient) prepareFactory(clientCtx client.Context, txf tx.Factory) (tx.Factory, error) {
	from := clientCtx.GetFromAddress()

	if err := txf.AccountRetriever().EnsureExists(clientCtx, from); err != nil {
		return txf, err
	}

	initNum, initSeq := txf.AccountNumber(), txf.Sequence()
	if initNum == 0 || initSeq == 0 {
		num, seq, err := txf.AccountRetriever().GetAccountNumberSequence(clientCtx, from)
		if err != nil {
			return txf, err
		}

		if initNum == 0 {
			txf = txf.WithAccountNumber(num)
		}

		if initSeq == 0 {
			txf = txf.WithSequence(seq)
		}
	}

	return txf, nil
}

// QueueBroadcastMsg enqueues a list of messages. Messages will added to the queue
// and grouped into Txns in chunks. Use this method to mass broadcast Txns with efficiency.
func (c *cosmosClient) QueueBroadcastMsg(msgs ...sdk.Msg) error {
	if !c.canSign {
		return ErrReadOnly
	} else if atomic.LoadInt64(&c.closed) == 1 {
		return ErrQueueClosed
	}

	t := time.NewTimer(10 * time.Second)
	for _, msg := range msgs {
		select {
		case <-t.C:
			return ErrEnqueueTimeout
		case c.msgC <- msg:
		}
	}
	t.Stop()

	return nil
}

func (c *cosmosClient) Close() {
	if !c.canSign {
		return
	}

	if atomic.CompareAndSwapInt64(&c.closed, 0, 1) {
		close(c.msgC)
	}

	<-c.doneC

	if c.conn != nil {
		c.conn.Close()
	}
}

const (
	msgCommitBatchSizeLimit = 1024
	msgCommitBatchTimeLimit = 500 * time.Millisecond
)

func (c *cosmosClient) runBatchBroadcast() {
	expirationTimer := time.NewTimer(msgCommitBatchTimeLimit)
	msgBatch := make([]sdk.Msg, 0, msgCommitBatchSizeLimit)

	submitBatch := func(toSubmit []sdk.Msg) {
		c.syncMux.Lock()
		defer c.syncMux.Unlock()

		c.txFactory = c.txFactory.WithSequence(c.accSeq)
		c.txFactory = c.txFactory.WithAccountNumber(c.accNum)
		//c.logger.Debug().Uint64("nonce", c.accSeq).Msg("broadcastTx with nonce")
		res, err := c.broadcastTx(c.ctx, c.txFactory, true, toSubmit...)
		if err != nil {
			if strings.Contains(err.Error(), "account sequence mismatch") {
				c.syncNonce()
				c.txFactory = c.txFactory.WithSequence(c.accSeq)
				c.txFactory = c.txFactory.WithAccountNumber(c.accNum)
				//c.logger.Debug().Uint64("nonce", c.accSeq).Msg("retrying broadcastTx with nonce")
				res, err = c.broadcastTx(c.ctx, c.txFactory, true, toSubmit...)
			}
			if err != nil {
				//resJSON, _ := json.MarshalIndent(res, "", "\t")
				/*c.logger.Err(err).
				Int("size", len(toSubmit)).
				RawJSON("tx_response", resJSON).
				Msg("failed to (sync) broadcast batch tx")*/
				return
			}
		}

		if res.Code != 0 {
			err = errors.Errorf("error %d (%s): %s", res.Code, res.Codespace, res.RawLog)
			//c.logger.Err(err).Str("tx_hash", res.TxHash).Msg("failed to (sync) broadcast batch tx")
		} else {
			//c.logger.Debug().Str("tx_hash", res.TxHash).Msg("batch tx committed successfully")
		}

		c.accSeq++
		//c.logger.Debug().Uint64("nonce", c.accSeq).Msg("nonce incremented")
	}

	for {
		select {
		case msg, ok := <-c.msgC:
			if !ok {
				// exit required
				if len(msgBatch) > 0 {
					submitBatch(msgBatch)
				}

				close(c.doneC)
				return
			}

			msgBatch = append(msgBatch, msg)

			if len(msgBatch) >= msgCommitBatchSizeLimit {
				toSubmit := msgBatch
				msgBatch = msgBatch[:0]
				expirationTimer.Reset(msgCommitBatchTimeLimit)

				submitBatch(toSubmit)
			}
		case <-expirationTimer.C:
			if len(msgBatch) > 0 {
				toSubmit := msgBatch
				msgBatch = msgBatch[:0]
				expirationTimer.Reset(msgCommitBatchTimeLimit)

				submitBatch(toSubmit)
			} else {
				expirationTimer.Reset(msgCommitBatchTimeLimit)
			}
		}
	}
}
