package query

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (c *Client) BankQueryClient() banktypes.QueryClient {
	return banktypes.NewQueryClient(c.grpcConn)
}

func (c *Client) QueryAllBalances(address string) (sdk.Coins, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	queryResponse, err := c.BankQueryClient().AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: address})
	if err != nil {
		return nil, err
	}
	return queryResponse.Balances, nil
}
