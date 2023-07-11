package sdkclient

import (
	"context"
	"time"

	rpcclient "github.com/cometbft/cometbft/rpc/client"

	"github.com/rs/zerolog"
	"github.com/umee-network/umee/v5/sdkclient/query"
	"github.com/umee-network/umee/v5/sdkclient/tx"

	"github.com/cosmos/cosmos-sdk/types/module/testutil"
)

// Client provides basic capabilities to connect to a Cosmos SDK based chain and execute
// transactions and queries. The object should be extended by another struct to provide
// chain specific transactions and queries. Example:
// https://github.com/umee-network/umee/blob/main/client
// Accounts are generated using the list of mnemonics. Each string must be a sequence of words,
// eg: `["w11 w12 w13", "w21 w22 w23"]`. Keyring names for created accounts will be: val1, val2....
type Client struct {
	Query *query.Client
	Tx    *tx.Client
}

func NewClient(
	chainDataDir,
	chainID,
	tmrpcEndpoint,
	grpcEndpoint string,
	mnemonics map[string]string,
	gasAdjustment float64,
	encCfg testutil.TestEncodingConfig,
) (uc Client, err error) {
	uc = Client{}
	uc.Query, err = query.NewClient(grpcEndpoint, 15*time.Second)
	if err != nil {
		return Client{}, err
	}
	uc.Tx, err = tx.NewClient(chainDataDir, chainID, tmrpcEndpoint, mnemonics, gasAdjustment, encCfg)
	return uc, err
}

func (c Client) NewChainHeightListener(ctx context.Context, logger zerolog.Logger) (*ChainHeightListener, error) {
	return NewChainHeightListener(ctx, c.TmClient(), logger)
}

func (c Client) QueryTimeout() time.Duration {
	return c.Query.QueryTimeout
}

func (c Client) TmClient() rpcclient.Client {
	return c.Tx.ClientContext.Client.(rpcclient.Client)
}
