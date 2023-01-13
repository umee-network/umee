package client

// UmeeClient is a helper for initializing a keychain, a cosmos-sdk client context,
// and sending transactions/queries to a specific Umee node
// It also starts up a websocket connection to track the current block height and
// uses the block height to ensure transactions happen within a certain window.
type UmeeClient struct {
	QueryClient *QueryClient
	TxClient    *TxClient
}

func NewUmeeClient(
	chainID string,
	tmrpcEndpoint string,
	grpcEndpoint string,
	accountName string,
	accountMnemonic string,
) (uc *UmeeClient, err error) {
	uc = &UmeeClient{}
	uc.QueryClient = NewQueryClient(grpcEndpoint)
	uc.TxClient, err = NewTxClient(chainID, tmrpcEndpoint, accountName, accountMnemonic)
	return uc, err
}
