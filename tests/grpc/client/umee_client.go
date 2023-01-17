package client

import (
	"github.com/umee-network/umee/v4/tests/grpc/client/query"
	"github.com/umee-network/umee/v4/tests/grpc/client/tx"
)

// UmeeClient is a helper for initializing a keychain, a cosmos-sdk client context,
// and sending transactions/queries to a specific Umee node
type UmeeClient struct {
	QueryClient *query.Client
	TxClient    *tx.Client
}

func NewUmeeClient(
	chainID string,
	tmrpcEndpoint string,
	grpcEndpoint string,
	accountName string,
	accountMnemonic string,
) (uc *UmeeClient, err error) {
	uc = &UmeeClient{}
	uc.QueryClient, err = query.NewQueryClient(grpcEndpoint)
	if err != nil {
		return nil, err
	}
	uc.TxClient, err = tx.NewTxClient(chainID, tmrpcEndpoint, accountName, accountMnemonic)
	return uc, err
}
