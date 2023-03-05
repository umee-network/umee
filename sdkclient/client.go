package sdkclient

import (
	"context"
	"time"

	sdkparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/rs/zerolog"
	"github.com/umee-network/umee/v4/sdkclient/query"
	"github.com/umee-network/umee/v4/sdkclient/tx"
)

// OjoClient is a helper for initializing a keychain, a cosmos-sdk client context,
// and sending transactions/queries to a specific Umee node
type Client struct {
	Query *query.Client
	Tx    *tx.Client
}

func NewClient(
	chainID,
	tmrpcEndpoint,
	grpcEndpoint,
	accountName,
	accountMnemonic string,
	gasAdjustment float64,
	ecfg sdkparams.EncodingConfig,
) (uc Client, err error) {
	uc = Client{}
	uc.Query, err = query.NewClient(grpcEndpoint, 15*time.Second)
	if err != nil {
		return Client{}, err
	}
	uc.Tx, err = tx.NewClient(chainID, tmrpcEndpoint, accountName, accountMnemonic, gasAdjustment, ecfg)
	return uc, err
}

func (c Client) NewChainHeightListener(ctx context.Context, logger zerolog.Logger) (*ChainHeightListener, error) {
	return NewChainHeightListener(ctx, c.Tx.ClientContext.Client, logger)
}

func (c Client) QueryTimeout() time.Duration {
	return c.Query.QueryTimeout
}
