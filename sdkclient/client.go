package sdkclient

import (
	"context"

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
) (uc *Client, err error) {
	uc = &Client{}
	uc.Query, err = query.NewClient(grpcEndpoint)
	if err != nil {
		return nil, err
	}
	uc.Tx, err = tx.NewClient(chainID, tmrpcEndpoint, accountName, accountMnemonic, ecfg)
	return uc, err
}

func (oc *Client) NewChainHeightListener(ctx context.Context, logger zerolog.Logger) (*ChainHeightListener, error) {
	return NewChainHeightListener(ctx, oc.Tx.ClientContext.Client, logger)
}
