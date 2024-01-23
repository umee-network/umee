package client

import (
	"github.com/umee-network/umee/v6/x/uibc"
)

func (c Client) UIBCQueryClient() uibc.QueryClient {
	return uibc.NewQueryClient(c.GrpcConn)
}

func (c Client) QueryUIBCParams() (uibc.Params, error) {
	ctx, cancel := c.NewCtxWithTimeout()
	defer cancel()

	queryResponse, err := c.UIBCQueryClient().Params(ctx, &uibc.QueryParams{})
	if err != nil {
		return uibc.Params{}, err
	}
	return queryResponse.Params, nil
}
