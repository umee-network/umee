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
// Accounts are generated using the list of mnemonics. Each string must be a sequence of words,
// eg: `["w11 w12 w13", "w21 w22 w23"]`. Keyring names for created accounts will be: val1, val2....
func NewClient(
	chainID,
	tmrpcEndpoint,
	grpcEndpoint string,
	mnemonics []string,
	gasAdjustment float64,
	encCfg sdkparams.EncodingConfig,
) (Client, error) {
	c, err := sdkclient.NewClient(chainID, tmrpcEndpoint, grpcEndpoint, mnemonics, gasAdjustment, encCfg)
	if err != nil {
		return Client{}, err
	}
	return Client{
		Client: c,
	}, nil
}

func (c Client) NewQCtx() (context.Context, context.CancelFunc) {
	return c.Query.NewCtx()
}

func (c Client) NewQCtxWithCancel() (context.Context, context.CancelFunc) {
	return c.Query.NewCtxWithCancel()
}
