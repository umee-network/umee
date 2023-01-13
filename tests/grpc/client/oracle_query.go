package client

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
)

type OracleQuery struct {
	OracleQueryClient oracletypes.QueryClient
}

func (qc *QueryClient) OracleQueryClient() oracletypes.QueryClient {
	return oracletypes.NewQueryClient(qc.grpcConn)
}

func (qc *QueryClient) QueryParams() (oracletypes.Params, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	queryResponse, err := qc.OracleQueryClient().Params(ctx, &oracletypes.QueryParams{})
	if err != nil {
		return oracletypes.Params{}, err
	}
	return queryResponse.Params, nil
}

func (qc *QueryClient) QueryExchangeRates() ([]sdk.DecCoin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	queryResponse, err := qc.OracleQueryClient().ExchangeRates(ctx, &oracletypes.QueryExchangeRates{})
	if err != nil {
		return nil, err
	}
	return queryResponse.ExchangeRates, nil
}

func (qc *QueryClient) QueryMedians() ([]sdk.DecCoin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	queryResponse, err := qc.OracleQueryClient().Medians(ctx, &oracletypes.QueryMedians{})
	if err != nil {
		return nil, err
	}
	return queryResponse.Medians, nil
}

func (qc *QueryClient) QueryMedianDeviations() ([]sdk.DecCoin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	queryResponse, err := qc.OracleQueryClient().MedianDeviations(ctx, &oracletypes.QueryMedianDeviations{})
	if err != nil {
		return nil, err
	}
	return queryResponse.MedianDeviations, nil
}
