package client

import (
	"github.com/umee-network/umee/v6/x/metoken"
)

func (c Client) MetokenQClient() metoken.QueryClient {
	return metoken.NewQueryClient(c.Query.GrpcConn)
}

func (c Client) QueryMetoken() (metoken.Params, error) {
	ctx, cancel := c.NewQCtx()
	defer cancel()

	resp, err := c.MetokenQClient().Params(ctx, &metoken.QueryParams{})
	return resp.Params, err
}

//
// Tx
//

// func (c Client) MetokenTxClient() metoken.QueryClient {
// 	return metoken.NewMsgClient(c.Tx.ClientContext)
// }
