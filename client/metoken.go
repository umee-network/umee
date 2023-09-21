package client

import (
	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/x/metoken"
)

func (c Client) MetokenQClient() metoken.QueryClient {
	return metoken.NewQueryClient(c.Query.GrpcConn)
}

func (c Client) QueryMetokenParams() (metoken.Params, error) {
	ctx, cancel := c.NewQCtx()
	defer cancel()

	resp, err := c.MetokenQClient().Params(ctx, &metoken.QueryParams{})
	if err != nil {
		return metoken.Params{}, err
	}
	return resp.Params, err
}

func (c Client) QueryMetokenIndexBalances(denom string) (*metoken.QueryIndexBalancesResponse, error) {
	ctx, cancel := c.NewQCtx()
	defer cancel()

	msg := &metoken.QueryIndexBalances{MetokenDenom: denom}
	if err := sdkutil.ValidateProtoMsg(msg); err != nil {
		return nil, err
	}
	return c.MetokenQClient().IndexBalances(ctx, msg)
}

func (c Client) QueryMetokenIndexes(denom string) (*metoken.QueryIndexesResponse, error) {
	ctx, cancel := c.NewQCtx()
	defer cancel()

	msg := &metoken.QueryIndexes{MetokenDenom: denom}
	if err := sdkutil.ValidateProtoMsg(msg); err != nil {
		return nil, err
	}
	return c.MetokenQClient().Indexes(ctx, msg)
}

func (c Client) QueryMetokenIndexPrices(denom string) (*metoken.QueryIndexPricesResponse, error) {
	ctx, cancel := c.NewQCtx()
	defer cancel()

	msg := &metoken.QueryIndexPrices{MetokenDenom: denom}
	if err := sdkutil.ValidateProtoMsg(msg); err != nil {
		return nil, err
	}
	return c.MetokenQClient().IndexPrices(ctx, msg)
}

//
// Tx
//
