package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (c Client) AuthQClient() authtypes.QueryClient {
	return authtypes.NewQueryClient(c.Query.GrpcConn)
}

func (c Client) QueryAuthSeq(accAddr string) (uint64, error) {
	ctx, cancel := c.NewQCtx()
	defer cancel()

	queryResponse, err := c.AuthQClient().Account(ctx, &authtypes.QueryAccountRequest{
		Address: accAddr,
	})
	if err != nil {
		return 0, err
	}

	var baseAccount authtypes.AccountI
	err = c.codec.UnpackAny(queryResponse.Account, &baseAccount)
	if err != nil {
		return 0, err
	}
	accSeq := baseAccount.GetSequence()
	return accSeq, nil
}

func (c Client) QueryTxHash(hash string) (*sdk.TxResponse, error) {
	return authtx.QueryTx(*c.Tx.ClientContext, hash)
}
