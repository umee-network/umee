package sdkclient

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (c Client) AuthQClient() authtypes.QueryClient {
	return authtypes.NewQueryClient(c.GrpcConn)
}

func (c Client) QueryAuthSeq(accAddr string) (uint64, error) {
	ctx, cancel := c.NewCtx()
	defer cancel()

	queryResponse, err := c.AuthQClient().Account(ctx, &authtypes.QueryAccountRequest{
		Address: accAddr,
	})
	if err != nil {
		return 0, err
	}

	var baseAccount authtypes.AccountI
	err = c.encCfg.Codec.UnpackAny(queryResponse.Account, &baseAccount)
	if err != nil {
		return 0, err
	}
	accSeq := baseAccount.GetSequence()
	return accSeq, nil
}

func (c Client) QueryTxHash(hash string) (*sdk.TxResponse, error) {
	return authtx.QueryTx(*c.ClientContext, hash)
}

// GetTxResponse waits for tx response and checks if the receipt contains at least `minExpectedLogs`.
func (c Client) GetTxResponse(txHash string, minExpectedLogs int) (resp *sdk.TxResponse, err error) {
	for i := 0; i < 6; i++ {
		resp, err = c.QueryTxHash(txHash)
		if err == nil {
			break
		}

		// TODO: configure sleep time
		// Ideally, we should subscribe to block websocket and query by tx hash
		time.Sleep(500 * time.Millisecond)
	}

	if n := len(resp.Logs); n < minExpectedLogs {
		return nil, fmt.Errorf("expecting at least %d logs in response, got: %d", minExpectedLogs, n)
	}

	return resp, err
}
