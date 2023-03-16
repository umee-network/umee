package client

import (
	"context"

	sdkparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/umee-network/umee/v4/sdkclient"
)

// Client sdkclient.Client and provides umee chain specific transactions and queries.
type Client struct {
	sdkclient.Client
}

// NewClient constructs Client object.
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
