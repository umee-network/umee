package sdkclient

import (
	"context"
	"time"

	sdkparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/rs/zerolog"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"github.com/umee-network/umee/v4/sdkclient/query"
	"github.com/umee-network/umee/v4/sdkclient/tx"
)

// Client provides basic capabilities to connect to a Cosmos SDK based chain and execute
// transactions and queries. The object should be extended by another struct to provide
// chain specific transactions and queries. Example:
// https://github.com/umee-network/umee/blob/main/client
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
	encCfg sdkparams.EncodingConfig,
) (uc Client, err error) {
	uc = Client{}
	uc.Query, err = query.NewClient(grpcEndpoint, 15*time.Second)
	if err != nil {
		return Client{}, err
	}
	uc.Tx, err = tx.NewClient(chainID, tmrpcEndpoint, accountName, accountMnemonic, gasAdjustment, encCfg)
	return uc, err
}

func (c Client) NewChainHeightListener(ctx context.Context, logger zerolog.Logger) (*ChainHeightListener, error) {
	return NewChainHeightListener(ctx, c.Tx.ClientContext.Client, logger)
}

func (c Client) QueryTimeout() time.Duration {
	return c.Query.QueryTimeout
}

func (c Client) TmClient() rpcclient.Client {
	return c.Tx.ClientContext.Client
}
