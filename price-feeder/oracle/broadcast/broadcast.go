package broadcast

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
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
	CosmosChain struct {
		ChainID          string
		KeyringBackend   string
		KeyringDir       string
		TMRPC            string
		RPCTimeout       time.Duration
		OracleAddr       sdk.AccAddress
		OracleAddrString string
		ValidatorAddr    string
		Encoding         umeeparams.EncodingConfig
		GasPrices        string
		GasAdjustment    float64
	}

	passReader struct {
		pass string
		buf  *bytes.Buffer
	}
)

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

type Broadcast struct {
	CosmosChain       CosmosChain
	KeyringPassphrase string
	Ctx               client.Context
	Factory           tx.Factory
}

func (b Broadcast) Broadcast(msgs ...sdk.Msg) error {
	return tx.BroadcastTx(b.Ctx, b.Factory, msgs...)
}

func NewCosmosChain(ChainID string,
	KeyringBackend string,
	KeyringDir string,
	TMRPC string,
	RPCTimeout time.Duration,
	OracleAddrString string,
	ValidatorAddr string) (*CosmosChain, error) {

	var chain CosmosChain

	chain.ChainID = ChainID
	chain.KeyringBackend = KeyringBackend
	chain.TMRPC = TMRPC
	chain.RPCTimeout = RPCTimeout
	var err error
	chain.OracleAddr, err = sdk.AccAddressFromBech32(OracleAddrString)
	chain.OracleAddrString = OracleAddrString
	if err != nil {
		return nil, err
	}
	chain.ValidatorAddr = ValidatorAddr
	chain.Encoding = umeeapp.MakeEncodingConfig()

	return &chain, nil

}

func NewBroadcast(cc *CosmosChain, keyringPassphrase string) (*Broadcast, error) {
	var keyringInput io.Reader
	if len(keyringPassphrase) > 0 {
		keyringInput = newPassReader(keyringPassphrase)
	} else {
		keyringInput = os.Stdin
	}

	kr, err := keyring.New("oracle", cc.KeyringBackend, cc.KeyringDir, keyringInput)
	if err != nil {
		return nil, err
	}

	httpClient, err := tmjsonclient.DefaultHTTPClient(cc.TMRPC)
	if err != nil {
		return nil, err
	}

	httpClient.Timeout = cc.RPCTimeout

	tmRPC, err := rpchttp.NewWithClient(cc.TMRPC, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}

	keyInfo, keyerr := kr.KeyByAddress(cc.OracleAddr)

	if keyerr != nil {
		return nil, err
	}

	clientCtx := client.Context{
		ChainID:           cc.ChainID,
		JSONCodec:         cc.Encoding.Marshaler,
		InterfaceRegistry: cc.Encoding.InterfaceRegistry,
		Output:            os.Stderr,
		BroadcastMode:     flags.BroadcastSync,
		TxConfig:          cc.Encoding.TxConfig,
		AccountRetriever:  authtypes.AccountRetriever{},
		Codec:             cc.Encoding.Marshaler,
		LegacyAmino:       cc.Encoding.Amino,
		Input:             os.Stdin,
		NodeURI:           cc.TMRPC,
		Client:            tmRPC,
		Keyring:           kr,
		FromAddress:       cc.OracleAddr,
		FromName:          keyInfo.GetName(),
		From:              keyInfo.GetName(),
		OutputFormat:      "json",
		UseLedger:         false,
		Simulate:          false,
		GenerateOnly:      false,
		Offline:           false,
		SkipConfirm:       true,
	}

	txFactory := tx.Factory{}.
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithChainID(cc.ChainID).
		WithTxConfig(clientCtx.TxConfig).
		WithGasAdjustment(cc.GasAdjustment).
		WithGasPrices(cc.GasPrices).
		WithKeybase(clientCtx.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithSimulateAndExecute(true)

	b := Broadcast{
		Factory:           txFactory,
		Ctx:               clientCtx,
		CosmosChain:       *cc,
		KeyringPassphrase: keyringPassphrase,
	}

	return &b, nil
}
