package client

import (
	"context"

	"github.com/umee-network/umee/v6/sdkclient"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
)

// Client sdkclient.Client and provides umee chain specific transactions and queries.
type Client struct {
	sdkclient.Client
	codec codec.Codec
}

// NewClient constructs Client object.
// Accounts are generated using the list of mnemonics. Each string must be a sequence of words,
// eg: `["w11 w12 w13", "w21 w22 w23"]`. Keyring names for created accounts will be: val1, val2....
func NewClient(
	chainDataDir,
	chainID,
	tmrpcEndpoint,
	grpcEndpoint string,
	mnemonics map[string]string,
	gasAdjustment float64,
	encCfg testutil.TestEncodingConfig,
) (Client, error) {
	c, err := sdkclient.NewClient(chainDataDir, chainID, tmrpcEndpoint, grpcEndpoint, mnemonics, gasAdjustment, encCfg)
	if err != nil {
		return Client{}, err
	}
	return Client{
		Client: c,
		codec:  encCfg.Codec,
	}, nil
}

func (c Client) NewQCtx() (context.Context, context.CancelFunc) {
	return c.Query.NewCtx()
}

func (c Client) NewQCtxWithCancel() (context.Context, context.CancelFunc) {
	return c.Query.NewCtxWithCancel()
}
