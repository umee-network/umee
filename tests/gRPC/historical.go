package gRPC

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/umee-network/umee/v3/util/decmath"
)

const (
	rpcTimeout    = "100ms"
	gasAdjustment = 1
)

func MedianCheck(
	chainID string,
	tmrpcEndpoint string,
	grpcEndpoint string,
	val1Mnemonic string,
) error {

	val1Account, keyring, err := CreateAccountFromMnemonic("val1", val1Mnemonic)
	if err != nil {
		return err
	}

	client1, err := NewClient(
		context.Background(),
		zerolog.Nop(),
		chainID,
		keyring,
		keyringPassphrase,
		val1Account,
		tmrpcEndpoint,
		rpcTimeout,
		grpcEndpoint,
		gasAdjustment,
	)
	if err != nil {
		return err
	}

	price1 := client1.GetAtomPrice()
	price2 := client1.GetAtomPrice()
	price3 := client1.GetAtomPrice()

	median := client1.GetAtomMedianPrice()

	calcMedian, err := decmath.Median([]sdk.Dec{price1, price2, price3})
	if err != nil {
		return err
	}
	if median != calcMedian {
		return fmt.Errorf("Expected %d for the median but got %d")
	}

	return nil
}
