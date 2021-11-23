package broadcast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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
	cosmosClient "github.com/umee-network/umee/price-feeder/oracle/broadcast/client"
)

type (
	CosmosChain struct {
		ChainID             string
		KeyringBackend      string
		KeyringDir          string
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

// Pre-vote and vote are separated out mainly for readability
// Prevote doesn't need the timeout functionality that vote needs,
// Because of the block timing validation on the node side

// Ref : https://github.com/terra-money/oracle-feeder/blob/baef2a4a02f57a2ffeaa207932b2e03d7fb0fb25/feeder/src/vote.ts#L230

func (b Broadcast) BroadcastPrevote(msgs ...sdk.Msg) (*sdk.TxResponse, error) {

	daemonClient, err := cosmosClient.NewCosmosClient(b.Ctx, b.CosmosChain.GRPCEndpoint, b.Factory)
	if err != nil {
		return nil, err
	}

	res, err := daemonClient.SyncBroadcastMsg(msgs[0])

	return res, err
}

// Need to try this continually, since err is either timing issue or whitelist error
// Ref : https://github.com/terra-money/core/blob/746a15f1bd83d62cd284e4af9471dc58701b3e33/x/oracle/keeper/msg_server.go#L89

func (b Broadcast) BroadcastVote(nextBlockHeight int, timeoutHeight int, msgs ...sdk.Msg) (*sdk.TxResponse, error) {

	daemonClient, err := cosmosClient.NewCosmosClient(b.Ctx, b.CosmosChain.GRPCEndpoint, b.Factory)
	maxBlockHeight := nextBlockHeight + timeoutHeight
	lastCheckHeight := nextBlockHeight - 1
	height := 0
	var res *sdk.TxResponse

	for height == 0 && lastCheckHeight < maxBlockHeight {

		time.Sleep(1500)

		latestBlockHeight, _ := b.GetHeight()

		if latestBlockHeight <= lastCheckHeight {
			continue
		}

		// set last check height to latest block height
		lastCheckHeight = latestBlockHeight

		// wait for indexing (not sure; but just for safety)
		time.Sleep(500)

		res, err = daemonClient.SyncBroadcastMsg(msgs[0])

		if err != nil {
			continue
		} else {
			height = int(res.Height)
		}
	}

	return res, err

}

func (b Broadcast) GetHeight() (int, error) {

	type header struct {
		Height string `json:"height"`
	}

	type block struct {
		Header header `json:"header"`
	}

	type result struct {
		Block block `json:"block"`
	}

	type getBlockResponse struct {
		Result result `json:"result"`
	}

	resp, err := http.Get(fmt.Sprintf("%s/block", b.CosmosChain.TMRPC))
	if err != nil {
		return 0, err
	}

	//Create a variable of the same type as our model
	var getBlockResp getBlockResponse
	if err := json.NewDecoder(resp.Body).Decode(&getBlockResp); err != nil {
		panic(err)
	}

	// Try to convert string to int
	height, err := strconv.Atoi(getBlockResp.Result.Block.Header.Height)

	if err != nil {
		panic(err)
	}

	return height, nil

}

func NewCosmosChain(ChainID string,
	KeyringBackend string,
	KeyringDir string,
	TMRPC string,
	RPCTimeout time.Duration,
	OracleAddrString string,
	ValidatorAddrString string,
	GRPCEndpoint string) (*CosmosChain, error) {

	var chain CosmosChain

	chain.ChainID = ChainID
	chain.KeyringBackend = KeyringBackend
	chain.KeyringDir = KeyringDir
	chain.TMRPC = TMRPC
	chain.RPCTimeout = RPCTimeout
	var err error
	chain.OracleAddr, err = sdk.AccAddressFromBech32(OracleAddrString)
	chain.OracleAddrString = OracleAddrString
	if err != nil {
		return nil, err
	}
	chain.ValidatorAddr = sdk.ValAddress(ValidatorAddrString)
	chain.ValidatorAddrString = ValidatorAddrString
	chain.Encoding = umeeapp.MakeEncodingConfig()
	// Static gas adjustment
	chain.GasAdjustment = 1.15
	chain.GRPCEndpoint = GRPCEndpoint
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

	keyInfo, err := kr.KeyByAddress(cc.OracleAddr)

	if err != nil {
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
