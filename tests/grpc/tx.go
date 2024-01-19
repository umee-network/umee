package grpc

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/client"
)

// GetTxResponse waits for tx response and checks if the receipt contains at least `minExpectedLogs`.
func GetTxResponse(umeeClient client.Client, txHash string, minExpectedLogs int) (resp *sdk.TxResponse, err error) {
	for i := 0; i < 5; i++ {
		resp, err = umeeClient.QueryTxHash(txHash)
		if err == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}

	if n := len(resp.Logs); n < minExpectedLogs {
		return nil, fmt.Errorf("expecting at least %d logs in response, got: %d", minExpectedLogs, n)
	}

	return resp, err
}
