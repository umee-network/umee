package gRPC

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	umeeapp "github.com/umee-network/umee/v3/app"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sdk "github.com/cosmos/cosmos-sdk/types"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	tmjsonclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	oracletypes "github.com/umee-network/umee/v3/x/oracle/types"
)

const (
	GAS_ADJUSTMENT = 1
	QUERY_TIMEOUT  = 15 * time.Second
)

// UmeeClient is a helper for initializing a keychain, a cosmos-sdk client context,
// and sending transactions/queries to a specific Umee node
// It also starts up a websocket connection to track the current block height and
// uses the block height to ensure transactions happen within a certain window.
type UmeeClient struct {
	ChainID       string
	AccountName   string
	GRPCEndpoint  string
	TMRPCEndpoint string
	GasPrices     string

	keyringKeyring keyring.Keyring
	keyringRecord  *keyring.Record

	clientContext *client.Context
	txFactory     tx.Factory
	ChainHeight   *ChainHeight
	QueryClient   oracletypes.QueryClient
}

func NewUmeeClient(
	chainID string,
	tmrpcEndpoint string,
	grpcEndpoint string,
	accountName string,
	accountMnemonic string,
) (*UmeeClient, error) {
	var err error

	uc := UmeeClient{
		ChainID:       chainID,
		AccountName:   accountName,
		GRPCEndpoint:  grpcEndpoint,
		TMRPCEndpoint: tmrpcEndpoint,
	}

	uc.keyringRecord, uc.keyringKeyring, err = CreateAccountFromMnemonic(accountName, accountMnemonic)
	if err != nil {
		return nil, err
	}

	return &uc, nil
}

func (uc *UmeeClient) createClientContext() error {
	encoding := umeeapp.MakeEncodingConfig()
	fromAddress, _ := uc.keyringRecord.GetAddress()

	tmHttpClient, err := tmjsonclient.DefaultHTTPClient(uc.TMRPCEndpoint)
	if err != nil {
		return err
	}

	tmRpcClient, err := rpchttp.NewWithClient(uc.TMRPCEndpoint, "/websocket", tmHttpClient)
	if err != nil {
		return err
	}

	uc.clientContext = &client.Context{
		ChainID:           uc.ChainID,
		InterfaceRegistry: encoding.InterfaceRegistry,
		Output:            os.Stderr,
		BroadcastMode:     flags.BroadcastSync,
		TxConfig:          encoding.TxConfig,
		AccountRetriever:  authtypes.AccountRetriever{},
		Codec:             encoding.Codec,
		LegacyAmino:       encoding.Amino,
		Input:             os.Stdin,
		NodeURI:           uc.TMRPCEndpoint,
		Client:            tmRpcClient,
		Keyring:           uc.keyringKeyring,
		FromAddress:       fromAddress,
		FromName:          uc.keyringRecord.Name,
		From:              uc.keyringRecord.Name,
		OutputFormat:      "json",
		UseLedger:         false,
		Simulate:          false,
		GenerateOnly:      false,
		Offline:           false,
		SkipConfirm:       true,
	}
	return nil
}

// CreateTxFactory creates an SDK Factory instance used for transaction
// generation, signing and broadcasting.
func (uc *UmeeClient) createTxFactory() {
	uc.txFactory = tx.Factory{}.
		WithAccountRetriever(uc.clientContext.AccountRetriever).
		WithChainID(uc.ChainID).
		WithTxConfig(uc.clientContext.TxConfig).
		WithGasAdjustment(GAS_ADJUSTMENT).
		WithGasPrices(uc.GasPrices).
		WithKeybase(uc.clientContext.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithSimulateAndExecute(true)
}

func (uc *UmeeClient) createQueryClient() error {
	grpcConn, err := grpc.Dial(
		uc.GRPCEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerFunc),
	)
	if err != nil {
		return fmt.Errorf("failed to dial Umee gRPC service: %w", err)
	}
	uc.QueryClient = oracletypes.NewQueryClient(grpcConn)
	return nil
}

func (uc *UmeeClient) QueryParams() (oracletypes.Params, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()

	queryResponse, err := uc.QueryClient.Params(ctx, &oracletypes.QueryParams{})
	if err != nil {
		return oracletypes.Params{}, err
	}
	return queryResponse.Params, nil
}

func (uc *UmeeClient) QueryExchangeRates() ([]sdk.DecCoin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()

	queryResponse, err := uc.QueryClient.ExchangeRates(ctx, &oracletypes.QueryExchangeRates{})
	if err != nil {
		return nil, err
	}
	return queryResponse.ExchangeRates, nil
}

func (uc *UmeeClient) QueryMedians() ([]sdk.DecCoin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()

	queryResponse, err := uc.QueryClient.Medians(ctx, &oracletypes.QueryMedians{})
	if err != nil {
		return nil, err
	}
	return queryResponse.Medians, nil
}
