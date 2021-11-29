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
	//OracleClient defines a structure that interfaces
	//with the umee node
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

func NewOracleClient(ChainID string,
	KeyringBackend string,
	KeyringDir string,
	KeyringPass string,
	TMRPC string,
	RPCTimeout time.Duration,
	OracleAddrString string,
	ValidatorAddrString string,
	GRPCEndpoint string,
	GasAdjustment float64) (*OracleClient, error) {

	oracleAddr, err := sdk.AccAddressFromBech32(OracleAddrString)
	if err != nil {
		return nil, err
	}

	validatorAddr := sdk.ValAddress(ValidatorAddrString)
	if err != nil {
		return nil, err
	}

	return &OracleClient{
		ChainID:             ChainID,
		KeyringBackend:      KeyringBackend,
		KeyringDir:          KeyringDir,
		KeyringPass:         KeyringPass,
		TMRPC:               TMRPC,
		RPCTimeout:          RPCTimeout,
		OracleAddr:          oracleAddr,
		OracleAddrString:    OracleAddrString,
		ValidatorAddr:       validatorAddr,
		ValidatorAddrString: ValidatorAddrString,
		Encoding:            umeeapp.MakeEncodingConfig(),
		GasAdjustment:       GasAdjustment,
		GRPCEndpoint:        GRPCEndpoint,
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

// Pre-vote and vote are separated out mainly for readability
// Prevote doesn't need the timeout functionality that vote needs,
// Because of the block timing validation on the node side

// Ref : https://github.com/terra-money/oracle-feeder/blob/baef2a4a02f57a2ffeaa207932b2e03d7fb0fb25/feeder/src/vote.ts#L230
func (oc OracleClient) BroadcastPrevote(msgs ...sdk.Msg) error {
	ctx, err := oc.CreateContext()
	if err != nil {
		return err
	}

	factory, err := oc.CreateTxFactory()
	if err != nil {
		return err
	}

	return tx.BroadcastTx(*ctx, *factory, msgs...)
}

// Broadcast vote - tries to vote within the next voting period
// Ref : https://github.com/terra-money/oracle-feeder/blob/baef2a4a02f57a2ffeaa207932b2e03d7fb0fb25/feeder/src/vote.ts#L230
func (oc OracleClient) BroadcastVote(nextBlockHeight int64, timeoutHeight int64, msgs ...sdk.Msg) error {
	maxBlockHeight := nextBlockHeight + timeoutHeight
	lastCheckHeight := nextBlockHeight - 1
	height := int64(0)

	// Create Context, factory
	ctx, err := oc.CreateContext()
	if err != nil {
		return err
	}

	factory, err := oc.CreateTxFactory()
	if err != nil {
		return err
	}

	// Re-try voting until timeout
	for height == 0 && lastCheckHeight < maxBlockHeight {

		latestBlockHeight, _ := rpcClient.GetChainHeight(*ctx)

		if latestBlockHeight <= lastCheckHeight {
			continue
		}

		// set last check height to latest block height
		lastCheckHeight = latestBlockHeight

		err := tx.BroadcastTx(*ctx, *factory, msgs...)
		if err != nil {
			return err
		}

		height, err = rpcClient.GetChainHeight(*ctx)
		if err != nil {
			return err
		}

	}

	return nil

}

func (oc *OracleClient) CreateContext() (*client.Context, error) {
	var keyringInput io.Reader
	if len(oc.KeyringPass) > 0 {
		keyringInput = newPassReader(oc.KeyringPass)
	} else {
		keyringInput = os.Stdin
	}

	kr, err := keyring.New("oracle", oc.KeyringBackend, oc.KeyringDir, keyringInput)
	if err != nil {
		return nil, err
	}

	httpClient, err := tmjsonclient.DefaultHTTPClient(oc.TMRPC)
	if err != nil {
		return nil, err
	}

	httpClient.Timeout = oc.RPCTimeout

	tmRPC, err := rpchttp.NewWithClient(oc.TMRPC, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}

	keyInfo, err := kr.KeyByAddress(oc.OracleAddr)
	if err != nil {
		return nil, err
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

	return &clientCtx, nil

}

func (oc *OracleClient) CreateTxFactory() (*tx.Factory, error) {
	clientCtx, err := oc.CreateContext()
	if err != nil {
		return nil, err
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

	return &txFactory, nil
}
