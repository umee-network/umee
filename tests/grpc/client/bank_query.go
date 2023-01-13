package client

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (qc *QueryClient) BankQueryClient() banktypes.QueryClient {
	return banktypes.NewQueryClient(qc.grpcConn)
}

func (qc *QueryClient) QueryAllBalances(address string) (sdk.Coins, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	queryResponse, err := qc.BankQueryClient().AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: address})
	if err != nil {
		return nil, err
	}
	return queryResponse.Balances, nil
}

func (bq *QueryClient) PrintBalances(address string) error {
	fmt.Printf("Balances for address %s\n", address)

	balances, err := bq.QueryAllBalances(address)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", balances)
	return nil
}
