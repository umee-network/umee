package client

import (
	"github.com/umee-network/umee/v4/x/uibc"
)

func (c Client) UIBCQueryClient() uibc.QueryClient {
	return uibc.NewQueryClient(c.Query.GrpcConn)
}

func (c Client) QueryUIBCParams() (uibc.Params, error) {
	ctx, cancel := c.NewQCtx()
	defer cancel()

	queryResponse, err := c.UIBCQueryClient().Params(ctx, &uibc.QueryParams{})
	if err != nil {
		return uibc.Params{}, err
	}
	return queryResponse.Params, nil
}
