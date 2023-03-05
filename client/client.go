package client

import (
	"context"

	sdkparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/umee-network/umee/v4/sdkclient"
)

type Client struct {
	sdkclient.Client
}

// Initializes a cosmos sdk client context and transaction factory for
// signing and broadcasting transactions
func NewClient(
	chainID,
	tmrpcEndpoint,
	grpcEndpoint,
	accountName,
	accountMnemonic string,
	gasAdjustment float64,
	ecfg sdkparams.EncodingConfig,
) (Client, error) {
	c, err := sdkclient.NewClient(chainID, tmrpcEndpoint, grpcEndpoint, accountName, accountMnemonic, gasAdjustment, ecfg)
	if err != nil {
		return Client{}, err
	}
	return Client{
		Client: c,
	}, nil
}

func (c *Client) NewQCtx() (context.Context, context.CancelFunc) {
	return c.Query.NewCtx()
}
