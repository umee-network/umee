package query

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
)

func (c *Client) OracleQueryClient() oracletypes.QueryClient {
	return oracletypes.NewQueryClient(c.grpcConn)
}

func (c *Client) QueryParams() (oracletypes.Params, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	queryResponse, err := c.OracleQueryClient().Params(ctx, &oracletypes.QueryParams{})
	if err != nil {
		return oracletypes.Params{}, err
	}
	return queryResponse.Params, nil
}

func (c *Client) QueryExchangeRates() ([]sdk.DecCoin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	queryResponse, err := c.OracleQueryClient().ExchangeRates(ctx, &oracletypes.QueryExchangeRates{})
	if err != nil {
		return nil, err
	}
	return queryResponse.ExchangeRates, nil
}

func (c *Client) QueryMedians() ([]sdk.DecCoin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	queryResponse, err := c.OracleQueryClient().Medians(ctx, &oracletypes.QueryMedians{})
	if err != nil {
		return nil, err
	}
	return queryResponse.Medians, nil
}

func (c *Client) QueryMedianDeviations() ([]sdk.DecCoin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	queryResponse, err := c.OracleQueryClient().MedianDeviations(ctx, &oracletypes.QueryMedianDeviations{})
	if err != nil {
		return nil, err
	}
	return queryResponse.MedianDeviations, nil
}
