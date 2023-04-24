package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
)

func (c Client) OracleQueryClient() oracletypes.QueryClient {
	return oracletypes.NewQueryClient(c.Query.GrpcConn)
}

func (c Client) QueryOracleParams() (oracletypes.Params, error) {
	ctx, cancel := c.NewQCtx()
	defer cancel()

	queryResponse, err := c.OracleQueryClient().Params(ctx, &oracletypes.QueryParams{})
	return queryResponse.Params, err
}

func (c Client) QueryExchangeRates() ([]sdk.DecCoin, error) {
	ctx, cancel := c.NewQCtx()
	defer cancel()

	queryResponse, err := c.OracleQueryClient().ExchangeRates(ctx, &oracletypes.QueryExchangeRates{})
	return queryResponse.ExchangeRates, err
}

func (c Client) QueryMedians() ([]oracletypes.Price, error) {
	ctx, cancel := c.NewQCtx()
	defer cancel()

	resp, err := c.OracleQueryClient().Medians(ctx, &oracletypes.QueryMedians{})
	return resp.Medians, err
}

func (c Client) QueryMedianDeviations() ([]oracletypes.Price, error) {
	ctx, cancel := c.NewQCtx()
	defer cancel()

	queryResponse, err := c.OracleQueryClient().MedianDeviations(ctx, &oracletypes.QueryMedianDeviations{})
	return queryResponse.MedianDeviations, err
}
