package client

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	rpcClient "github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	tmjsonclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"

	umeeapp "github.com/umee-network/umee/app"
	umeeparams "github.com/umee-network/umee/app/params"
)

type (
	// OracleClient defines a structure that interfaces with the Umee node.
	OracleClient struct {
		ChainID             string
		KeyringBackend      string
		KeyringDir          string
		KeyringPass         string
		TMRPC               string
		RPCTimeout          time.Duration
		OracleAddr          sdk.AccAddress
		OracleAddrString    string
		ValidatorAddr       sdk.ValAddress
		ValidatorAddrString string
		Encoding            umeeparams.EncodingConfig
		GasPrices           string
		GasAdjustment       float64
		GRPCEndpoint        string
		KeyringPassphrase   string
	}

	passReader struct {
		pass string
		buf  *bytes.Buffer
	}
)

func NewOracleClient(
	chainID string,
	keyringBackend string,
	keyringDir string,
	keyringPass string,
	tmrpc string,
	rpcTimeout time.Duration,
	oracleAddrString string,
	validatorAddrString string,
	grpcEndpoint string,
	gasAdjustment float64,
) (*OracleClient, error) {

	oracleAddr, err := sdk.AccAddressFromBech32(oracleAddrString)
	if err != nil {
		return nil, err
	}

	validatorAddr := sdk.ValAddress(validatorAddrString)
	if err != nil {
		return nil, err
	}

	return &OracleClient{
		ChainID:             chainID,
		KeyringBackend:      keyringBackend,
		KeyringDir:          keyringDir,
		KeyringPass:         keyringPass,
		TMRPC:               tmrpc,
		RPCTimeout:          rpcTimeout,
		OracleAddr:          oracleAddr,
		OracleAddrString:    oracleAddrString,
		ValidatorAddr:       validatorAddr,
		ValidatorAddrString: validatorAddrString,
		Encoding:            umeeapp.MakeEncodingConfig(),
		GasAdjustment:       gasAdjustment,
		GRPCEndpoint:        grpcEndpoint,
	}, nil
}

func newPassReader(pass string) io.Reader {
	return &passReader{
		pass: pass,
		buf:  new(bytes.Buffer),
	}
}

func (r *passReader) Read(p []byte) (n int, err error) {
	n, err = r.buf.Read(p)
	if err == io.EOF || n == 0 {
		r.buf.WriteString(r.pass + "\n")

		n, err = r.buf.Read(p)
	}

	return n, err
}

// BroadcastPrevote attempts to broadcast a pre-vote transaction. Note, submiting
// a pre-vote transaction does not require timeout functionality that vote
// transactions require due to the fact that block timing validation exists on
// the Umee node.
//
// Ref: https://github.com/terra-money/oracle-feeder/blob/baef2a4a02f57a2ffeaa207932b2e03d7fb0fb25/feeder/src/vote.ts#L230
func (oc OracleClient) BroadcastPrevote(msgs ...sdk.Msg) error {
	ctx, err := oc.CreateContext()
	if err != nil {
		return err
	}

	factory, err := oc.CreateTxFactory()
	if err != nil {
		return err
	}

	return tx.BroadcastTx(ctx, factory, msgs...)
}

// BroadcastVote attempts to broadcast a vote transaction within the next voting
// period.
//
// Ref: https://github.com/terra-money/oracle-feeder/blob/baef2a4a02f57a2ffeaa207932b2e03d7fb0fb25/feeder/src/vote.ts#L230
func (oc OracleClient) BroadcastVote(nextBlockHeight int64, timeoutHeight int64, msgs ...sdk.Msg) error {
	var height int64

	maxBlockHeight := nextBlockHeight + timeoutHeight
	lastCheckHeight := nextBlockHeight - 1

	ctx, err := oc.CreateContext()
	if err != nil {
		return err
	}

	factory, err := oc.CreateTxFactory()
	if err != nil {
		return err
	}

	// re-try voting until timeout
	for height == 0 && lastCheckHeight < maxBlockHeight {
		latestBlockHeight, err := rpcClient.GetChainHeight(ctx)
		if err != nil {
			return err
		}

		if latestBlockHeight <= lastCheckHeight {
			continue
		}

		// set last check height to latest block height
		lastCheckHeight = latestBlockHeight

		err = tx.BroadcastTx(ctx, factory, msgs...)
		if err != nil {
			return err
		}

		height, err = rpcClient.GetChainHeight(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateContext creates an SDK client Context instance used for transaction
// generation, signing and broadcasting.
func (oc OracleClient) CreateContext() (client.Context, error) {
	var keyringInput io.Reader
	if len(oc.KeyringPass) > 0 {
		keyringInput = newPassReader(oc.KeyringPass)
	} else {
		keyringInput = os.Stdin
	}

	kr, err := keyring.New("oracle", oc.KeyringBackend, oc.KeyringDir, keyringInput)
	if err != nil {
		return client.Context{}, err
	}

	httpClient, err := tmjsonclient.DefaultHTTPClient(oc.TMRPC)
	if err != nil {
		return client.Context{}, err
	}

	httpClient.Timeout = oc.RPCTimeout

	tmRPC, err := rpchttp.NewWithClient(oc.TMRPC, "/websocket", httpClient)
	if err != nil {
		return client.Context{}, err
	}

	keyInfo, err := kr.KeyByAddress(oc.OracleAddr)
	if err != nil {
		return client.Context{}, err
	}

	clientCtx := client.Context{
		ChainID:           oc.ChainID,
		JSONCodec:         oc.Encoding.Marshaler,
		InterfaceRegistry: oc.Encoding.InterfaceRegistry,
		Output:            os.Stderr,
		BroadcastMode:     flags.BroadcastSync,
		TxConfig:          oc.Encoding.TxConfig,
		AccountRetriever:  authtypes.AccountRetriever{},
		Codec:             oc.Encoding.Marshaler,
		LegacyAmino:       oc.Encoding.Amino,
		Input:             os.Stdin,
		NodeURI:           oc.TMRPC,
		Client:            tmRPC,
		Keyring:           kr,
		FromAddress:       oc.OracleAddr,
		FromName:          keyInfo.GetName(),
		From:              keyInfo.GetName(),
		OutputFormat:      "json",
		UseLedger:         false,
		Simulate:          false,
		GenerateOnly:      false,
		Offline:           false,
		SkipConfirm:       true,
	}

	return clientCtx, nil
}

// CreateTxFactory creates an SDK Factory instance used for transaction
// generation, signing and broadcasting.
func (oc OracleClient) CreateTxFactory() (tx.Factory, error) {
	clientCtx, err := oc.CreateContext()
	if err != nil {
		return tx.Factory{}, err
	}

	txFactory := tx.Factory{}.
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithChainID(oc.ChainID).
		WithTxConfig(clientCtx.TxConfig).
		WithGasAdjustment(oc.GasAdjustment).
		WithGasPrices(oc.GasPrices).
		WithKeybase(clientCtx.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithSimulateAndExecute(true)

	return txFactory, nil
}
