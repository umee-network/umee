package grpc

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/client"
)

func GetTxResponse(umeeClient client.Client, txHash string) (resp *sdk.TxResponse, err error) {
	for i := 0; i < 5; i++ {
		resp, err = umeeClient.QueryTxHash(txHash)
		if err == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return resp, err
}

func GetTxResponseAndCheckLogs(umeeClient client.Client, txHash string) (*sdk.TxResponse, error) {
	fullResp, err := GetTxResponse(umeeClient, txHash)
	if err != nil {
		return nil, err
	}

	if len(fullResp.Logs) == 0 {
		return nil, fmt.Errorf("no logs in response")
	}

	return fullResp, nil
}
